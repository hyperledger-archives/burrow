package log

//XXX this pkg is being deprecated in favour of logrus

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

type LogLevel uint8

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

const (
	Version = "0.1.1" // atomic running
)

//--------------------------------------------------------------------------------
// thread safe logger that fires messages from multiple packages one at a time

func init() {
	go readLoop()
}

var (
	// control access to loggers
	mtx     sync.Mutex
	loggers = make(map[string]*Logger)

	// access to writers managed by channels
	writer    io.Writer = os.Stdout
	errWriter io.Writer = os.Stderr

	chanBuffer = 100
	writeCh    = make(chan []byte, chanBuffer)
	errorCh    = make(chan []byte, chanBuffer)

	quitCh = make(chan struct{})

	running uint32 // atomic
)

type Logger struct {
	Level LogLevel
	Pkg   string

	// these are here for easy access
	// for functions that want the writer
	Writer    *SafeWriter
	ErrWriter *SafeWriter
}

// add a default logger with pkg name
func AddLogger(pkg string) *Logger {
	l := &Logger{
		Level:     LogLevelError,
		Pkg:       pkg,
		Writer:    NewSafeWriter(writeCh),
		ErrWriter: NewSafeWriter(errorCh),
	}
	mtx.Lock()
	loggers[pkg] = l
	mtx.Unlock()
	return l
}

func SetLogLevelGlobal(level LogLevel) {
	mtx.Lock()
	defer mtx.Unlock()

	for _, l := range loggers {
		l.Level = level
	}
}

// set levels for individual packages
func SetLogLevel(pkg string, level LogLevel) {
	mtx.Lock()
	defer mtx.Unlock()

	if l, ok := loggers[pkg]; ok {
		l.Level = level
		if level > LogLevelInfo {
			// TODO: wrap the writers to print [<pkg>]
		}
	}
}

// set level and writer for all loggers
func SetLoggers(level LogLevel, w io.Writer, ew io.Writer) {
	mtx.Lock()
	defer mtx.Unlock()
	for _, l := range loggers {
		l.Level = level
		if l.Level > LogLevelInfo {
			// TODO: wrap the writers to print [<pkg>]
		}
	}
	writer = w
	errWriter = ew
}

//--------------------------------------------------------------------------------
// concurrency

func readLoop() {
	if atomic.CompareAndSwapUint32(&running, 0, 1) {
	LOOP:
		for {
			select {
			case b := <-writeCh:
				writer.Write(b)
			case b := <-errorCh:
				errWriter.Write(b)
			case <-quitCh:
				break LOOP

			}
		}
	}
}

func flush(ch chan []byte, w io.Writer) {
	for {
		select {
		case b := <-ch:
			w.Write(b)
		default:
			return
		}
	}
}

// Flush the log channels. Concurrent users of the logger should quit before
// Flush() is called to ensure it completes.
func Flush() {
	if atomic.CompareAndSwapUint32(&running, 1, 0) {
		flush(writeCh, writer)
		flush(errorCh, errWriter)
		quitCh <- struct{}{}
	}
}

//--------------------------------------------------------------------------------
// a SafeWriter implements Writer and fires its bytes over the channel
// to be written to the writer or errWriter

type SafeWriter struct {
	ch chan []byte
}

func NewSafeWriter(ch chan []byte) *SafeWriter {
	return &SafeWriter{ch}
}

func (sw *SafeWriter) Write(b []byte) (int, error) {
	sw.ch <- b
	return len(b), nil
}

// thread safe writes
func writef(s string, args ...interface{}) {
	writeCh <- []byte(fmt.Sprintf(s, args...))
}

func writeln(s ...interface{}) {
	writeCh <- []byte(fmt.Sprintln(s...))
}

func errorf(s string, args ...interface{}) {
	errorCh <- []byte(fmt.Sprintf(s, args...))
}

func errorln(s ...interface{}) {
	errorCh <- []byte(fmt.Sprintln(s...))
}

//--------------------------------------------------------------------------------
// public logger functions

// Printf and Println write to the Writer no matter what
func (l *Logger) Printf(s string, args ...interface{}) {
	writef(s, args...)
}

func (l *Logger) Println(s ...interface{}) {
	writeln(s...)
}

// Errorf and Errorln write to the ErrWriter no matter what
func (l *Logger) Errorf(s string, args ...interface{}) {
	errorf(s, args...)
}

func (l *Logger) Errorln(s ...interface{}) {
	errorln(s...)
}

// Warnf and Warnln  write to the Writer if log level >= 1
func (l *Logger) Warnf(s string, args ...interface{}) {
	if l.Level > LogLevelError {
		writef(s, args...)
	}
}

func (l *Logger) Warnln(s ...interface{}) {
	if l.Level > LogLevelError {
		writeln(s...)
	}
}

// Infof and Infoln write to the Writer if log level >= 2
func (l *Logger) Infof(s string, args ...interface{}) {
	if l.Level > LogLevelWarn {
		writef(s, args...)
	}
}

func (l *Logger) Infoln(s ...interface{}) {
	if l.Level > LogLevelWarn {
		writeln(s...)
	}
}

// Debugf and Debugln write to the Writer if log level >= 3
func (l *Logger) Debugf(s string, args ...interface{}) {
	if l.Level > LogLevelInfo {
		writef(s, args...)
	}
}

func (l *Logger) Debugln(s ...interface{}) {
	if l.Level > LogLevelInfo {
		writeln(s...)
	}
}
