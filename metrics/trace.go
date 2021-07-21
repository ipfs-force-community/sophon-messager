package metrics

import (
	"context"
	"github.com/ipfs-force-community/metrics"
	"go.uber.org/fx"

	"github.com/filecoin-project/venus-messager/log"
)

func SetupJaeger(lc fx.Lifecycle, tcfg *metrics.TraceConfig, log *log.Logger) error {
	log.Infof("tracing config:enabled: %v, serverName:%s, jaeger-url:%s, sample:%.2f\n",
		tcfg.JaegerTracingEnabled, tcfg.ServerName,
		tcfg.JaegerEndpoint, tcfg.ProbabilitySampler)

	if !tcfg.JaegerTracingEnabled {
		return nil
	}

	exporter, err := metrics.RegisterJaeger(tcfg.ServerName, tcfg)
	if err != nil {
		return err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			metrics.UnregisterJaeger(exporter)
			return nil
		},
	})
	return nil
}
