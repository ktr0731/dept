package logger

import (
	"io"
	"io/ioutil"
	"log"
)

var (
	defaultLogger = log.New(ioutil.Discard, "dept: ", 0)
)

func SetOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}

func Println(v ...interface{}) {
	defaultLogger.Println(v...)
}

func Printf(format string, v ...interface{}) {
	defaultLogger.Printf(format, v...)
}

func Fatal(v ...interface{}) {
	defaultLogger.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	defaultLogger.Fatalf(format, v...)
}
