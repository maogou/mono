package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	DefaultConfigPath = "config/go_template.yaml"
)

type Config struct {
	Name            string `yaml:"name" mapstructure:"name" validate:"required"`
	Mode            string `yaml:"mode" mapstructure:"mode" validate:"required,oneof=debug test release"`
	ShutdownTimeout int    `yaml:"shutdown_timeout" mapstructure:"shutdown_timeout" validate:"required,min=5"`
	ReadTimeout     int    `yaml:"read_timeout" mapstructure:"read_timeout" validate:"required,min=5"`
	Port            int    `yaml:"port" mapstructure:"port" validate:"required"`
	Log             *Log   `yaml:"log" mapstructure:"log" validate:"required"`
	DB              *DB    `yaml:"db" mapstructure:"db" validate:"required"`
	Redis           *Redis `yaml:"redis" mapstructure:"redis" validate:"required"`
}

type DB struct {
	Dsn             string `yaml:"dsn" mapstructure:"dsn" validate:"required"`
	MaxIdleConns    int    `yaml:"max_idle_conns" mapstructure:"max_idle_conns" validate:"required,min=1"`
	MaxOpenConns    int    `yaml:"max_open_conns" mapstructure:"max_open_conns" validate:"required,min=1"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime" validate:"required,min=1"`
}

type Redis struct {
	Addr        string `yaml:"addr" mapstructure:"addr" validate:"required"`
	Password    string `yaml:"password" mapstructure:"password"`
	DB          int    `yaml:"db" mapstructure:"db"`
	LockTimeout int    `yaml:"lock_timeout" mapstructure:"lock_timeout" validate:"required,min=1"`
}

type Log struct {
	Filename   string `yaml:"filename" mapstructure:"filename"`
	Level      string `yaml:"level" mapstructure:"level" validate:"required,oneof=debug info warn error"`
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size" validate:"required,min=100"`
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age" validate:"required,min=1"`
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups" validate:"required,min=1"`
	Compress   bool   `yaml:"compress" mapstructure:"compress"`
	Encoding   string `yaml:"encoding" mapstructure:"encoding" validate:"required,oneof=json console"`
	Mode       string `yaml:"mode" mapstructure:"mode" validate:"required,oneof=file console both"`
}

func LoadConfig(cfgFile string) (Config, error) {
	if len(cfgFile) == 0 {
		cfgFile = DefaultConfigPath
	}

	conf := viper.New()

	conf.SetConfigFile(cfgFile)
	conf.SetConfigType("yaml")

	if err := conf.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("读取配置文件 %s 失败: %w", cfgFile, err)
	}

	var config Config
	if err := conf.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("解析配置文件失败: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return Config{}, fmt.Errorf("配置校验失败: %w", err)
	}

	return config, nil
}
