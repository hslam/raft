// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"io"
	"log"
	"os"
)

// Level defines the level for log.
// Higher levels log less info.
type Level int

const (
	//DebugLevel defines the level of debug in test environments.
	DebugLevel Level = 1
	//TraceLevel defines the level of trace in test environments.
	TraceLevel Level = 2
	//AllLevel defines the lowest level in production environments.
	AllLevel Level = 3
	//InfoLevel defines the level of info.
	InfoLevel Level = 4
	//NoticeLevel defines the level of notice.
	NoticeLevel Level = 5
	//WarnLevel defines the level of warn.
	WarnLevel Level = 6
	//ErrorLevel defines the level of error.
	ErrorLevel Level = 7
	//PanicLevel defines the level of panic.
	PanicLevel Level = 8
	//FatalLevel defines the level of fatal.
	FatalLevel Level = 9
	//OffLevel defines the level of no log.
	OffLevel Level = 10
)

var (
	logPrefix = "raft"
	logger    = New()
	redBg     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	magentaBg = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	black     = string([]byte{27, 91, 57, 48, 109})
	red       = string([]byte{27, 91, 51, 49, 109})
	green     = string([]byte{27, 91, 51, 50, 109})
	yellow    = string([]byte{27, 91, 51, 51, 109})
	blue      = string([]byte{27, 91, 51, 52, 109})
	magenta   = string([]byte{27, 91, 51, 53, 109})
	cyan      = string([]byte{27, 91, 51, 54, 109})
	white     = string([]byte{27, 91, 51, 55, 109})
	reset     = string([]byte{27, 91, 48, 109})
)

// Logger defines the logger.
type Logger struct {
	prefix       string
	out          io.Writer
	level        Level
	debugLogger  *log.Logger
	traceLogger  *log.Logger
	allLogger    *log.Logger
	infoLogger   *log.Logger
	noticeLogger *log.Logger
	warnLogger   *log.Logger
	errorLogger  *log.Logger
	panicLogger  *log.Logger
	fatalLogger  *log.Logger
}

// New creates a new Logger.
func New() *Logger {
	l := &Logger{
		prefix: logPrefix,
		out:    os.Stdout,
		level:  InfoLevel,
	}
	l.init()
	return l
}

//SetPrefix sets log's prefix
func SetPrefix(prefix string) {
	logger.SetPrefix(prefix)
}

//SetPrefix sets log's prefix
func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
	l.init()
}

//GetPrefix returns log's prefix
func GetPrefix() (prefix string) {
	return logger.GetPrefix()
}

//GetPrefix returns log's prefix
func (l *Logger) GetPrefix() (prefix string) {
	return l.prefix
}

//SetLogLevel sets log's level
func SetLogLevel(level Level) {
	logger.SetLogLevel(level)
}

//SetLogLevel sets log's level
func (l *Logger) SetLogLevel(level Level) {
	l.level = level
	l.init()
}

//SetOut sets log's writer. The out variable sets the
// destination to which log data will be written.
func SetOut(w io.Writer) {
	logger.SetOut(w)
}

//SetOut sets log's writer. The out variable sets the
// destination to which log data will be written.
func (l *Logger) SetOut(w io.Writer) {
	l.out = w
	l.init()
}

//GetLevel returns log's level
func GetLevel() Level {
	return logger.GetLevel()
}

//GetLevel returns log's level
func (l *Logger) GetLevel() Level {
	return l.level
}

func (l *Logger) init() {
	l.debugLogger = log.New(l.out, blue+"["+l.prefix+"][D]"+reset, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC)
	l.traceLogger = log.New(l.out, cyan+"["+l.prefix+"][T]"+reset, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC)
	l.allLogger = log.New(l.out, white+"["+l.prefix+"][A]"+reset, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC)
	l.infoLogger = log.New(l.out, black+"["+l.prefix+"][I]"+reset, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC)
	l.noticeLogger = log.New(l.out, green+"["+l.prefix+"][N]"+reset, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC)
	l.warnLogger = log.New(l.out, yellow+"["+l.prefix+"][W]"+reset, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC)
	l.errorLogger = log.New(l.out, red+"["+l.prefix+"][E]"+reset, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC)
	l.panicLogger = log.New(l.out, magentaBg+"["+l.prefix+"][P]"+reset, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC)
	l.fatalLogger = log.New(l.out, redBg+"["+l.prefix+"][F]"+reset, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC)
}

// Debug is equivalent to log.Print() for debug.
func (l *Logger) Debug(v ...interface{}) {
	if l.level <= DebugLevel {
		l.debugLogger.Print(v...)
	}
}

// Debugf is equivalent to log.Printf() for debug.
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level <= DebugLevel {
		l.debugLogger.Printf(format, v...)
	}
}

// Debugln is equivalent to log.Println() for debug.
func (l *Logger) Debugln(v ...interface{}) {
	if l.level <= DebugLevel {
		l.debugLogger.Println(v...)
	}
}

// Trace is equivalent to log.Print() for trace.
func (l *Logger) Trace(v ...interface{}) {
	if l.level <= TraceLevel {
		l.traceLogger.Print(v...)
	}
}

// Tracef is equivalent to log.Printf() for trace.
func (l *Logger) Tracef(format string, v ...interface{}) {
	if l.level <= TraceLevel {
		l.traceLogger.Printf(format, v...)
	}
}

// Traceln is equivalent to log.Println() for trace.
func (l *Logger) Traceln(v ...interface{}) {
	if l.level <= TraceLevel {
		l.traceLogger.Println(v...)
	}
}

// All is equivalent to log.Print() for all log.
func (l *Logger) All(v ...interface{}) {
	if l.level <= AllLevel {
		l.allLogger.Print(v...)
	}
}

// Allf is equivalent to log.Printf() for all log.
func (l *Logger) Allf(format string, v ...interface{}) {
	if l.level <= AllLevel {
		l.allLogger.Printf(format, v...)
	}
}

// Allln is equivalent to log.Println() for all log.
func (l *Logger) Allln(v ...interface{}) {
	if l.level <= AllLevel {
		l.allLogger.Println(v...)
	}
}

// Info is equivalent to log.Print() for info.
func (l *Logger) Info(v ...interface{}) {
	if l.level <= InfoLevel {
		l.infoLogger.Print(v...)
	}
}

// Infof is equivalent to log.Printf() for info.
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.infoLogger.Printf(format, v...)
	}
}

// Infoln is equivalent to log.Println() for info.
func (l *Logger) Infoln(v ...interface{}) {
	if l.level <= InfoLevel {
		l.infoLogger.Println(v...)
	}
}

// Notice is equivalent to log.Print() for notice.
func (l *Logger) Notice(v ...interface{}) {
	if l.level <= NoticeLevel {
		l.noticeLogger.Print(v...)
	}
}

// Noticef is equivalent to log.Printf() for notice.
func (l *Logger) Noticef(format string, v ...interface{}) {
	if l.level <= NoticeLevel {
		l.noticeLogger.Printf(format, v...)
	}
}

// Noticeln is equivalent to log.Println() for notice.
func (l *Logger) Noticeln(v ...interface{}) {
	if l.level <= NoticeLevel {
		l.noticeLogger.Println(v...)
	}
}

// Warn is equivalent to log.Print() for warn.
func (l *Logger) Warn(v ...interface{}) {
	if l.level <= WarnLevel {
		l.warnLogger.Print(v...)
	}
}

// Warnf is equivalent to log.Printf() for warn.
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.warnLogger.Printf(format, v...)
	}
}

// Warnln is equivalent to log.Println() for warn.
func (l *Logger) Warnln(v ...interface{}) {
	if l.level <= WarnLevel {
		l.warnLogger.Println(v...)
	}
}

// Error is equivalent to log.Print() for error.
func (l *Logger) Error(v ...interface{}) {
	if l.level <= ErrorLevel {
		l.errorLogger.Print(v...)
	}
}

// Errorf is equivalent to log.Printf() for error.
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.level <= ErrorLevel {
		l.errorLogger.Printf(format, v...)
	}
}

// Errorln is equivalent to log.Println() for error.
func (l *Logger) Errorln(v ...interface{}) {
	if l.level <= ErrorLevel {
		l.errorLogger.Println(v...)
	}
}

// Panic is equivalent to log.Print() for panic.
func (l *Logger) Panic(v ...interface{}) {
	if l.level <= PanicLevel {
		l.panicLogger.Print(v...)
	}
}

// Panicf is equivalent to log.Printf() for panic.
func (l *Logger) Panicf(format string, v ...interface{}) {
	if l.level <= PanicLevel {
		l.panicLogger.Printf(format, v...)
	}
}

// Panicln is equivalent to log.Println() for panic.
func (l *Logger) Panicln(v ...interface{}) {
	if l.level <= PanicLevel {
		l.panicLogger.Println(v...)
	}
}

// Fatal is equivalent to log.Print() for fatal.
func (l *Logger) Fatal(v ...interface{}) {
	if l.level <= FatalLevel {
		l.fatalLogger.Print(v...)
	}
}

// Fatalf is equivalent to log.Printf() for fatal.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.level <= FatalLevel {
		l.fatalLogger.Printf(format, v...)
	}
}

// Fatalln is equivalent to log.Println() for fatal.
func (l *Logger) Fatalln(v ...interface{}) {
	if l.level <= FatalLevel {
		l.fatalLogger.Println(v...)
	}
}

// Debug is equivalent to log.Print() for debug.
func Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Debugf is equivalent to log.Printf() for debug.
func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Debugln is equivalent to log.Println() for debug.
func Debugln(v ...interface{}) {
	logger.Debugln(v...)
}

// Trace is equivalent to log.Print() for trace.
func Trace(v ...interface{}) {
	logger.Trace(v...)
}

// Tracef is equivalent to log.Printf() for trace.
func Tracef(format string, v ...interface{}) {
	logger.Tracef(format, v...)
}

// Traceln is equivalent to log.Println() for trace.
func Traceln(v ...interface{}) {
	logger.Traceln(v...)
}

// All is equivalent to log.Print() for all log.
func All(v ...interface{}) {
	logger.All(v...)
}

// Allf is equivalent to log.Printf() for all log.
func Allf(format string, v ...interface{}) {
	logger.Allf(format, v...)
}

// Allln is equivalent to log.Println() for all log.
func Allln(v ...interface{}) {
	logger.Allln(v...)
}

// Info is equivalent to log.Print() for info.
func Info(v ...interface{}) {
	logger.Info(v...)
}

// Infof is equivalent to log.Printf() for info.
func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Infoln is equivalent to log.Println() for info.
func Infoln(v ...interface{}) {
	logger.Infoln(v...)
}

// Notice is equivalent to log.Print() for notice.
func Notice(v ...interface{}) {
	logger.Notice(v...)
}

// Noticef is equivalent to log.Printf() for notice.
func Noticef(format string, v ...interface{}) {
	logger.Noticef(format, v...)
}

// Noticeln is equivalent to log.Println() for notice.
func Noticeln(v ...interface{}) {
	logger.Noticeln(v...)
}

// Warn is equivalent to log.Print() for warn.
func Warn(v ...interface{}) {
	logger.Warn(v...)
}

// Warnf is equivalent to log.Printf() for warn.
func Warnf(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

// Warnln is equivalent to log.Println() for warn.
func Warnln(v ...interface{}) {
	logger.Warnln(v...)
}

// Error is equivalent to log.Print() for error.
func Error(v ...interface{}) {
	logger.Error(v...)
}

// Errorf is equivalent to log.Printf() for error.
func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// Errorln is equivalent to log.Println() for error.
func Errorln(v ...interface{}) {
	logger.Errorln(v...)
}

// Panic is equivalent to log.Print() for panic.
func Panic(v ...interface{}) {
	logger.Panic(v...)
}

// Panicf is equivalent to log.Printf() for panic.
func Panicf(format string, v ...interface{}) {
	logger.Panicf(format, v...)
}

// Panicln is equivalent to log.Println() for panic.
func Panicln(v ...interface{}) {
	logger.Panicln(v...)
}

// Fatal is equivalent to log.Print() for fatal.
func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// Fatalf is equivalent to log.Printf() for fatal.
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

// Fatalln is equivalent to log.Println() for fatal.
func Fatalln(v ...interface{}) {
	logger.Fatalln(v...)
}
