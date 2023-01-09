package config

import (
	"context"
	"encoding/xml"
	"github.com/RealJonathanYip/framework/context0"
	"github.com/RealJonathanYip/framework/log"
	"os"
)

type LogOutput struct {
	Path       string `xml:"path,attr"`
	FileRotate string `xml:"file_rotate,attr"`
	Value      string `xml:",chardata"`
}

type Config struct {
	XMLName   xml.Name  `xml:"server"`
	LogLevel  string    `xml:"log_level"`
	LogOutput LogOutput `xml:"log_output"`
}

var FrameWorkConfig Config

func init() {
	ctx := context0.NewContext(context.TODO())
	dirs := []string{"./", "./conf/", "../conf/", "../../conf/"}
	for _, dir := range dirs {
		data, err := os.ReadFile(dir + "framework.xml")
		if err != nil {
			continue
		}
		err = xml.Unmarshal(data, &FrameWorkConfig)
		if err != nil {
			log.Panicf(ctx, "load config file: %s framework.xml %s", dir, err)
		}

		log.Infof(ctx, "load config file %s framework.xml\n %#v", dir, FrameWorkConfig)
		return
	}

	log.Panic(ctx, "framework config not find")
}
