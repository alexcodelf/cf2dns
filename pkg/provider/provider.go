package provider

import (
	"context"
)

type Provider interface {
	UpdateRecord(ctx context.Context, record UpdateRecord) error
}

type UpdateRecord struct {
	Domain string
	Names  []string
	IPs    []string
}