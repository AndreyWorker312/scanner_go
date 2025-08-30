package logger

import (
	"log"
	"os"
)

type LoggerInterface interface {
	Info(msg string)
	Infof(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
}

type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
}

func New() *Logger {
	return &Logger{
		infoLog:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(msg string) {
	l.infoLog.Println(msg)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.infoLog.Printf(format, args...)
}

func (l *Logger) Error(msg string) {
	l.errorLog.Println(msg)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.errorLog.Printf(format, args...)
}
