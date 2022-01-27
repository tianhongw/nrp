package log

type DumbLogger struct{}

func (logging *DumbLogger) Debug(args ...interface{}) {}

func (logging *DumbLogger) Debugf(format string, args ...interface{}) {}

func (logging *DumbLogger) Debugln(args ...interface{}) {}

func (logging *DumbLogger) Info(args ...interface{}) {}

func (logging *DumbLogger) Infof(format string, args ...interface{}) {}

func (logging *DumbLogger) Infoln(args ...interface{}) {}

func (logging *DumbLogger) Warning(args ...interface{}) {}

func (logging *DumbLogger) Warningf(format string, args ...interface{}) {}

func (logging *DumbLogger) Warningln(args ...interface{}) {}

func (logging *DumbLogger) Error(args ...interface{}) {}

func (logging *DumbLogger) Errorf(format string, args ...interface{}) {}

func (logging *DumbLogger) Errorln(args ...interface{}) {}

func (logging *DumbLogger) Fatal(args ...interface{}) {}

func (logging *DumbLogger) Fatalf(format string, args ...interface{}) {}

func (logging *DumbLogger) Fatalln(args ...interface{}) {}

func (logging *DumbLogger) Level() Level {
	return LevelInfo
}

func (logging *DumbLogger) SetLevel(l Level) {}

func (logging *DumbLogger) V(l int) bool {
	return l <= int(LevelInfo)
}

func (logging *DumbLogger) Flush() error {
	return nil
}
