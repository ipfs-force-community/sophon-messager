package metrics

import (
	"context"

	"github.com/ipfs-force-community/metrics"
	"go.uber.org/fx"
)

func SetupJaeger(lc fx.Lifecycle, tcfg *metrics.TraceConfig) error {
	log.Infof("tracing config:enabled: %v, serverName:%s, jaeger-url:%s, sample:%.2f\n",
		tcfg.JaegerTracingEnabled, tcfg.ServerName,
		tcfg.JaegerEndpoint, tcfg.ProbabilitySampler)

	if !tcfg.JaegerTracingEnabled {
		return nil
	}

	exporter, err := metrics.SetupJaegerTracing(tcfg.ServerName, tcfg)
	if err != nil {
		return err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return metrics.ShutdownJaeger(ctx, exporter)
		},
	})
	return nil
}
