package testsupport

import (
	"bytes"
	"flag"
	"io"
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
			Logger: logr.New(io.MultiWriter(os.Stdout, buff), logr.DebugLevel),
			buff:   buff,
		}
	}
	return &TestLogger{
		Logger: logr.New(io.MultiWriter(io.Discard, buff), logr.InfoLevel),
		buff:   buff,
	}
}

type TestLogger struct {
	logr.Logger
	buff *bytes.Buffer
}

func (l *TestLogger) Output() string {
	return l.buff.String()
}
