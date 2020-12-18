package log

import (
	"fmt"
	"secLayer/conf"
)

func InitLog(path string, conf *conf.SecLayerConfig) error {
	err := Init(path, conf.Log.MinLevel)
	if err != nil {
		return fmt.Errorf("Log init error:%v", err)
	}

	Info("Log init success")
	return nil
}
