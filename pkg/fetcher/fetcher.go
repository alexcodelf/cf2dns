package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/alexcodelf/cf2dns/pkg/config"
	"go.uber.org/zap"
)

type IPInfo struct {
	IP        string `json:"ip"`
	Address   string `json:"address"`
	LineName  string `json:"line_name"`
	Bandwidth int    `json:"bandwidth"`
	Speed     int    `json:"speed"`
	Colo      string `json:"colo"`
	Delay     int    `json:"delay"`
}

type APIResponse struct {
	Status bool                `json:"status"`
	Code   int                 `json:"code"`
	Msg    string              `json:"msg"`
	Info   map[string][]IPInfo `json:"info"`
}

type Fetcher struct {
	URL    string
	Domain string
	Names  []string
	logger *zap.Logger
}

func NewFetcher(pc config.PriorityConfig, log *zap.Logger) *Fetcher {
	return &Fetcher{
		URL:    pc.URL,
		Domain: pc.Domain,
		Names:  pc.Names,
		logger: log,
	}
}

func (f *Fetcher) FetchIPs(ctx context.Context, maxDelay int, minBandwidth int) ([]IPInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Status {
		return nil, fmt.Errorf("API returned error: %s", apiResp.Msg)
	}

	var allIPs []IPInfo
	for _, ips := range apiResp.Info {
		allIPs = append(allIPs, ips...)
	}

	filteredIPs := []IPInfo{}
	for _, ip := range allIPs {
		if ip.Delay > maxDelay {
			continue
		}

		if ip.Bandwidth < minBandwidth {
			continue
		}

		filteredIPs = append(filteredIPs, ip)
	}

	return filteredIPs, nil
}

// GetSortedIPs returns the IPs sorted by delay
func (f *Fetcher) GetSortedIPs(ctx context.Context, maxDelay int, minBandwidth int) ([]IPInfo, error) {
	ips, err := f.FetchIPs(ctx, maxDelay, minBandwidth)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IPs: %w", err)
	}

	// group by line name
	groupedIPs := make(map[string][]IPInfo)

	for _, ip := range ips {
		if _, ok := groupedIPs[ip.LineName]; !ok {
			groupedIPs[ip.LineName] = []IPInfo{}
		}

		groupedIPs[ip.LineName] = append(groupedIPs[ip.LineName], ip)
	}

	// sort grouped IPs by delay
	for _, ips := range groupedIPs {
		sort.Slice(ips, func(i, j int) bool {
			return ips[i].Delay < ips[j].Delay
		})
	}

	var allIPs []IPInfo

	// 获取每个线路的第一个 IP, 取出的 IP 从 groupedIPs 中删除，直到 groupedIPs 为空
	for len(groupedIPs) > 0 {
		for group, ips := range groupedIPs {
			allIPs = append(allIPs, ips[0])
			if len(ips) > 1 {
				groupedIPs[group] = ips[1:]
			} else {
				delete(groupedIPs, group)
			}
		}
	}

	return allIPs, nil
}
