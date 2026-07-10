# GoTemplate基础模板代码

### 命名规范

统一的命名规范有助于快速定位日志中的问题文件：

- `handler` 层：文件名统一为 `xxx_handler.go`
- `model` 层：文件名统一为 `xxx_model.go`
- `repository` 层：文件名统一为 `xxx_repository.go`
- `service` 层：文件名统一为 `xxx_service.go`

### 数据库事务

数据库事务使用示例可参考 `demo_service.go` 文件。

### 目录结构

| 目录 | 职责说明 |
|------|----------|
| `api` | 定义请求结构体和响应结构体 |
| `handler` | 编写控制器逻辑 |
| `model` | 定义数据库表模型及作用域查询 |
| `repository` | 封装模型操作，包括第三方接口调用（如 park 接口）可封装为独立模块 |
| `service` | 实现核心业务逻辑，建议每个方法独立一个文件，避免代码臃肿 |

### 推荐工具库

| 用途 | 推荐库 |
|------|--------|
| 工具包函数 | [samber/lo](https://github.com/samber/lo) |
| 复杂 JSON 处理 | [tidwall/gjson](https://github.com/tidwall/gjson) |
| 唯一 ID 生成 | [rs/xid](https://github.com/rs/xid) |
| 类型转换 | [spf13/cast](https://github.com/spf13/cast) |

### 使用说明

- 克隆本仓库到本地
- go run scripts/replace.go go_template 你的项目名称

### 注意事项
- 本项目没有使用接口层,因为小项目根本没有必要
- 本模板仅作为参考，适合则用，不适可弃，请尊重开源精神
- 使用前请充分评估是否符合自身项目需求
- 任何线上问题请先自查，勿轻易归咎于框架模板代码
- 如果当时自己认可别人的东西并使用,那么请你尊重别人的工作,不要出了Bug就对模板代码作者进行C语言输出.


### 作者
- Email:
  - kinyou_xy@foxmail.com
- Wechat:
  - xingmaogou
