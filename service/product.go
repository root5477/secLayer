package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"secLayer/conf"
	"secLayer/log"
	"time"
)


func LoadProductFromEtcd() (err error) {
	err = InitEtcd()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	resp, err := conf.SecLayerContextInstance.EtcdClient.Get(ctx, conf.SecLayerConf.Etcd.SecKey)
	defer cancel()
	if err != nil {
		return err
	}
	var SecInfos []conf.SecProductInfoConf
	for _, v := range resp.Kvs {
		err = json.Unmarshal([]byte(v.Value), &SecInfos)
		if err != nil {
			return err
		}
	}
	fmt.Println("secInfos:", SecInfos)
	conf.SecLayerContextInstance.RWSecProductLock.Lock()
	for i := range SecInfos {
		SecInfos[i].SecLimitInfo = &conf.SecLimit{} // 初始化secLimit
		conf.SecLayerConf.SecProductInfoMap[SecInfos[i].ProductId] = &SecInfos[i]
	}
	conf.SecLayerContextInstance.RWSecProductLock.Unlock()
	log.Debugf("get product info success, data:%v", SecInfos)
	InitSecProductWatcher()
	log.Debugf("init etcd watch success!")
	return nil
}

func InitSecProductWatcher()  {
	go WatchSecProductInfoKey(conf.SecLayerConf.Etcd.SecKey)
}

func WatchSecProductInfoKey(key string)  {
	key1WatchChan := conf.SecLayerContextInstance.EtcdClient.Watch(context.Background(), key)
	var secProductInfo []conf.SecProductInfoConf
	var getConfSucc = true

	for wresp := range key1WatchChan {
		for _, ev := range wresp.Events {
			fmt.Printf("Type: %s Key:%s Value:%s\n", ev.Type, ev.Kv.Key, ev.Kv.Value)

			if ev.Type == clientv3.EventTypeDelete {
				log.Warnf("key[%s]的config deleted", key)
				continue
			}

			if ev.Type == clientv3.EventTypePut && string(ev.Kv.Key) == key {
				err := json.Unmarshal(ev.Kv.Value, &secProductInfo)
				if err != nil {
					log.Errorf("key [%s] unmarshal[%v] failed, err is:%v", key, secProductInfo, err)
					getConfSucc = false
					continue
				}
			}
			log.Debugf("get config from etcd, type:%v, key:%v, value:%v", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}

		if getConfSucc {
			log.Debugf("get config from etcd success, %v", secProductInfo)
			updateSecProductInfo(secProductInfo)
		}
	}
}

func updateSecProductInfo(infos []conf.SecProductInfoConf)  {
	tmp := make(map[int] *conf.SecProductInfoConf, 1024)
	for _, v := range infos {
		productInfo := v
		productInfo.SecLimitInfo = &conf.SecLimit{}
		tmp[v.ProductId] = &productInfo
	}
	conf.SecLayerContextInstance.RWSecProductLock.Lock()
	conf.SecLayerConf.SecProductInfoMap = tmp
	conf.SecLayerContextInstance.RWSecProductLock.Unlock()
}
