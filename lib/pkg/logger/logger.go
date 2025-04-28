package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Level int

const (
	DebugLevel Level = Level(zerolog.DebugLevel)
	InfoLevel        = Level(zerolog.InfoLevel)
	WarnLevel        = Level(zerolog.WarnLevel)
	ErrorLevel       = Level(zerolog.ErrorLevel)
)

var Log zerolog.Logger

func Init(logFile string, level Level, maxFileSize int, maxBackups int, maxBackupAge int) error {
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	zerolog.DurationFieldUnit = time.Second

	if len(logFile) > 0 {
		err := checkIfPathWritable(logFile)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}

		fileWriter := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    maxFileSize,  // megabytes per file
			MaxBackups: maxBackups,   // number of backups to keep
			MaxAge:     maxBackupAge, // days to keep backups
			Compress:   true,
		}

		multi := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
		Log = zerolog.New(multi).With().Timestamp().Logger()
	} else {
		Log = zerolog.New(consoleWriter).With().Timestamp().Logger()
		Log.Warn().Msg("No log file name was set. Logs are written only to stdout")
	}

	zerolog.SetGlobalLevel(zerolog.Level(level))

	return nil
}

func checkIfPathWritable(pathWithFile string) error {
	dirPath := filepath.Dir(pathWithFile)

	// Check if the path exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0o700)
		if err != nil {
			return fmt.Errorf("failed to create directory `%s`: %w", dirPath, err)
		}
		fmt.Printf("Created directory `%s`\n", dirPath)
	}

	// Check if the path is writable
	tempFile := filepath.Join(dirPath, ".write_test")
	f, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("directory `%s` is not writable: %w", dirPath, err)
	}

	// Clean up the test file
	f.Close()
	os.Remove(tempFile)

	return nil
}

func DefaultLogStartWork(component string) {
	Log.Info().
		Str("component", component).
		Msg("Starting work")
}

func DefaultLogFinishWork(component string) {
	Log.Info().
		Str("component", component).
		Msg("Finished work")
}

func ParseLevel(levelStr string, defaultLevel Level) (Level, error) {
	if len(levelStr) == 0 {
		return defaultLevel, nil
	}

	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return DebugLevel, nil
	case "INFO":
		return InfoLevel, nil
	case "WARN":
		return WarnLevel, nil
	case "ERROR":
		return ErrorLevel, nil
	default:
		return InfoLevel, errors.New("invalid level: " + levelStr)
	}
}
