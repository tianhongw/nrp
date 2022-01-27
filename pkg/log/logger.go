package log

import (
	"errors"
	"os"
	"strings"
)

type Level int

const (
	LevelDebug Level = iota - 1
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

type Format int

const (
	FormatUniversal = iota
	FormatJSON
	FormatPlain
)

var DummyLogger = new(DumbLogger)

type options struct {
	Development  bool
	Level        Level
	Format       Format
	Outputs      []string
	ErrorOutputs []string
	Prefix       string
	MaxSize      int
	MaxAge       int
	MaxBackups   int
	Compress     bool
}

type Option func(*options)

// https://github.com/grpc/grpc-go/blob/master/grpclog/loggerv2.go
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})

	Warning(args ...interface{})
	Warningf(format string, args ...interface{})
	Warningln(args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})

	Level() Level
	SetLevel(l Level)

	V(l int) bool
	Flush() error
}

func WithLevel(levelText string) Option {
	return func(o *options) {
		var level Level
		switch strings.ToLower(levelText) {
		case "debug":
			level = LevelDebug
		case "info", "":
			level = LevelInfo
		case "warning", "warn":
			level = LevelWarning
		case "error", "err":
			level = LevelError
		case "fatal":
			level = LevelFatal
		default:
			level = LevelInfo
		}

		o.Level = level
	}
}

func WithFormat(formatText string) Option {
	return func(o *options) {
		var format Format

		switch strings.ToLower(formatText) {
		case "json":
			format = FormatJSON
		case "plain", "":
			format = FormatPlain
		default:
			format = FormatPlain
		}

		o.Format = format
	}
}

func WithMaxSize(maxSize int) Option {
	return func(o *options) {
		if maxSize > 0 {
			o.MaxSize = maxSize
		}
	}
}

func WithMaxAge(maxAge int) Option {
	return func(o *options) {
		if maxAge > 0 {
			o.MaxAge = maxAge
		}
	}
}

func WithMaxBackups(maxBackups int) Option {
	return func(o *options) {
		if maxBackups > 0 {
			o.MaxBackups = maxBackups
		}
	}
}

func WithOutputs(outputs []string) Option {
	return func(o *options) {
		if len(outputs) > 0 {
			o.Outputs = make([]string, len(outputs))
			copy(o.Outputs, outputs)
		}
	}
}

func WithErrorOutputs(errorOutputs []string) Option {
	return func(o *options) {
		if len(errorOutputs) > 0 {
			o.ErrorOutputs = make([]string, len(errorOutputs))
			copy(o.ErrorOutputs, errorOutputs)
		}
	}
}

func WithPrefix(prefix string) Option {
	return func(o *options) {
		o.Prefix = prefix
	}
}

func IsDevelopment(isDev bool) Option {
	return func(o *options) {
		o.Development = isDev
	}
}

func IsCompress(isCompress bool) Option {
	return func(o *options) {
		o.Compress = isCompress
	}
}

const (
	LoggerTypeZap = "zap"
	LoggerTypeStd = "std"
)

func NewLogger(loggerType string, opts ...Option) (Logger, error) {
	o := &options{
		Development:  false,
		Level:        LevelInfo,
		Format:       FormatPlain,
		Outputs:      []string{"stderr"},
		ErrorOutputs: []string{"stderr"},
		MaxSize:      500,
		MaxAge:       7,
		MaxBackups:   10,
		Compress:     true,
	}

	for _, opt := range opts {
		opt(o)
	}

	switch strings.ToLower(loggerType) {
	case LoggerTypeZap:
		return NewZapLogger(o), nil
	case LoggerTypeStd:
		return NewStdLogger(os.Stderr, o), nil
	default:
		return nil, errors.New("unknown logger type")
	}
}
