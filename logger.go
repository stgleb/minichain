package minichain

import (
	"github.com/Sirupsen/logrus"
)

var logInstance *logrus.Logger

func InitLogger(logLevel int) {
	logInstance = logrus.New()
	logInstance.Level = logrus.AllLevels[logLevel]

}

func GetLogger() *logrus.Logger {
	return logInstance
}
