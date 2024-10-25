package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 定义应用程序的配置结构
type Config struct {
	Cloudflare         PriorityConfig `mapstructure:"cloudflare"`
	Gcore              PriorityConfig `mapstructure:"gcore"`
	MaxDelay           time.Duration  `mapstructure:"maxDelay"`
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
		MaxDelay: 500 * time.Millisecond,
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

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	return cfg, nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if err := c.Cloudflare.validate(); err != nil {
		return fmt.Errorf("cloudflare 配置无效: %w", err)
	}
	if err := c.Gcore.validate(); err != nil {
		return fmt.Errorf("gcore 配置无效: %w", err)
	}
	if c.MaxDelay <= 0 {
		return errors.New("最大延迟必须大于 0")
	}
	if c.CloudflareAPIToken == "" {
		return errors.New("缺少 Cloudflare API 密钥")
	}
	return nil
}

func (cc *PriorityConfig) validate() error {
	if cc.URL == "" {
		return errors.New("缺少优选 IP 解析 URL")
	}
	if cc.Domain == "" {
		return errors.New("缺少优选 IP 解析域名")
	}
	if len(cc.Names) == 0 {
		return errors.New("缺少子域名，如cf1,cf2,cf3")
	}
	return nil
}
