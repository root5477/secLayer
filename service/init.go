package service

import (
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"secLayer/conf"
	"secLayer/log"
	"time"
)

func InitEtcd() (err error) {

	conf.SecLayerContextInstance.EtcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:[]string {conf.SecLayerConf.Etcd.Addr},
		DialTimeout: time.Duration(conf.SecLayerConf.Etcd.TimeOut) * time.Second,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err is:%v\n", err)
		return err
	}
	fmt.Println("connect to etcd success!")
	log.Debugf("connect to etcd success!")
	return
}

func InitSecLayer(secLayerconf *conf.SecLayerConfig) (err error) {
	err = InitRedis(secLayerconf)
	if err != nil {
		log.Errorf("init redis failed, err:%v", err)
		fmt.Println("init redis failed, err:", err)
		return
	}

	err = InitEtcd()
	if err != nil {
		log.Errorf("init etcd failed, err:%v", err)
		return
	}

	err = LoadProductFromEtcd()
	if err != nil {
		log.Errorf("load product from etcd failed, err:%v", err)
		return
	}

	conf.SecLayerContextInstance.SecLayerConfInContext = secLayerconf
	log.Debugf("load product from etcd success!")
	return
}
