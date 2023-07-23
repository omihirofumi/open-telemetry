package main

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	samplingConfig := &zap.SamplingConfig{
		Initial:    3,
		Thereafter: 3,
		Hook: func(entry zapcore.Entry, decision zapcore.SamplingDecision) {
			if decision == zapcore.LogDropped {
				fmt.Println("event dropped...")
			}
		},
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Sampling = samplingConfig

	logger, _ := cfg.Build()

	zap.ReplaceGlobals(logger)
}

func main() {
	for i := 1; i <= 9; i++ {
		zap.S().Infow(
			"Test sampling",
			"index", i)
	}
}
