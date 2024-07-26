package tqm

import (
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger() *logrus.Logger {
	// Set up Logrus
	log := logrus.New()

	// Set up Lumberjack for log file rotation
	logFile := &lumberjack.Logger{
		Filename:   "log/app.log",
		MaxSize:    5,    // Max size in MB
		MaxBackups: 3,    // Max number of old log files to keep
		MaxAge:     28,   // Max age in days to keep the log files
		Compress:   true, // Compress/zip old log files
	}

	// Set up multi-writer to write to both file and console
	mw := io.MultiWriter(os.Stdout, logFile)

	// Set the output of logrus to multi-writer
	log.SetOutput(mw)

	// Optionally, set the log level and formatter
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
		ForceColors:     true,
		DisableColors:   false,
	})
	return log
}

var Logrus = NewLogger()
