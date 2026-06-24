package logger

import (
	"flag"
	"os"
	"strings"
	"sync"

	"github.com/golang/glog"
)

type severity int

const (
	debugLevel severity = iota
	infoLevel
	warningLevel
	errorLevel
)

var configureOnce sync.Once
var minSeverity = infoLevel

func ConfigureFromEnv() {
	configureOnce.Do(func() {
		levelName := strings.ToUpper(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
		if levelName == "" {
			levelName = "INFO"
		}

		setFlag("logtostderr", envDefault("LOG_TO_STDERR", "true"))
		setFlag("stderrthreshold", envDefault("LOG_STDERR_THRESHOLD", "ERROR"))
		if logDir := strings.TrimSpace(os.Getenv("LOG_DIR")); logDir != "" {
			setFlag("log_dir", logDir)
		}

		switch levelName {
		case "DEBUG":
			minSeverity = debugLevel
			setFlag("v", "1")
		case "INFO":
			minSeverity = infoLevel
			setFlag("v", "0")
		case "WARN", "WARNING":
			minSeverity = warningLevel
			setFlag("v", "0")
		case "ERROR":
			minSeverity = errorLevel
			setFlag("v", "0")
		default:
			minSeverity = infoLevel
			setFlag("v", "0")
			Warningf("unknown LOG_LEVEL=%q, using INFO", levelName)
		}
	})
}

func Debugf(format string, args ...any) {
	if minSeverity > debugLevel {
		return
	}
	glog.V(1).Infof(format, args...)
}

func Infof(format string, args ...any) {
	if minSeverity > infoLevel {
		return
	}
	glog.Infof(format, args...)
}

func Warningf(format string, args ...any) {
	if minSeverity > warningLevel {
		return
	}
	glog.Warningf(format, args...)
}

func Errorf(format string, args ...any) {
	if minSeverity > errorLevel {
		return
	}
	glog.Errorf(format, args...)
}

func Flush() {
	glog.Flush()
}

func envDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func setFlag(name string, value string) {
	if flag.Lookup(name) == nil {
		return
	}
	_ = flag.Set(name, value)
}
