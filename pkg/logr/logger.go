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

func (l *DefaultLogger) Debugf(msg string, args ...interface{}) {
	if l.loglevel == 0 {
		return
	}
	fmt.Fprintln(l.out, fmt.Sprintf(msg, args...))
}

func (l *DefaultLogger) Infof(msg string, args ...interface{}) {
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
