//go:build ignore

// REFERENCE ONLY — real OpenTelemetry tracing setup with a stdout exporter.
// Build-tagged `ignore` so this folder builds without the OTel deps.
// To run for real:
//   remove the build tag, then:
//   go get go.opentelemetry.io/otel \
//          go.opentelemetry.io/otel/sdk/trace \
//          go.opentelemetry.io/otel/exporters/stdout/stdouttrace
//   go run .
package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// initTracer wires a TracerProvider with a stdout exporter (use OTLP -> Jaeger
// in production). Call the returned shutdown func on exit.
func initTracer() (func(context.Context) error, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		// sample 100% in dev; use sdktrace.TraceIDRatioBased(0.05) in prod
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

var tracer = otel.Tracer("orders-service")

func placeOrderOTel(ctx context.Context, userID string) error {
	ctx, span := tracer.Start(ctx, "PlaceOrder")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", userID))

	if err := chargeOTel(ctx); err != nil { // ctx carries the parent span
		span.RecordError(err)
		span.SetStatus(codes.Error, "charge failed")
		return err
	}
	return nil
}

func chargeOTel(ctx context.Context) error {
	_, span := tracer.Start(ctx, "Charge") // automatically a child of PlaceOrder
	defer span.End()
	time.Sleep(15 * time.Millisecond)
	return nil
}

func demo() {
	shutdown, err := initTracer()
	if err != nil {
		panic(err)
	}
	defer shutdown(context.Background())
	if err := placeOrderOTel(context.Background(), "u_1"); err != nil {
		fmt.Println("error:", err)
	}
}
