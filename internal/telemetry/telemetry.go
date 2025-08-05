package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/zap"
)

// InitTelemetry initializes OpenTelemetry with Prometheus exporter
func InitTelemetry(logger *zap.Logger) error {
	// Create Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return err
	}

	// Create meter provider with Prometheus exporter
	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
	)

	// Set global meter provider
	otel.SetMeterProvider(provider)

	logger.Info("OpenTelemetry initialized with Prometheus exporter")
	return nil
}

// Shutdown gracefully shuts down telemetry
func Shutdown(ctx context.Context) error {
	// Get the meter provider and shutdown
	if provider := otel.GetMeterProvider(); provider != nil {
		if mp, ok := provider.(*metric.MeterProvider); ok {
			return mp.Shutdown(ctx)
		}
	}
	return nil
}
