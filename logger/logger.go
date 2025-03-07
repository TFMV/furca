package logger

import (
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.SugaredLogger
	once   sync.Once
)

// GetLogger returns a singleton instance of the logger
func GetLogger() *zap.SugaredLogger {
	once.Do(func() {
		// Create a custom encoder config
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		// Create a console encoder
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

		// Get log level from environment or config
		logLevel := getLogLevel()

		// Create a core that writes to stdout
		core := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			logLevel,
		)

		// Create a logger
		log := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		logger = log.Sugar()
	})

	return logger
}

// getLogLevel returns the appropriate log level based on the LOG_LEVEL environment variable
func getLogLevel() zapcore.Level {
	// Default to info level
	level := zapcore.InfoLevel

	// Get log level from environment or config
	logLevelStr := viper.GetString("LOG_LEVEL")
	if logLevelStr == "" {
		return level
	}

	// Convert to lowercase for case-insensitive comparison
	switch strings.ToLower(logLevelStr) {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn", "warning":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "dpanic":
		level = zapcore.DPanicLevel
	case "panic":
		level = zapcore.PanicLevel
	case "fatal":
		level = zapcore.FatalLevel
	}

	return level
}
