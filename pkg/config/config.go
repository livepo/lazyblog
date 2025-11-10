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
	Domain string `mapstructure:"domain"`
}

type Auth struct {
	XAdminToken string `mapstructure:"XAdminToken"`
}

type ImageHostingConfig struct {
	Enable       bool   `mapstructure:"enable"`
	Provider     string `mapstructure:"provider"`
	ClientId     string `mapstructure:"clientId"`
	ClientSecret string `mapstructure:"clientSecret"`
	AlbumId      string `mapstructure:"albumId"`
}

type Config struct {
	Mysql         MysqlConfig          `mapstructure:"mysql"`
	Auth          Auth                 `mapstructure:"auth"`
	Site          SiteConfig           `mapstructure:"site"`
	ImageHostings []ImageHostingConfig `mapstructure:"imageHostings"`
}

func init() {
	cfg := Config{}

	viper.SetConfigFile("config/config.toml")
	viper.ReadInConfig()
	viper.Unmarshal(&cfg)
	Cfg = &cfg
	fmt.Println("Cfg:", Cfg)
}
