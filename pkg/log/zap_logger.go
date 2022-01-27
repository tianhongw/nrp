package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

type ZapLogger struct {
	*zap.Logger

	level zap.AtomicLevel
}

func NewZapLogger(opt *options) *ZapLogger {
	var encoderCfg zapcore.EncoderConfig
	var zapLevel zap.AtomicLevel

	if opt.Development {
		zapLevel = zap.NewAtomicLevelAt(zap.DebugLevel)

		encoderCfg = zap.NewDevelopmentEncoderConfig()
	} else {
		zapLevel = zap.NewAtomicLevelAt(zapcore.Level(opt.Level))

		encoderCfg = zap.NewProductionEncoderConfig()
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	var zapEncoder zapcore.Encoder
	switch opt.Format {
	case FormatJSON:
		zapEncoder = zapcore.NewJSONEncoder(encoderCfg)
	default:
		zapEncoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return zapLevel.Enabled(lvl) && lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return zapLevel.Enabled(lvl) && lvl < zapcore.ErrorLevel
	})

	var cores []zapcore.Core

	for _, output := range opt.Outputs {
		cores = append(cores, zapcore.NewCore(zapEncoder, newWriteSyncer(&lumberjack.Logger{
			Filename:   output,
			MaxSize:    opt.MaxSize,
			MaxAge:     opt.MaxAge,
			MaxBackups: opt.MaxBackups,
			LocalTime:  true,
			Compress:   opt.Compress,
		}), lowPriority))
	}

	for _, errorOutput := range opt.ErrorOutputs {
		cores = append(cores, zapcore.NewCore(zapEncoder, newWriteSyncer(&lumberjack.Logger{
			Filename:   errorOutput,
			MaxSize:    opt.MaxSize,
			MaxAge:     opt.MaxAge,
			MaxBackups: opt.MaxBackups,
			LocalTime:  true,
			Compress:   opt.Compress,
		}), highPriority))
	}

	return &ZapLogger{
		Logger: zap.New(
			zapcore.NewTee(cores...),
			zap.AddStacktrace(zap.ErrorLevel),
			zap.AddCaller(),
			zap.AddCallerSkip(1),
			zap.WrapCore(func(core zapcore.Core) zapcore.Core {
				if !opt.Development {
					return zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
				}

				return core
			}),
		),
		level: zapLevel,
	}
}

func newWriteSyncer(logger *lumberjack.Logger) (w zapcore.WriteSyncer) {
	switch logger.Filename {
	case "stdout":
		w = zapcore.Lock(os.Stdout)
	case "stderr":
		w = zapcore.Lock(os.Stderr)
	default:
		w = zapcore.AddSync(logger)
	}
	return
}

func (logging *ZapLogger) Debug(args ...interface{}) {
	logging.Logger.Debug(fmt.Sprint(args...))
}

func (logging *ZapLogger) Debugf(format string, args ...interface{}) {
	logging.Logger.Debug(fmt.Sprintf(format, args...))
}

func (logging *ZapLogger) Debugln(args ...interface{}) {
	logging.Logger.Debug(fmt.Sprint(args...))
}

func (logging *ZapLogger) Info(args ...interface{}) {
	logging.Logger.Info(fmt.Sprint(args...))
}

func (logging *ZapLogger) Infof(format string, args ...interface{}) {
	logging.Logger.Info(fmt.Sprintf(format, args...))
}

func (logging *ZapLogger) Infoln(args ...interface{}) {
	logging.Logger.Info(fmt.Sprint(args...))
}

func (logging *ZapLogger) Warning(args ...interface{}) {
	logging.Logger.Warn(fmt.Sprint(args...))
}

func (logging *ZapLogger) Warningf(format string, args ...interface{}) {
	logging.Logger.Warn(fmt.Sprintf(format, args...))
}

func (logging *ZapLogger) Warningln(args ...interface{}) {
	logging.Logger.Warn(fmt.Sprint(args...))
}

func (logging *ZapLogger) Error(args ...interface{}) {
	logging.Logger.Error(fmt.Sprint(args...))
}

func (logging *ZapLogger) Errorf(format string, args ...interface{}) {
	logging.Logger.Error(fmt.Sprintf(format, args...))
}

func (logging *ZapLogger) Errorln(args ...interface{}) {
	logging.Logger.Error(fmt.Sprint(args...))
}

func (logging *ZapLogger) Fatal(args ...interface{}) {
	logging.Logger.Fatal(fmt.Sprint(args...))
}

func (logging *ZapLogger) Fatalf(format string, args ...interface{}) {
	logging.Logger.Fatal(fmt.Sprintf(format, args...))
}

func (logging *ZapLogger) Fatalln(args ...interface{}) {
	logging.Logger.Fatal(fmt.Sprint(args...))
}

func (logging *ZapLogger) Level() Level {
	zapLevel := logging.level.Level()

	switch zapLevel {
	case zap.DebugLevel:
		return LevelDebug
	case zap.InfoLevel:
		return LevelInfo
	case zap.WarnLevel:
		return LevelWarning
	case zap.ErrorLevel:
		return LevelError
	case zap.PanicLevel, zap.DPanicLevel, zap.FatalLevel:
		return LevelFatal
	default:
		return LevelInfo
	}
}

func (logging *ZapLogger) SetLevel(l Level) {
	var zapLevel zapcore.Level
	switch l {
	case LevelDebug:
		zapLevel = zapcore.DebugLevel
	case LevelInfo:
		zapLevel = zapcore.InfoLevel
	case LevelWarning:
		zapLevel = zapcore.WarnLevel
	case LevelError:
		zapLevel = zapcore.ErrorLevel
	case LevelFatal:
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	logging.level.SetLevel(zapLevel)
}

func (logging *ZapLogger) V(l int) bool {
	return l <= int(LevelInfo)
}

func (logging *ZapLogger) Flush() error {
	return logging.Logger.Sync()
}
