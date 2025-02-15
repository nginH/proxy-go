package logs

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

var (
	fnRuntime  = runtime.Caller
	origLogger = logrus.New()
	baseLogger = logger{entry: logrus.NewEntry(origLogger)}
)

func InitLogger() {
	formatter := &logrus.JSONFormatter{
		TimestampFormat:  "2006-01-02 15:04:05",
		CallerPrettyfier: nil,
		PrettyPrint:      false,
	}
	SetFormatter(formatter)
	// SetLevel(logrus.DebugLevel.String()) //FIXME: Set log level from .env file
	SetLevel(os.Getenv("LOG_LEVEL")) //FIXME: Set log level from .env file
	SetLevel("logrus.DebugLevel")    //FIXME: Set log level from .env file
}

type Logger interface {
	Debug(...interface{})
	Debugln(...interface{})
	Debugf(string, ...interface{})

	Info(...interface{})
	Infoln(...interface{})
	Infof(string, ...interface{})

	Warn(...interface{})
	Warnln(...interface{})
	Warnf(string, ...interface{})

	Error(...interface{})
	Errorln(...interface{})
	Errorf(string, ...interface{})

	Fatal(...interface{})
	Fatalln(...interface{})
	Fatalf(string, ...interface{})

	WithField(key string, value interface{}) Logger
	WithFields(fields logrus.Fields) Logger

	SetLevel(string) error
}

type logger struct {
	entry *logrus.Entry
}

func (l logger) sourced() *logrus.Entry {
	_, file, line, ok := fnRuntime(2)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		list := strings.Split(file, "/")
		file = strings.Join(list[len(list)-3:], "/")
	}
	return l.entry.WithField("prefix", fmt.Sprintf("%s:%d", file, line))
}

func (l logger) Debug(args ...interface{}) {
	l.sourced().Debug(args...)
}

func (l logger) Debugln(args ...interface{}) {
	l.sourced().Debugln(args...)
}

func (l logger) Debugf(format string, args ...interface{}) {
	l.sourced().Debugf(format, args...)
}

func (l logger) Info(args ...interface{}) {
	l.sourced().Info(args...)
}

func (l logger) Infoln(args ...interface{}) {
	l.sourced().Infoln(args...)
}

func (l logger) Infof(format string, args ...interface{}) {
	l.sourced().Infof(format, args...)
}

func (l logger) Warn(args ...interface{}) {
	l.sourced().Warn(args...)
}

func (l logger) Warnln(args ...interface{}) {
	l.sourced().Warnln(args...)
}

func (l logger) Warnf(format string, args ...interface{}) {
	l.sourced().Warnf(format, args...)
}

func (l logger) Error(args ...interface{}) {
	l.sourced().Error(args...)
}

func (l logger) Errorln(args ...interface{}) {
	l.sourced().Errorln(args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	l.sourced().Errorf(format, args...)
}

func (l logger) Fatal(args ...interface{}) {
	l.sourced().Fatal(args...)
}

func (l logger) Fatalln(args ...interface{}) {
	l.sourced().Fatalln(args...)
}

func (l logger) Fatalf(format string, args ...interface{}) {
	l.sourced().Fatalf(format, args...)
}

func (l logger) WithFields(fields logrus.Fields) Logger {
	return logger{entry: l.entry.WithFields(fields)}
}

func (l logger) SetLevel(level string) error {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	l.entry.Logger.Level = lvl
	return nil
}

func (l logger) WithField(key string, value interface{}) Logger {
	return logger{entry: l.entry.WithField(key, value)}
}

func FromFiberCtx(c *fiber.Ctx) Logger {
	l, ok := c.Locals("logger").(Logger)
	if !ok {
		return baseLogger
	}
	return l
}

func SetFormatter(formatter logrus.Formatter) {
	origLogger.Formatter = formatter
}

func SetLevel(level string) {
	baseLogger.SetLevel(level)
}

func Base() Logger {
	return baseLogger
}

func WithField(key string, value interface{}) Logger {
	return baseLogger.WithField(key, value)
}

func WithFields(fields logrus.Fields) Logger {
	return baseLogger.WithFields(fields)
}

func Debug(args ...interface{}) {
	baseLogger.sourced().Debug(args...)
}

func Debugln(args ...interface{}) {
	baseLogger.sourced().Debugln(args...)
}

func Debugf(format string, args ...interface{}) {
	baseLogger.sourced().Debugf(format, args...)
}

func Info(args ...interface{}) {
	baseLogger.sourced().Info(args...)
}

func Infoln(args ...interface{}) {
	baseLogger.sourced().Infoln(args...)
}

func Infof(format string, args ...interface{}) {
	baseLogger.sourced().Infof(format, args...)
}

func Warn(args ...interface{}) {
	baseLogger.sourced().Warn(args...)
}

func Warnln(args ...interface{}) {
	baseLogger.sourced().Warnln(args...)
}

func Warnf(format string, args ...interface{}) {
	baseLogger.sourced().Warnf(format, args...)
}

func Error(args ...interface{}) {
	baseLogger.sourced().Error(args...)
}

func Errorln(args ...interface{}) {
	baseLogger.sourced().Errorln(args...)
}

func Errorf(format string, args ...interface{}) {
	baseLogger.sourced().Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	baseLogger.sourced().Fatal(args...)
}

func Fatalln(args ...interface{}) {
	baseLogger.sourced().Fatalln(args...)
}

func Fatalf(format string, args ...interface{}) {
	baseLogger.sourced().Fatalf(format, args...)
}
