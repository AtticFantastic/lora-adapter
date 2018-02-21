package lora

import "go.uber.org/zap"

var logger *zap.Logger

// InitLogger - Init zap Logger
func InitLogger(l *zap.Logger) {
	logger = l
	return
