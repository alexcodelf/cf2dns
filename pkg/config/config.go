package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 定义应用程序的配置结构
type Config struct {
	Cloudflare         PriorityConfig `mapstructure:"cloudflare"`
	Gcore              PriorityConfig `mapstructure:"gcore"`
	MaxDelay           int            `mapstructure:"maxDelay"`
	MinBandwidth       int            `mapstructure:"minBandwidth"`
	CloudflareAPIToken string         `mapstructure:"cloudflareApiToken"`
}

// PriorityConfig 定义优选 IP 解析的配置
type PriorityConfig struct {
	// 优选 IP 解析的 URL
	URL string `mapstructure:"url"`
	// 优选 IP 解析的域名s
	Domain string `mapstructure:"domain"`
	// 优选 IP 解析的子域名，如cf1,cf2,cf3
	Names []string `mapstructure:"names"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Cloudflare: PriorityConfig{
			URL: "https://www.wetest.vip/api/cf2dns/get_cloudflare_ip",
		},
		Gcore: PriorityConfig{
			URL: "https://www.wetest.vip/api/cf2dns/get_gcore_ip",
		},
		MaxDelay:     250,
		MinBandwidth: 15,
	}
}

// Load 从指定的文件加载配置
func Load(configFile string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configFile)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	cfg := DefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return cfg, nil
}
