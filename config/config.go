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
	if err := ReadXml(context0.NewContext(), "./conf/framework.xml", &FrameWorkConfig, true); err != nil {
		panic(err)
	}
}

func ReadXml(ctx context.Context, file string, output interface{}, panicOnFail ...bool) error {
	doPanic := false
	if len(panicOnFail) > 0 {
		doPanic = panicOnFail[0]
	}

	data, err := os.ReadFile(file)
	if err != nil && doPanic {
		log.Panicf(ctx, "load config file fail: %s:%s", file, err)
	} else {
		log.Warningf(ctx, "load config file fail: %s:%s", file, err)
		return err
	}

	err = xml.Unmarshal(data, output)
	if err != nil && doPanic {
		log.Panicf(ctx, "load config file fail: %s:%s", file, err)
	} else {
		log.Warningf(ctx, "load config file fail: %s:%s", file, err)
		return err
	}

	log.Infof(ctx, "load config file %s \n %#v", file, output)
	return nil
}
