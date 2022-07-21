package metrics

import (
	"context"
	"fmt"

	"go.opencensus.io/stats/view"
	"go.uber.org/fx"

	"github.com/ipfs-force-community/metrics"

	"github.com/filecoin-project/venus-messager/log"
)

func SetupMetrics(lc fx.Lifecycle, metricsConfig *metrics.MetricsConfig, log *log.Logger) error {
	log.Infof("metrics config: enabled: %v, exporter type: %s, prometheus: %v, graphite: %v",
		metricsConfig.Enabled, metricsConfig.Exporter.Type, metricsConfig.Exporter.Prometheus,
		metricsConfig.Exporter.Graphite)

	if !metricsConfig.Enabled {
		return nil
	}

	if err := view.Register(
		MessagerNodeViews...,
	); err != nil {
		return fmt.Errorf("cannot register the view: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			switch metricsConfig.Exporter.Type {
			case metrics.ETPrometheus:
				go func() {
					if err := metrics.RegisterPrometheusExporter(ctx, metricsConfig.Exporter.Prometheus); err != nil {
						log.Errorf("Register prometheus exporter err: %v", err)
					}
					log.Info("Prometheus exporter server graceful shutdown successful")
				}()

			case metrics.ETGraphite:
				if err := metrics.RegisterGraphiteExporter(ctx, metricsConfig.Exporter.Graphite); err != nil {
					log.Errorf("failed to register the exporter: %v", err)
				}
			default:
				log.Warnf("invalid exporter type: %s", metricsConfig.Exporter.Type)
			}

			return nil
		},
	})

	return nil
}
