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
	cfFetcherName    = "cloudflare"
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
		gcoreFetcherName: fetcher.NewFetcher(cfg.Gcore, logger),
		cfFetcherName:    fetcher.NewFetcher(cfg.Cloudflare, logger),
	}

	if err := updateDNS(ctx, cfProvider, fetchers); err != nil {
		logger.Error("更新 DNS 失败", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("DNS 更新完成")
}

func validateConfig(cfg *config.Config) error {
	if cfg.CloudflareAPIToken == "" {
		return fmt.Errorf("cloudflare API 令牌未设置")
	}

	return nil
}

func updateDNS(ctx context.Context, p provider.Provider, fetchers map[string]*fetcher.Fetcher) error {
	logger := logger.NewLogger()
	defer logger.Sync()

	for t, fetcher := range fetchers {
		if fetcher.URL == "" {
			logger.Warn("优选 IP 解析 URL 为空，跳过更新", zap.String("fetcher", t))
			continue
		}

		if fetcher.Domain == "" {
			logger.Warn("优选 IP 解析域名为空，跳过更新", zap.String("fetcher", t))
			continue
		}

		if len(fetcher.Names) == 0 {
			logger.Warn("优选 IP 解析子域名为空，跳过更新", zap.String("fetcher", t))
			continue
		}

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
			IPs:    ips,
		}

		if err := p.UpdateRecord(ctx, record); err != nil {
			return err
		}
	}

	return nil
}
