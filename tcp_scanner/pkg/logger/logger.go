package logger

import (
	"log"
)

type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

type stdLogger struct{}

func New() Logger { return &stdLogger{} }

func (l *stdLogger) Info(args ...interface{})          { log.Println(args...) }
func (l *stdLogger) Infof(f string, a ...interface{})  { log.Printf(f, a...) }
func (l *stdLogger) Warn(args ...interface{})          { log.Println(args...) }
func (l *stdLogger) Warnf(f string, a ...interface{})  { log.Printf(f, a...) }
func (l *stdLogger) Error(args ...interface{})         { log.Println(args...) }
func (l *stdLogger) Errorf(f string, a ...interface{}) { log.Printf(f, a...) }
