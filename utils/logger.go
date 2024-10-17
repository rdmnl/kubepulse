// KubePulse - Kubernetes Cluster Monitor (TUI)
//
// Author: Erdem Unal
// Year: 2024
// Version: 0.1.0
// License: MIT

package utils

import (
	"log"
	"os"
)

const (
    INFO  = "INFO"
    WARN  = "WARN"
    ERROR = "ERROR"
)

var (
    infoLogger  *log.Logger
    warnLogger  *log.Logger
    errorLogger *log.Logger
)

func init() {
    logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("Failed to open log file: %v", err)
    }

    infoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    warnLogger = log.New(logFile, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
    errorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(message string) {
    infoLogger.Println(message)
}

func Warn(message string) {
    warnLogger.Println(message)
}

func Error(message string) {
    errorLogger.Println(message)
}

func Errorf(format string, v ...interface{}) {
    errorLogger.Printf(format, v...)
}
