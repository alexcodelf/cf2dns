package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

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

func NewFetcher(url, domain string, names []string, log *zap.Logger) *Fetcher {
	return &Fetcher{
		URL:    url,
		Domain: domain,
		Names:  names,
		logger: log,
	}
}

func (f *Fetcher) FetchIPs(ctx context.Context) ([]IPInfo, error) {
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

	return allIPs, nil
}

// GetSortedIPs returns the IPs sorted by delay
func (f *Fetcher) GetSortedIPs(ctx context.Context) ([]IPInfo, error) {
	ips, err := f.FetchIPs(ctx)
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

	for _, ips := range groupedIPs {
		// get the first 2 IPs, if there are less than 2, get all of them
		if len(ips) < 2 {
			allIPs = append(allIPs, ips...)
		} else {
			allIPs = append(allIPs, ips[:2]...)
		}
	}

	return allIPs, nil
}