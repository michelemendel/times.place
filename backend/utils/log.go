package utils

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
)

// Constants for custom log levels
const (
	LevelTrace  = slog.Level(-8)
	LevelNotice = slog.Level(2)
	LevelFatal  = slog.Level(12)
)

// LevelNames maps custom log levels to their string representations
var LevelNames = map[slog.Leveler]string{
	LevelTrace:  "TRACE",
	LevelNotice: "NOTICE",
	LevelFatal:  "FATAL",
}

// NewLogger creates a new logger with a specific component name
func NewLogger(component string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				var severity string
				switch {
				case level < slog.LevelInfo:
					severity = "DEBUG"
				case level == slog.LevelInfo:
					severity = "INFO"
				case level == slog.LevelWarn:
					severity = "WARNING"
				case level >= slog.LevelError:
					severity = "ERROR"
				default:
					severity = "DEFAULT"
				}
				return slog.Attr{Key: "severity", Value: slog.StringValue(severity)}
			}

			if a.Key == slog.SourceKey {
				filename := a.Value.Any().(*slog.Source).File
				lineNumber := a.Value.Any().(*slog.Source).Line
				ps := strings.Split(filename, "/")
				a.Value = slog.StringValue(fmt.Sprintf("%s/%s:%v", ps[len(ps)-2], ps[len(ps)-1], lineNumber))
			}

			return a
		},
		// We add a component attribute to the log entry for more granular filtering.
		// We tried to use logName, but it seems that Cloud Run overrides it with the default logName.
	})).With("component", component)
}

var SlogShort = slog.New(slog.NewTextHandler(os.Stderr, nil))

var LogL = log.New(
	os.Stderr,
	"",
	log.Llongfile,
)

var LogS = log.New(
	os.Stderr,
	"",
	log.Lshortfile,
)
