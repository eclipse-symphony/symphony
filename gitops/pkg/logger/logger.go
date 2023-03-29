package logger

import (
	"context"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type Logger = logrus.FieldLogger

func NewLogger(ctx context.Context, name string) Logger {
	return log.WithContext(ctx).WithField("logger", name)
}
