package provider

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"go.uber.org/zap"
)

type CloudflareProvider struct {
	api *cloudflare.API
	log *zap.Logger
}

func NewCloudflareProvider(apiToken string, log *zap.Logger) *CloudflareProvider {
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		log.Fatal("Failed to create Cloudflare API client", zap.Error(err))
	}
	return &CloudflareProvider{api: api, log: log}
}

func (p *CloudflareProvider) UpdateRecord(ctx context.Context, record UpdateRecord) error {
	zoneID, err := p.api.ZoneIDByName(record.Domain)
	if err != nil {
		return err
	}

	for i, name := range record.Names {
		if i >= len(record.IPs) {
			break
		}

		ip := record.IPs[i]

		name = strings.TrimSuffix(name, "."+record.Domain)

		records, _, err := p.api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{Name: name})
		if err != nil {
			return err
		}

		if len(records) == 0 {
			_, err = p.api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
				Type:    "A",
				Name:    name,
				Content: ip,
				TTL:     1,
			})
			if err != nil {
				return err
			}
			continue // Skip the update, since the record was just created
		}

		_, err = p.api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.UpdateDNSRecordParams{
			ID:      records[0].ID,
			Type:    "A",
			Name:    name,
			Content: ip,
			TTL:     1,
			Proxied: cloudflare.BoolPtr(false),
		})
		if err != nil {
			return err
		}
	}

	return nil
}