package log

import (
	"secLayer/conf"
	"testing"
)

func TestInitLog(t *testing.T) {
	err := InitLog("./seelog.xml", conf.SecLayerConf)
	if err != nil {
		panic(err)
	}
	defer Flush()
	Debug("test success !")
}
