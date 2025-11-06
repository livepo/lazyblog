package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var (
	Cfg *Config
)

type MysqlConfig struct {
	Host     string
	Port     int32
	User     string
	Password string
	Database string
}

type SiteConfig struct {
	Prefix string `mapstructure:"prefix"`
	Title  string `mapstructure:"title"`
	About  string `mapstructure:"about"`
}

type Auth struct {
	XAdminToken string
}

type Config struct {
	Mysql MysqlConfig `mapstructure:"mysql"`
	Auth  Auth        `mapstructure:"auth"`
	Site  SiteConfig  `mapstructure:"site"`
}

func init() {
	cfg := Config{}

	viper.SetConfigFile("config/config.toml")
	viper.ReadInConfig()
	viper.Unmarshal(&cfg)
	Cfg = &cfg
	fmt.Println("Cfg:", Cfg)
}
