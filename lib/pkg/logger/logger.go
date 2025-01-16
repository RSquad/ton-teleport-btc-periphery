package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init() {
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	zerolog.DurationFieldUnit = time.Second

	Log = zerolog.New(consoleWriter).With().Timestamp().Logger()
}
