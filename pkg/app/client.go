package app

import (
	"context"
	"crypto/tls"

	"github.com/gammazero/nexus/v3/client"
)

func NewTlsClient(ctx context.Context, addr string, cfg client.Config) (*client.Client, error) {
	cfg.WsCfg.EnableCompression = true
	if cfg.TlsCfg == nil {
		cfg.TlsCfg = &tls.Config{}
	}
	cfg.TlsCfg.InsecureSkipVerify = true
	c, err := client.ConnectNet(ctx, addr, cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}
