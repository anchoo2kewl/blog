package main

import (
	"log"

	"go.uber.org/zap"
)

func sugarLog() *zap.SugaredLogger {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()
	defer logger.Sync()

	return sugar
}
