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

func DefaultLogStartWork(component string) {
	Log.Info().
		Str("component", component).
		Msg("Starting work")
}

func DefaultLogFinishWork(component string, err error) {
	event := Log.Info().
		Str("component", component)

	if err != nil {
		event.Err(err).Msg("Finished work with error")
	} else {
		event.Msg("Finished work")
	}
}
