package metrics

import (
	"context"
	"fmt"

	"github.com/ipfs-force-community/metrics"
	logging "github.com/ipfs/go-log/v2"
	"go.opencensus.io/stats/view"
	"go.uber.org/fx"
)

var log = logging.Logger("metric")

func SetupMetrics(lc fx.Lifecycle, metricsConfig *metrics.MetricsConfig) error {
	if err := view.Register(
		MessagerNodeViews...,
	); err != nil {
		return fmt.Errorf("cannot register the view: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if metricsConfig != nil {
				return metrics.SetupMetrics(ctx, metricsConfig)
			}
			return nil
		},
	})
	return nil
}
