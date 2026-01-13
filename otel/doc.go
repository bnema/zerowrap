// Package otel provides OpenTelemetry log bridging for zerowrap.
//
// This is an optional sub-package that adds OpenTelemetry dependencies.
// Only import it if you need to forward zerolog events to OpenTelemetry.
//
// # Usage
//
//	import (
//	    "github.com/bnema/zerowrap"
//	    "github.com/bnema/zerowrap/otel"
//	)
//
//	// Create logger with OTel hook
//	log := zerowrap.New(zerowrap.Config{
//	    Level:  "info",
//	    Format: "console",
//	}).Hook(otel.NewHook("my-service"))
//
//	// Attach to context
//	ctx := zerowrap.WithCtx(context.Background(), log)
//
//	// Logs now flow to both zerolog output AND OpenTelemetry
//	zerowrap.FromCtx(ctx).Info().Msg("hello world")
//
// # Custom Provider
//
// To use a specific logger provider instead of the global one:
//
//	provider := // your OTel logger provider
//	hook := otel.NewHookWithProvider(provider, "my-service")
//	log := zerowrap.New(cfg).Hook(hook)
package otel
