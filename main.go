package main

import (
	"fmt"
	"secLayer/conf"
	"secLayer/log"
	"secLayer/service"
)

func main()  {
	//1.加载配置文件
	if conf.SecLayerConf.Server.Addr == "" {
		err := fmt.Errorf("init config err")
		panic(err)
	}
	//2.init log
	err := log.InitLog("./log/seelog.xml", conf.SecLayerConf)
	if err != nil {
		panic(err)
	}
	log.Debugf("init log success!")
	log.Debugf("init config success, conf:[%v]", conf.SecLayerConf)

	//3.初始化秒杀逻辑层
	err = service.InitSecLayer(conf.SecLayerConf)
	if err != nil {
		log.Errorf("init sec kill failed, err:%v", err)
		panic(err)
	}
	log.Debugf("init sec kill success!")

	//4.运行业务逻辑
	err = service.Run()
	if err != nil {
		log.Errorf("service run failed, err:%v", err)
		panic(err)
	}
	log.Infof("service run success!")
	defer log.Flush()
}