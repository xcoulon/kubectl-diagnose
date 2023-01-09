package testsupport

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/xcoulon/kubectl-diagnose/pkg/logr"
)

var logLevel string

func init() {
	flag.StringVar(&logLevel, "loglevel", "info", "log level to set [debug|info|error|]")
}

func NewLogger() *TestLogger {
	buff := bytes.NewBuffer(nil)
	switch logLevel {
	case "debug":
		return &TestLogger{
			logger: logr.New(os.Stdout, logr.DebugLevel),
			buff:   buff,
		}
	}
	return &TestLogger{
		logger: logr.New(buff, logr.InfoLevel),
		buff:   buff,
	}
}

type TestLogger struct {
	logger logr.Logger
	buff   *bytes.Buffer
}

var _ logr.Logger = &TestLogger{}

func (l *TestLogger) Debugf(msg string, args ...interface{}) {
	l.logger.Debugf(msg, args...)
	// do not write 'debug' message to internal buffer
}

func (l *TestLogger) Infof(msg string, args ...interface{}) {
	l.logger.Infof(msg, args...)
	fmt.Fprintf(l.buff, msg, args...)
}

func (l *TestLogger) Errorf(msg string, args ...interface{}) {
	l.logger.Errorf(msg, args...)
	fmt.Fprintf(l.buff, msg, args...)
}

func (l *TestLogger) Output() string {
	return l.buff.String()
}
