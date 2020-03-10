package loglevel

import (
	"strings"

	log "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// NewLevelFilterFromString filter the log level using the string "DEBUG|INFO|WARN|ERROR"
func NewLevelFilterFromString(next log.Logger, ls string) log.Logger {
	switch strings.ToLower(ls) {
	case "debug":
		return level.NewFilter(next, level.AllowDebug())
	case "info":
		return level.NewFilter(next, level.AllowInfo())
	case "warn", "warning":
		return level.NewFilter(next, level.AllowWarn())
	case "error", "err":
		return level.NewFilter(next, level.AllowError())
	}

	return level.NewFilter(next, level.AllowAll())
}
