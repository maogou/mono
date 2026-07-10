package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	DefaultConfigPath = "config/go_template.yaml"
)

type Config struct {
	Name  string `yaml:"name" validate:"required"`
	Mode  string `yaml:"mode" validate:"required,oneof=debug test release"`
	Port  int    `yaml:"port" validate:"required"`
	Log   *Log   `yaml:"log" validate:"required"`
	DB    *DB    `yaml:"db" validate:"required"`
	Redis *Redis `yaml:"redis" validate:"required"`
}

type DB struct {
	Dsn             string `yaml:"dsn" validate:"required"`
	MaxIdleConns    int    `yaml:"max_idle_conns" validate:"required,min=1"`
	MaxOpenConns    int    `yaml:"max_open_conns" validate:"required,min=1"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" validate:"required,min=1"`
}

type Redis struct {
	Addr        string `yaml:"addr" validate:"required"`
	Password    string `yaml:"password"`
	DB          int    `yaml:"db"`
	LockTimeout int    `yaml:"lock_timeout" validate:"required,min=1"`
}

type Log struct {
	Filename   string `yaml:"filename"`
	Level      string `yaml:"level" validate:"required,oneof=debug info warn error"`
	MaxSize    int    `yaml:"max_size" validate:"required,min=100"`
	MaxAge     int    `yaml:"max_age" validate:"required,min=1"`
	MaxBackups int    `yaml:"max_backups" validate:"required,min=1"`
	Compress   bool   `yaml:"compress"`
	Encoding   string `yaml:"encoding" validate:"required,oneof=json console"`
	Mode       string `yaml:"mode" validate:"required,oneof=file console both"`
}

func MustLoadConfig(cfgFile string) Config {
	if len(cfgFile) == 0 {
		cfgFile = DefaultConfigPath
	}
	
	conf := viper.New()
	
	conf.SetConfigFile(cfgFile)
	conf.SetConfigType("yaml")
	
	if err := conf.ReadInConfig(); err != nil {
		panic(err)
	}
	
	var config Config
	if err := conf.Unmarshal(&config); err != nil {
		panic(err)
	}
	
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		panic(err)
	}
	
	return config
}
