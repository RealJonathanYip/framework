package config

import (
	"context"
	"encoding/xml"
	"github.com/RealJonathanYip/framework/log"
	"os"
)

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
