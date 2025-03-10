package utils

import (
	"fmt"
	"github.com/charmbracelet/log"
)

// LogAndWrapError logs an error with context and wraps it with additional information
func LogAndWrapError(err error, message string, keyvals ...any) error {
	log.Error(message, append([]any{"error", err}, keyvals...)...)
	return fmt.Errorf("%s: %w", message, err)
}