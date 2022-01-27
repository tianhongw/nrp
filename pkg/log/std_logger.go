package log

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"sync"
)

var _ Logger = (*StdLogger)(nil)

type StdLogger struct {
	*stdlog.Logger
	mu sync.RWMutex

	logLevel Level
}

func NewStdLogger(out io.Writer, opts *options) *StdLogger {
	stdLogger := &StdLogger{
		Logger:   stdlog.New(out, opts.Prefix, stdlog.Ldate|stdlog.Ltime|stdlog.Lmicroseconds|stdlog.LUTC),
		logLevel: opts.Level,
	}

	return stdLogger
}

func (l *StdLogger) Debug(args ...interface{}) {
	if !l.canLogAt(LevelDebug) {
		return
	}
	l.printPrefix("DEBUG: ", args...)
}

func (l *StdLogger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *StdLogger) Debugln(args ...interface{}) {
	if !l.canLogAt(LevelDebug) {
		return
	}
	l.printlnPrefix("DEBUG: ", args...)
}

func (l *StdLogger) Info(args ...interface{}) {
	if !l.canLogAt(LevelInfo) {
		return
	}
	l.printPrefix("INFO: ", args...)
}

func (l *StdLogger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *StdLogger) Infoln(args ...interface{}) {
	if !l.canLogAt(LevelInfo) {
		return
	}
	l.printlnPrefix("INFO: ", args...)
}

func (l *StdLogger) Warning(args ...interface{}) {
	if !l.canLogAt(LevelWarning) {
		return
	}
	l.printPrefix("WARNING: ", args...)
}

func (l *StdLogger) Warningf(format string, args ...interface{}) {
	l.Warning(fmt.Sprintf(format, args...))
}

func (l *StdLogger) Warningln(args ...interface{}) {
	if !l.canLogAt(LevelWarning) {
		return
	}
	l.printlnPrefix("WARNING: ", args...)
}

func (l *StdLogger) Error(args ...interface{}) {
	if !l.canLogAt(LevelError) {
		return
	}
	l.printPrefix("ERROR: ", args...)
}

func (l *StdLogger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *StdLogger) Errorln(args ...interface{}) {
	if !l.canLogAt(LevelError) {
		return
	}
	l.printlnPrefix("ERROR: ", args...)
}

func (l *StdLogger) Fatal(args ...interface{}) {
	if !l.canLogAt(LevelFatal) {
		return
	}
	l.printPrefix("FATAL: ", args...)
	os.Exit(1)
}

func (l *StdLogger) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format, args...))
}

func (l *StdLogger) Fatalln(args ...interface{}) {
	if !l.canLogAt(LevelFatal) {
		return
	}
	l.printlnPrefix("FATAL: ", args...)
	os.Exit(1)
}

func (l *StdLogger) Level() Level {
	l.mu.RLock()
	v := l.logLevel
	l.mu.RUnlock()
	return v
}

func (l *StdLogger) SetLevel(level Level) {
	l.mu.Lock()
	l.logLevel = level
	l.mu.Unlock()
}

func (l *StdLogger) V(level int) bool {
	return level <= int(LevelInfo)
}

func (l *StdLogger) Flush() error {
	return nil
}

func (l *StdLogger) printPrefix(prefix string, args ...interface{}) {
	args = append([]interface{}{prefix}, args...)
	l.Print(args...)
}

func (l *StdLogger) printlnPrefix(prefix string, args ...interface{}) {
	args = append([]interface{}{prefix}, args...)
	l.Println(args...)
}

func (l *StdLogger) canLogAt(v Level) bool {
	return v >= l.Level()
}
