package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	infoLogger  *log.Logger
	debugLogger *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	logFs       *os.File
	wrLock      *sync.RWMutex
	logLevel    int
	currentDay  int
	logFile     string
)

const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

func init() {
	wrLock = &sync.RWMutex{}
}

func SetLogLevel(level int) {
	logLevel = level
}

func SetLogFile(file string) {
	var err error
	currentDay = time.Now().YearDay()

	logFile = file
	splitLogFile := strings.Split(file, ".log")
	newLogFile := splitLogFile[0] + "_" + time.Now().Format("20060102") + ".log"

	logFs, err = os.OpenFile(newLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatal(err)
	}

	debugLogger = log.New(logFs, "[DEBUG] ", log.LstdFlags)
	infoLogger = log.New(logFs, "[INFO] ", log.LstdFlags)
	warnLogger = log.New(logFs, "[WARN] ", log.LstdFlags)
	errorLogger = log.New(logFs, "[ERROR] ", log.LstdFlags)
}

func checkDay() {
	wrLock.Lock()
	defer wrLock.Unlock()

	var err error
	day := time.Now().YearDay()
	if day != currentDay {
		logFs.Close()

		newLogFile := logFile + "_" + time.Now().Format("20060102")
		logFs, err = os.OpenFile(newLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			log.Fatal(err)
		}
		debugLogger = log.New(logFs, "[DEBUG] ", log.LstdFlags)
		infoLogger = log.New(logFs, "[INFO] ", log.LstdFlags)
		warnLogger = log.New(logFs, "[WARN] ", log.LstdFlags)
		errorLogger = log.New(logFs, "[ERROR] ", log.LstdFlags)
		currentDay = day
	}
}

func Debug(format string, i ...any) {
	if logLevel <= DebugLevel {
		checkDay()
		debugLogger.Printf(getCaller(2)+": "+format, i...)
	}
}

func Info(format string, i ...any) {
	if logLevel <= InfoLevel {
		checkDay()
		infoLogger.Printf(getCaller(2)+": "+format, i...)
	}
}

func Warn(format string, i ...any) {
	if logLevel <= WarnLevel {
		checkDay()
		warnLogger.Printf(getCaller(2)+": "+format, i...)
	}
}

func Error(format string, i ...any) {
	fmt.Println(111)
	if logLevel <= ErrorLevel {
		checkDay()
		errorLogger.Printf(getCaller(2)+": "+format, i...)
	}
}

func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	fileLine := file + ":" + strconv.Itoa(line)

	return fileLine
}
