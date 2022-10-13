package main

import (
	"lovebox/cmd"

	"go.uber.org/zap"
)

func main() {
	if err := cmd.Execute(); err != nil {
		zap.S().Error(err)
	}
}
