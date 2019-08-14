package push

import "go.uber.org/zap"

func logger() *zap.Logger {
	return zap.L().Named("push")
}
