package framework

import (
	"github.com/RealJonathanYip/framework/config"
	"github.com/RealJonathanYip/framework/log"
)

func init() {
	output := config.FrameWorkConfig.LogOutput
	log.InitLog(log.SetTarget(output.Value), log.LogFilePath(output.Path), log.LogFileRotate(output.FileRotate))
	log.SetLogLevel(config.FrameWorkConfig.LogLevel)
}
