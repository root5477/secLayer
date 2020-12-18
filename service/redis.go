package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math/rand"
	"secLayer/conf"
	"secLayer/log"
	"time"
)


func InitRedis(secLayerconf *conf.SecLayerConfig) (err error) {
	conf.SecLayerContextInstance.Proxy2LayerRedisPool = &redis.Pool{
		MaxIdle:secLayerconf.RedisProxy2Layer.MaxIdle,
		MaxActive:secLayerconf.RedisProxy2Layer.MaxActive,
		IdleTimeout:time.Duration(secLayerconf.RedisProxy2Layer.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			dialPwd := redis.DialPassword(secLayerconf.RedisProxy2Layer.Pwd)
			dbOption := redis.DialDatabase(secLayerconf.RedisProxy2Layer.DbOption)
			Conns, err := redis.Dial("tcp", secLayerconf.RedisProxy2Layer.Addr, dialPwd, dbOption)
			if err != nil {
				return nil, err
			}
			return Conns, nil
		},
	}
	c1 := conf.SecLayerContextInstance.Proxy2LayerRedisPool.Get()
	defer c1.Close()
	_, err = c1.Do("ping")
	if err != nil {
		return err
	}

	conf.SecLayerContextInstance.Layer2ProxyRedisPool = &redis.Pool{
		MaxIdle:secLayerconf.RedisLayer2Proxy.MaxIdle,
		MaxActive:secLayerconf.RedisLayer2Proxy.MaxActive,
		IdleTimeout:time.Duration(secLayerconf.RedisLayer2Proxy.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			dialPwd := redis.DialPassword(secLayerconf.RedisLayer2Proxy.Pwd)
			dbOption := redis.DialDatabase(secLayerconf.RedisLayer2Proxy.DbOption)
			Conns, err := redis.Dial("tcp", secLayerconf.RedisLayer2Proxy.Addr, dialPwd, dbOption)
			if err != nil {
				return nil, err
			}
			return Conns, nil
		},
	}
	c2 := conf.SecLayerContextInstance.Layer2ProxyRedisPool.Get()
	defer c2.Close()
	_, err = c2.Do("ping")
	if err != nil {
		return err
	}
	return nil
}


func RunProcess()  {
	for i :=0; i < conf.SecLayerContextInstance.SecLayerConfInContext.ReadGoroutineNum; i ++ {
		conf.SecLayerContextInstance.WaitGroups.Add(1)
		go HandleReader()
	}

	for i :=0; i < conf.SecLayerContextInstance.SecLayerConfInContext.WriteGoroutineNum; i ++ {
		conf.SecLayerContextInstance.WaitGroups.Add(1)
		go HandleWrite()
	}

	for i := 0; i< conf.SecLayerContextInstance.SecLayerConfInContext.HandleGoroutineNum; i++ {
		conf.SecLayerContextInstance.WaitGroups.Add(1)
		go HandleUser()
	}
	log.Debugf("all process goroutine started")
	conf.SecLayerContextInstance.WaitGroups.Wait()
	log.Debugf("wait all goroutine exited")
	return
}

func HandleReader()  {
	//双层for循环，复用redis连接
	log.Debugf("read goroutine running")
	for {
		conn := conf.SecLayerContextInstance.Proxy2LayerRedisPool.Get()
		for {
			data, err := redis.String(conn.Do("BLPOP", conf.SecLayerConf.RedisProxy2Layer.Proxy2LayerQueueName, 0))
			if err != nil {
				log.Errorf("pop from sec_queue failed, err:%v", err)
				break
			}
			log.Debugf("pop from sec_queue, data:%v", data)
			var secReq conf.SecRequestWithCookie
			err = json.Unmarshal([]byte(data), &secReq)
			if err != nil {
				log.Errorf("json unmarshal failed, err:%v", err)
				continue
			}
			now := time.Now().Unix()
			if now - secReq.AccessTime >= conf.SecLayerConf.MaxReqWaitTimeout {
				log.Warnf("req[%v] is expire.", secReq)
				continue
			}
			timer := time.NewTicker(time.Millisecond * time.Duration(conf.SecLayerConf.SendToHandleChanTimeOut))
			select {
			case conf.SecLayerContextInstance.Read2HandleChan <- &secReq:
			case <- timer.C:
				log.Warnf("send to handle chan timeout, req:%v", secReq)
				break
			}

		}
		conn.Close()
	}
}

func HandleWrite()  {
	log.Debugf("handle write running")
	for res := range conf.SecLayerContextInstance.Handle2WriteChan {
		err := SendToRedis(res)
		if err != nil {
			log.Errorf("send to redis failed, err:%v, res:%v", err, res)
			continue
		}
	}
}

func SendToRedis(res *conf.SecResponse) (err error) {
	conn := conf.SecLayerContextInstance.Layer2ProxyRedisPool.Get()
	data, err := json.Marshal(res)
	if err != nil {
		return
	}
	_, err = redis.String(conn.Do("rpush", "layer2proxy_queue", data))
	if err != nil {
		log.Errorf("rpush to redis layer2proxy_queue failed, err is:%v, data:%v", err, data)
		return
	}
	return
}

func HandleUser()  {
	log.Debugf("handle user running")
	for req := range conf.SecLayerContextInstance.Read2HandleChan {
		log.Debugf("begin process request:%v", req)
		res, err := HandleSecKill(req)
		if err != nil {
			log.Errorf("process request %v failed, err:%v", req, err)
			res = &conf.SecResponse{
				Code:ErrServiceBusy,
			}
		}
		//超时处理，防止处理不过来，chan满了的话，丢弃
		timer := time.NewTicker(time.Millisecond * time.Duration(conf.SecLayerConf.SendToWriteChanTimeOut))
		select {
		case conf.SecLayerContextInstance.Handle2WriteChan <- res:
		case <- timer.C:
			log.Warnf("send to resp chan timeout, res:%v", res)
			break
		}

	}
}

func HandleSecKill(req *conf.SecRequestWithCookie) (res *conf.SecResponse, err error) {
	conf.SecLayerContextInstance.RWSecProductLock.RLock()
	defer conf.SecLayerContextInstance.RWSecProductLock.Unlock()

	res = &conf.SecResponse{}
	product, ok := conf.SecLayerContextInstance.SecLayerConfInContext.SecProductInfoMap[req.Product]
	if !ok {
		log.Errorf("not found product:%v", req.Product)
		res.Code = ErrNotFoundProductId
		return
	}
	//1--是否售罄
	if product.Status == ProductStatusSoldOut {
		res.Code = ErrSoldOut
		return
	}
	//2--是否超速
	now := time.Now().Unix()
	alreadySoldCount := product.SecLimitInfo.CheckNumNow(now)
	if alreadySoldCount > product.SoldMaxLimit {
		res.Code = ErrRetry
		return
	}
	//3--判断用户是否已经购买过
	conf.SecLayerContextInstance.HistoryMapLock.Lock()
	userHistory, ok := conf.SecLayerContextInstance.HistoryMap[req.UserId]
	if !ok {
		userHistory = &conf.UserBuyHistory{
			History: make(map[int] int, 16),
		}
		conf.SecLayerContextInstance.HistoryMap[req.UserId] = userHistory
	}
	historyCount := userHistory.GetProductBuyCount(req.Product)
	if historyCount > conf.SecLayerContextInstance.SecLayerConfInContext.SecProductInfoMap[req.Product].SoldMaxLimit {
		res.Code = ErrAlreadyBuyLimit
		return
	}
	conf.SecLayerContextInstance.HistoryMapLock.Unlock()
	//4.是否总数超限
	curSoldCount := conf.SecLayerContextInstance.ProductCountMgrIns.Count(req.Product)
	if curSoldCount >= product.Total {
		res.Code = ErrSoldOut
		product.Status = ProductStatusSoldOut
		return
	}
	//5.是否黑名单（根据用户等级）
	//6.随机抽奖
	if curRate := rand.Float64(); curRate > product.BuyRate {
		res.Code = ErrRetry
		return
	}
	//7.买到，更新总数
	userHistory.Add(req.Product, 1)
	conf.SecLayerContextInstance.ProductCountMgrIns.Add(req.Product, 1)
	res.Code = ErrSecKillSucc
	//8.生成token user_id && product_id && timestamp && 密钥
	tokenStr := fmt.Sprintf("userId=%v&productId=%v&timestamp=%v&security=%v", req.UserId, req.Product, now, conf.SecLayerConf.SecKillTokenPassWord)
	res.Token = fmt.Sprintf("%x", md5.Sum([]byte(tokenStr)))
	res.TokenTime = now
	return
}
