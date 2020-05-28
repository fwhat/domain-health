package common

import (
	"github.com/qjues/domain-health/config"
	"github.com/op/go-logging"
	"os"
)

var Log = logging.MustGetLogger("domain-health")
var backendToModuleLeveled logging.LeveledBackend

// 初始化日志
func init() {
	var logBackend = make([]logging.Backend, 0)
	// 如果有logfile配置则增加写入到logfile中
	if config.Instance.LogFile != "" {
		file, err := os.OpenFile(config.Instance.LogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			panic("log file '" + config.Instance.LogFile + "' can not open")
		}

		format := logging.MustStringFormatter(`%{time:2006-01-02 15:04:05.000} %{longfunc}[%{shortfile}] ▶ [%{level:.4s}] %{message}`)

		backendToFile := logging.NewLogBackend(file, "", 0)
		backendToConsoleFormatter := logging.NewBackendFormatter(backendToFile, format)
		backendToFileLeveled := logging.AddModuleLevel(backendToConsoleFormatter)
		backendToFileLeveled.SetLevel(getLevel(), "")

		logBackend = append(logBackend, backendToFileLeveled)
	}

	backendToConsole := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(`%{color}%{time:15:04:05.000} %{shortfile} ▶ [%{level:.4s}]%{color:reset} %{message}`)

	backendToConsoleFormatter := logging.NewBackendFormatter(backendToConsole, format)
	backendToConsoleLeveled := logging.AddModuleLevel(backendToConsoleFormatter)
	backendToConsoleLeveled.SetLevel(getLevel(), "")

	logBackend = append(logBackend, backendToConsoleFormatter)

	logging.SetBackend(logBackend...)
}

func SetLoggerLevel(level logging.Level) {
	backendToModuleLeveled.SetLevel(level, "")
}

func getLevel() logging.Level {
	switch config.Instance.DebugLevel {
	case "critical":
		return logging.CRITICAL
	case "error":
		return logging.ERROR
	case "warning":
		return logging.WARNING
	case "notice":
		return logging.NOTICE
	case "info":
		return logging.INFO
	case "debug":
		return logging.DEBUG
	}

	return logging.NOTICE
}
