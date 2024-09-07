package framework

import (
	"encoding/xml"
	"github.com/RealJonathanYip/framework/config"
	"github.com/RealJonathanYip/framework/context0"
	"github.com/RealJonathanYip/framework/log"
)

type logOutput struct {
	Path       string `xml:"path,attr"`
	FileRotate string `xml:"file_rotate,attr"`
	Value      string `xml:",chardata"`
}

type localConfig struct {
	XMLName   xml.Name  `xml:"server"`
	LogLevel  string    `xml:"log_level"`
	LogOutput logOutput `xml:"log_output"`
	Env       string    `xml:"env"`
}

var (
	frameWorkConfig localConfig
	env             string = "prod"
)

func Init(configFile string) {
	if err := config.ReadXml(context0.NewContext(), configFile, &frameWorkConfig, true); err != nil {
		panic(err)
	}

	log.InitLog(log.SetTarget(frameWorkConfig.LogOutput.Value),
		log.LogFilePath(frameWorkConfig.LogOutput.Path), log.LogFileRotate(frameWorkConfig.LogOutput.FileRotate))
	log.SetLogLevel(frameWorkConfig.LogLevel)

	if frameWorkConfig.Env != "" {
		env = frameWorkConfig.Env
	}
}

func Env() string {
	return env
}
