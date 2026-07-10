//nolint:gosec
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "用法: go run scripts/replace.go <旧项目名> <新项目名>\n")
		fmt.Fprintf(os.Stderr, "示例: go run scripts/replace.go go_template my_project\n")
		os.Exit(1)
	}
	oldName := os.Args[1]
	newName := os.Args[2]

	re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(-[a-zA-Z0-9_]+)*$`)
	if !re.MatchString(oldName) {
		fmt.Fprintf(os.Stderr, "错误: 旧项目名 '%s' 不合法\n", oldName)
		os.Exit(1)
	}
	if !re.MatchString(newName) {
		fmt.Fprintf(os.Stderr, "错误: 新项目名 '%s' 不合法\n", newName)
		os.Exit(1)
	}
	if oldName == newName {
		fmt.Fprintf(os.Stderr, "错误: 新旧项目名不能相同\n")
		os.Exit(1)
	}

	projectDir := findProjectRoot()
	if projectDir == "" {
		fatalf("错误: 未找到 go.mod，请在项目根目录运行")
	}
	if err := os.Chdir(projectDir); err != nil {
		fatalf("无法进入项目目录 %s: %v", projectDir, err)
	}

	ensureCleanWorkTree()

	fmt.Printf("将项目从 '%s' 重命名为 '%s'\n\n", oldName, newName)
	fmt.Printf("项目目录: %s\n\n", projectDir)

	steps := []struct {
		name string
		fn   func() error
	}{
		{"替换 Go 源文件 import 路径", func() error { return replaceGoImports(oldName, newName) }},
		{"替换 go.mod module 名称", func() error {
			return replaceInFile(
				"go.mod", "module "+oldName, "module "+newName,
			)
		}},
		{"替换 Makefile", func() error { return replaceInFile("Makefile", oldName, newName) }},
		{"替换 Dockerfile", func() error { return replaceInFile("Dockerfile", oldName, newName) }},
		{"替换配置路径引用", func() error { return replaceConfigRefs(oldName, newName) }},
		{"重命名配置文件", func() error { return renameConfigFile(oldName, newName) }},
		{"重命名 cmd 入口目录", func() error { return renameCmdDir(oldName, newName) }},
	}

	for i, step := range steps {
		fmt.Printf("[%d/%d] %s...\n", i+1, len(steps)+1, step.name)
		if err := step.fn(); err != nil {
			gitRollback()
			fatalf("%s 失败: %v", step.name, err)
		}
	}

	lastStep := len(steps) + 1
	fmt.Printf("[%d/%d] 执行 go mod tidy...\n", lastStep, lastStep)
	if err := runCmd("go", "mod", "tidy"); err != nil {
		gitRollback()
		fatalf("go mod tidy 失败: %v", err)
	}

	fmt.Printf("\n编译验证...\n")
	if err := runCmd("go", "build", "./..."); err != nil {
		gitRollback()
		fatalf("编译验证 失败: %v", err)
	}

	fmt.Printf("\n重命名成功🎉🎉🎉!\n")
	fmt.Printf("  旧项目名: %s\n", oldName)
	fmt.Printf("  新项目名: %s\n", newName)
	fmt.Printf("  配置文件: config%s%s.yaml\n", string(filepath.Separator), newName)
	fmt.Printf("  入口目录: cmd%s%s\n", string(filepath.Separator), newName)
}

func findProjectRoot() string {
	dir, _ := filepath.Abs(".")
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func ensureCleanWorkTree() {
	// 检查未暂存的修改
	if out, err := gitCmd("diff", "--stat"); err != nil || len(out) > 0 {
		fatalf("错误: git 工作区有未暂存的修改，请先执行 git stash 或 git commit")
	}
	// 检查暂存区
	if out, err := gitCmd("diff", "--stat", "--cached"); err != nil || len(out) > 0 {
		fatalf("错误: git 暂存区有待提交的内容，请先执行 git commit")
	}
	// 检查未跟踪文件
	if out, err := gitCmd("ls-files", "--others", "--exclude-standard"); err != nil || len(out) > 0 {
		fatalf("错误: git 工作区有未跟踪的文件，请先清理或提交")
	}
}

func gitRollback() {
	fmt.Fprintf(os.Stderr, "\n通过 git 回滚所有修改...\n")

	// 恢复所有被修改/删除的已跟踪文件
	if _, err := gitCmd("checkout", "--", "."); err != nil {
		fmt.Fprintf(os.Stderr, "  警告: git checkout 失败: %v\n", err)
		return
	}

	// 删除脚本产生的未跟踪文件（重命名后的新目录/文件）
	if _, err := gitCmd("clean", "-fd"); err != nil {
		fmt.Fprintf(os.Stderr, "  警告: git clean 失败: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "回滚完成，项目已恢复原状。\n")
}

func replaceGoImports(oldName, newName string) error {
	oldPattern := `"` + oldName + `/`
	newPattern := `"` + newName + `/`

	var goFiles []string
	err := filepath.WalkDir(
		".", func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				switch filepath.Base(path) {
				case ".git", "vendor", "scripts":
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".go") {
				goFiles = append(goFiles, path)
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("遍历目录失败: %w", err)
	}

	for _, f := range goFiles {
		if err := replaceInFile(f, oldPattern, newPattern); err != nil {
			return err
		}
	}
	return nil
}

func replaceConfigRefs(oldName, newName string) error {
	if err := replaceInFile(
		filepath.Join("internal", "config", "config.go"),
		oldName+".yaml", newName+".yaml",
	); err != nil {
		return err
	}
	return replaceInFile(
		filepath.Join("internal", "pkg", "zlog", "zlog.go"),
		`"`+oldName+`.log"`, `"`+newName+`.log"`,
	)
}

func renameConfigFile(oldName, newName string) error {
	oldPath := filepath.Join("config", oldName+".yaml")
	newPath := filepath.Join("config", newName+".yaml")
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		fmt.Printf("  跳过: %s 不存在\n", oldPath)
		return nil
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("重命名 %s: %w", oldPath, err)
	}
	fmt.Printf("  已重命名: %s -> %s\n", oldPath, newPath)
	return nil
}

func renameCmdDir(oldName, newName string) error {
	oldDir := filepath.Join("cmd", oldName)

	var oldPath, oldPathRef string
	switch {
	case dirExists(oldDir):
		oldPath = oldDir
		oldPathRef = "cmd/" + oldName
	default:
		fmt.Printf("  跳过: cmd 下未找到 '%s' 目录\n", oldName)
		return nil
	}

	newDir := filepath.Join("cmd", newName)
	if err := os.Rename(oldPath, newDir); err != nil {
		return fmt.Errorf("重命名 %s: %w", oldPath, err)
	}
	fmt.Printf("  已重命名: %s -> %s\n", oldPath, newDir)

	// 修正 Makefile / Dockerfile 中的旧 cmd 路径引用
	for _, f := range []string{"Makefile", "Dockerfile"} {
		if err := replaceInFile(f, oldPathRef, "cmd/"+newName); err != nil {
			return err
		}
	}

	return nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func replaceInFile(filePath, old, new string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("  跳过: %s 不存在\n", filePath)
			return nil
		}
		return err
	}

	content := string(data)
	if !strings.Contains(content, old) {
		return nil
	}

	content = strings.ReplaceAll(content, old, new)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("  已更新: %s\n", filePath)
	return nil
}

func gitCmd(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
