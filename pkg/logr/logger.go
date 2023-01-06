package logr

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

type Logger interface {
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
}

type DefaultLogger struct {
	out      io.Writer
	loglevel Level
}

func New(out io.Writer, level Level) Logger {
	return &DefaultLogger{
		out:      out,
		loglevel: level,
	}
}

type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	ErrorLevel Level = "error"
)

func ParseLevel(level string) (Level, error) {
	switch level {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "error":
		return ErrorLevel, nil
	default:
		var l Level
		return l, fmt.Errorf("invalid level: '%s'", l)
	}
}

func (l *DefaultLogger) Debugf(msg string, args ...interface{}) {
	if l.loglevel == DebugLevel {
		c := color.New(color.FgHiBlue)
		c.Fprintln(l.out, fmt.Sprintf(msg, args...))
	}
}

func (l *DefaultLogger) Infof(msg string, args ...interface{}) {
	if l.loglevel == DebugLevel || l.loglevel == InfoLevel {
		fmt.Fprintln(l.out, fmt.Sprintf(msg, args...))
	}
}

func (l *DefaultLogger) Errorf(msg string, args ...interface{}) {
	c := color.New(color.FgRed)
	c.Fprintln(l.out, fmt.Sprintf(msg, args...))
}
