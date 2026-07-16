package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	DefaultConfigPath = "config/go_template.yaml"
)

type Config struct {
	Name            string `yaml:"name" mapstructure:"Name" validate:"required"`
	Mode            string `yaml:"mode" mapstructure:"Mode" validate:"required,oneof=debug test release"`
	ShutdownTimeout int    `yaml:"shutdown_timeout" mapstructure:"ShutdownTimeout" validate:"required,min=5"`
	ReadTimeout     int    `yaml:"read_timeout" mapstructure:"ReadTimeout" validate:"required,min=5"`
	Port            int    `yaml:"port" mapstructure:"port" validate:"required"`
	Log             *Log   `yaml:"log" mapstructure:"Log" validate:"required"`
	DB              *DB    `yaml:"db" mapstructure:"DB" validate:"required"`
	Redis           *Redis `yaml:"redis" mapstructure:"Redis" validate:"required"`
}

type DB struct {
	Dsn             string `yaml:"dsn" mapstructure:"Dsn" validate:"required"`
	MaxIdleConns    int    `yaml:"max_idle_conns" mapstructure:"MaxIdleConns" validate:"required,min=1"`
	MaxOpenConns    int    `yaml:"max_open_conns" mapstructure:"MaxOpenConns" validate:"required,min=1"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" mapstructure:"ConnMaxLifetime" validate:"required,min=1"`
}

type Redis struct {
	Addr        string `yaml:"addr" mapstructure:"Addr" validate:"required"`
	Password    string `yaml:"password" mapstructure:"Password"`
	DB          int    `yaml:"db" mapstructure:"DB"`
	LockTimeout int    `yaml:"lock_timeout" mapstructure:"LockTimeout" validate:"required,min=1"`
}

type Log struct {
	Filename   string `yaml:"filename" mapstructure:"Filename"`
	Level      string `yaml:"level" mapstructure:"Level" validate:"required,oneof=debug info warn error"`
	MaxSize    int    `yaml:"max_size" mapstructure:"MaxSize" validate:"required,min=100"`
	MaxAge     int    `yaml:"max_age" mapstructure:"MaxAge" validate:"required,min=1"`
	MaxBackups int    `yaml:"max_backups" mapstructure:"MaxBackups" validate:"required,min=1"`
	Compress   bool   `yaml:"compress" mapstructure:"Compress"`
	Encoding   string `yaml:"encoding" mapstructure:"Encoding" validate:"required,oneof=json console"`
	Mode       string `yaml:"mode" mapstructure:"Mode" validate:"required,oneof=file console both"`
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
