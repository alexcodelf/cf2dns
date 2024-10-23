package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/alexcodelf/cf2dns/pkg/config"
	"github.com/alexcodelf/cf2dns/pkg/fetcher"
	"github.com/alexcodelf/cf2dns/pkg/logger"
	"github.com/alexcodelf/cf2dns/pkg/provider"

	"go.uber.org/zap"
)

const (
	cfFetcherName = "cloudflare"
	gcoreFetcherName = "gcore"
)

var (
	version    = "dev"
	configFile string
)

func init() {
	flag.StringVar(&configFile, "config", "config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	logger := logger.NewLogger()
	defer logger.Sync()

	logger.Info("启动 DNS 更新器", zap.String("版本", version))

	cfg, err := config.Load(configFile)
	if err != nil {
		logger.Fatal("加载配置失败", zap.Error(err))
	}

	if err := validateConfig(cfg); err != nil {
		logger.Fatal("配置无效", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cfProvider := provider.NewCloudflareProvider(cfg.CloudflareAPIToken, logger)

	fetchers := map[string]*fetcher.Fetcher{
		cfFetcherName:    fetcher.NewFetcher(cfg.CFURL, cfg.CFDomain, cfg.CFNames, logger),
		gcoreFetcherName: fetcher.NewFetcher(cfg.GcoreURL, cfg.GcoreDomain, cfg.GcoreNames, logger),
	}

	if err := updateDNS(ctx, cfProvider, fetchers); err != nil {
		logger.Error("更新 DNS 失败", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("DNS 更新完成")
}

func validateConfig(cfg *config.Config) error {
	if cfg.CloudflareAPIToken == "" {
		return fmt.Errorf("Cloudflare API 令牌未设置")
	}
	// 添加其他必要的配置检查
	return nil
}

func updateDNS(ctx context.Context, p provider.Provider, fetchers map[string]*fetcher.Fetcher) error {
	for _, fetcher := range fetchers {
		ipInfos, err := fetcher.GetSortedIPs(ctx)
		if err != nil {
			return err
		}

		ips := make([]string, len(ipInfos))
		for i, ipInfo := range ipInfos {
			ips[i] = ipInfo.IP
		}

		record := provider.UpdateRecord{
			Domain: fetcher.Domain,
			Names:  fetcher.Names,
			IPs:     ips,
		}

		if err := p.UpdateRecord(ctx, record); err != nil {
			return err
		}
	}

	return nil
}