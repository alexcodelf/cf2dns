package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	// 其他可能需要的库

	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
)

// 定义结构体和之前的逻辑相同...

func main() {
	var fastUrl string
	var domain string
	var names string
	var apiToken string

	app := &cli.App{
		Name:  "Update Cloudflare DNS",
		Usage: "Updates the DNS records for given domain and subdomains with lowest latency IPs using Cloudflare API.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "url",
				Value:       "https://monitor.gacjie.cn/api/client/get_ip_address?cdn_server=3",
				Usage:       "Domain for the DNS record update",
				Destination: &fastUrl,
			},
			&cli.StringFlag{
				Name:        "domain",
				Value:       "example.com",
				Usage:       "Domain for the DNS record update",
				Destination: &domain,
			},
			&cli.StringFlag{
				Name:        "names",
				Usage:       "The subdomain names to update, use this flag multiple times for multiple names",
				Value:       "cdn1,cdn2,cdn3,cdn4,cdn5",
				Destination: &names,
			},
			&cli.StringFlag{
				Name:        "cloudflare-api-token",
				Usage:       "Cloudflare API token",
				Destination: &apiToken,
				EnvVars:     []string{"CLOUDFLARE_API_TOKEN"}, // Alternatively read from environment variable
			},
		},
		Action: func(c *cli.Context) error {
			api, err := cloudflare.NewWithAPIToken(apiToken) // Use the apiToken
			if err != nil {
				return err
			}

			domainNames := strings.Split(names, ",")

			// 获取 IP 地址列表
			ips, err := fetchIPAddresses(fastUrl)
			if err != nil {
				return fmt.Errorf("Error fetching IP address:", err)
			}

			// 延迟 超过 150ms 的 IP 地址不予考虑
			var validIPs []IPInfo
			for _, ip := range ips {
				if ip.Delay < 150 {
					validIPs = append(validIPs, ip)
				}
			}

			// 排序 IP 地址列表基于延迟
			sort.Slice(validIPs, func(i, j int) bool {
				return validIPs[i].Delay < validIPs[j].Delay
			})

			// 选取延迟最低的前n个地址
			topIPs := validIPs
			if len(validIPs) > len(domainNames) {
				topIPs = validIPs[:len(domainNames)]
			}

			_, err = resolveIPtoDomain(c.Context, api, domain, domainNames, topIPs)
			if err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// resolveIPtoDomain 使用 cloudflare API 将 IP 解析到域名
func resolveIPtoDomain(ctx context.Context, api *cloudflare.API, domain string, names []string, ips []IPInfo) (bool, error) {
	// Fetch the zone ID
	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		return false, err
	}

	proxied := false

	// 使用 G-Core API 解析每个 IP 地址到指定域名
	for i, ipInfo := range ips {
		// G-Core API 逻辑，可根据实际 API 文档进行调整
		name := names[i]
		recordName := fmt.Sprintf("%s.%s", name, domain)

		dnsRecords, _, err := api.ListDNSRecords(ctx, cloudflare.AccountIdentifier(zoneID), cloudflare.ListDNSRecordsParams{
			Name: recordName,
		})
		if err != nil {
			return false, err
		}

		if len(dnsRecords) == 0 {
			fmt.Println("No DNS record found for", name)
			continue
		}

		// Update the DNS record
		_, err = api.UpdateDNSRecord(ctx, cloudflare.AccountIdentifier(zoneID), cloudflare.UpdateDNSRecordParams{
			ID:      dnsRecords[0].ID,
			Type:    "A",
			Name:    name,
			Proxied: &proxied,
			Content: ipInfo.IP,
		})
		if err != nil {
			fmt.Println("Error resolving IP to domain:", err)
			continue
		} else {
			fmt.Printf("IP %s resolved to %s.%s \n", ipInfo.IP, name, domain)
		}
	}

	return true, nil
}

// fetchIPAddresses 和之前的逻辑相同...

// IPInfo 结构体用于存储单个 IP 相关的信息
type IPInfo struct {
	IP        string `json:"ip"`
	Address   string `json:"address"`
	LineName  string `json:"line_name"`
	Bandwidth int    `json:"bandwidth"`
	Speed     int    `json:"speed"`
	Colo      string `json:"colo"`
	Delay     int    `json:"delay"`
}

// ApiResponse 结构体用于存储 API 响应的数据
type ApiResponse struct {
	Status bool                `json:"status"`
	Code   int                 `json:"code"`
	Msg    string              `json:"msg"`
	Info   map[string][]IPInfo `json:"info"`
}

// fetchIPAddresses 发送 HTTP 请求并解析响应数据
func fetchIPAddresses(url string) ([]IPInfo, error) {
	// 发起 HTTP GET 请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析 JSON 数据
	var apiResponse ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	// 合并所有运营商的 IP 信息
	var allIPs []IPInfo
	for _, ips := range apiResponse.Info {
		allIPs = append(allIPs, ips...)
	}

	// 返回 IP 列表
	return allIPs, nil
}
