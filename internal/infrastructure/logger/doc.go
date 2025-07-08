// Package logger provides structured logging functionality for cli-recover.
//
// The logger package implements a flexible logging system with the following features:
//
// - Multiple log levels (Debug, Info, Warn, Error, Fatal)
// - Structured logging with fields
// - Multiple output targets (console, file, or both)
// - Log rotation for file output
// - JSON or text formatting
// - Context-aware logging
// - Global logger instance
//
// Basic Usage:
//
//	import "github.com/cagojeiger/cli-recover/internal/infrastructure/logger"
//
//	// Use global logger
//	logger.Info("Application started")
//	logger.Error("Failed to connect", logger.F("host", "localhost"), logger.F("port", 5432))
//
//	// With fields
//	log := logger.WithField("user", "john")
//	log.Info("User logged in")
//
// Configuration:
//
//	cfg := logger.Config{
//	    Level:      "info",
//	    Output:     "both",
//	    FilePath:   "/var/log/cli-recover.log",
//	    MaxSize:    100, // MB
//	    MaxAge:     7,   // days
//	    JSONFormat: true,
//	    UseColor:   false,
//	}
//	err := logger.InitializeFromConfig(cfg)
//
// Environment Variables:
//
//	CLI_RECOVER_LOG_LEVEL=debug      # Log level
//	CLI_RECOVER_LOG_OUTPUT=file      # Output type
//	CLI_RECOVER_LOG_FILE=/tmp/app.log # Log file path
//	CLI_RECOVER_LOG_FORMAT=json      # Use JSON format
//	CLI_RECOVER_LOG_COLOR=false      # Disable color output
//
// Custom Logger:
//
//	// Create custom logger
//	log, err := logger.NewLogger(logger.Config{
//	    Level:  "debug",
//	    Output: "console",
//	})
//
//	// Use in services
//	service := NewBackupService(log)
package logger
