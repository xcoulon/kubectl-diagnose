package logr

import (
	"bytes"
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
	buff     *bytes.Buffer
	out      io.Writer
	loglevel int
}

func New(out io.Writer) *DefaultLogger {
	buff := bytes.NewBuffer(nil)
	return &DefaultLogger{
		buff:     buff,
		out:      io.MultiWriter(out, buff),
		loglevel: 0,
	}
}

const (
	DebugLevel int = 1
)

func (l *DefaultLogger) SetLevel(level int) {
	l.loglevel = level
}

// Tee uses the given `out` as a secondary output.
// Usage: `Tee(os.Stdout)` to see in the console what's record in this terminal during the tests
// Note: it should be configured at the beginning of a test
// DEPRECATED
func (l *DefaultLogger) Tee(out io.Writer) {
	l.out = io.MultiWriter(l.out, out)
}

func (l *DefaultLogger) Debugf(msg string, args ...interface{}) {
	if l.loglevel == 0 {
		return
	}
	if msg == "" {
		fmt.Fprintln(l.out, "")
		return
	}
	fmt.Fprintln(l.out, fmt.Sprintf(msg, args...))
}

func (l *DefaultLogger) Infof(msg string, args ...interface{}) {
	if msg == "" {
		fmt.Fprintln(l.out, "")
		return
	}
	fmt.Fprintln(l.out, fmt.Sprintf(msg, args...))
}

func (l *DefaultLogger) Errorf(msg string, args ...interface{}) {
	c := color.New(color.FgRed)
	c.Fprintln(l.out, fmt.Sprintf(msg, args...))
}

func (l *DefaultLogger) Error(msg string) {
	c := color.New(color.FgHiRed)
	c.Fprintln(l.out, msg)
}

func (l *DefaultLogger) Output() string {
	return l.buff.String()
}
