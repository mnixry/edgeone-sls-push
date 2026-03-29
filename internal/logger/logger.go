package logger

import (
	"io"
	"os"
	"time"

	"github.com/mnixry/edgeone-sls-push/internal/config"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(cfg config.LogConfig) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(cfg.Level)

	var writer io.Writer
	switch cfg.Output {
	case "stdout", "":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		if cfg.MaxSize > 0 {
			writer = &lumberjack.Logger{
				Filename:   cfg.Output,
				MaxSize:    cfg.MaxSize,
				MaxAge:     cfg.MaxAge,
				MaxBackups: cfg.MaxBackups,
				Compress:   cfg.Compress,
			}
		} else {
			f, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				writer = os.Stdout
			} else {
				writer = f
			}
		}
	}

	if cfg.Format == config.LogFormatConsole {
		writer = zerolog.ConsoleWriter{Out: writer, TimeFormat: time.RFC3339}
	}

	return zerolog.New(writer).With().Timestamp().Caller().Logger()
}
