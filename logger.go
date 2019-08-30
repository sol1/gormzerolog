package gormzerolog

import (
	"github.com/rs/zerolog"
)

// New create logger object for *gorm.DB from *zeroLog.Logger
// By default it logs with debug level.
func New(origin *zerolog.Logger, opts ...LoggerOption) *Logger {
	l := &Logger{
		ZRLog: origin,
		Level: zerolog.DebugLevel,
		Event: origin.WithLevel(zerolog.DebugLevel),
	}

	for _, o := range opts {
		o(l)
	}

	return l
}

// Logger is an alternative implementation of *gorm.Logger
type Logger struct {
	ZRLog *zerolog.Logger
	Level zerolog.Level
	Event *zerolog.Event
}

// LoggerOption is an option for Logger.
type LoggerOption func(*Logger)

// WithLevel returns Logger option that sets level for gorm logs.
// It affects only general logs, e.g. those that contain SQL queries.
// Errors will be logged with error level independently of this option.
func WithLevel(level zerolog.Level) LoggerOption {
	return func(l *Logger) {
		l.Level = level
		l.Event = l.ZRLog.WithLevel(level)
	}
}

// Print passes arguments to Println
func (l *Logger) Print(values ...interface{}) {
	l.Println(values)
}

// Println format & print log
func (l *Logger) Println(values []interface{}) {
	l.toZRFields(l.createLog(values))
	l.Event.Send()
}
