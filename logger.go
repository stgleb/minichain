package minichain

import (
	"github.com/Sirupsen/logrus"
)

var logInstance *logrus.Logger

func InitLogger(config *Config) {
	logInstance = logrus.New()
	logLevel := logrus.AllLevels[config.Main.LogLevel]
	logInstance.Level = logLevel

}

func GetLogger() *logrus.Logger {
	return logInstance
}
