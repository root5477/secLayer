package conf

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"go.etcd.io/etcd/clientv3"
	"io/ioutil"
	"sync"
)

var (
	SecLayerConf            *SecLayerConfig
	SecLayerContextInstance = &SecLayerContext{}
)

type SecProductInfoConf struct {
	ProductId         int
	StartTime         int64
	EndTime           int64
	Status            int
	Total             int
	Left              int
	OnePersonBuyLimit int
	//抽奖概率
	BuyRate float64
	//每秒最多能够卖多少个，存于etcd动态配置
	SoldMaxLimit int
	//商品的限速控制
	SecLimitInfo *SecLimit
}

type Server struct {
	Addr string `json:"addr"`
	Port int    `json:"port"`
}

type Db struct {
	Addr     string `json:"addr"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	PassWord string `json:"pass_word"`
	Database string `json:"database"`
}

type RedisProxy2Layer struct {
	Addr                 string `json:"addr"`
	Pwd                  string `json:"pwd"`
	MaxIdle              int    `json:"max_idle"`
	MaxActive            int    `json:"max_active"`
	IdleTimeout          int    `json:"idle_timeout"`
	DbOption             int    `json:"db_option"`
	Proxy2LayerQueueName string `json:"proxy_2_layer_queue_name"`
}

//接入层redis -> 业务逻辑
type RedisLayer2Proxy struct {
	Addr                 string `json:"addr"`
	Pwd                  string `json:"pwd"`
	MaxIdle              int    `json:"max_idle"`
	MaxActive            int    `json:"max_active"`
	IdleTimeout          int    `json:"idle_timeout"`
	DbOption             int    `json:"db_option"`
	Layer2ProxyQueueName string `json:"layer_2_proxy_queue_name"`
}

type Etcd struct {
	Addr    string `json:"addr"`
	Pwd     string `json:"pwd"`
	TimeOut int    `json:"time_out"`
	SecKey  string `json:"sec_key"`
}

type Log struct {
	MinLevel string `json:"min_level"`
}

type SecLayerConfig struct {
	Server                        `json:"server"`
	Db                            `json:"db"`
	RedisProxy2Layer              `json:"redis_proxy2layer"`
	RedisLayer2Proxy              `json:"redis_layer2proxy"`
	Etcd                          `json:"etcd"`
	Log                           `json:"log"`
	WriteGoroutineNum       int   `json:"write_goroutine_num"`
	ReadGoroutineNum        int   `json:"read_goroutine_num"`
	HandleGoroutineNum      int   `json:"handle_goroutine_num"`
	Read2HandleChanSize     int   `json:"read_2_handle_chan_size"`
	Handle2WriteChanSize    int   `json:"handle_2_write_chan_size"`
	MaxReqWaitTimeout       int64 `json:"max_req_wait_timeout"`
	SendToWriteChanTimeOut  int   `json:"send_to_write_chan_time_out"`
	SendToHandleChanTimeOut int   `json:"send_to_handle_chan_time_out"`
	SecProductInfoMap       map[int]*SecProductInfoConf

	SecKillTokenPassWord string `json:"sec_kill_token_pass_word"`
}

type SecLayerContext struct {
	Proxy2LayerRedisPool *redis.Pool
	Layer2ProxyRedisPool *redis.Pool
	EtcdClient           *clientv3.Client
	RWSecProductLock     sync.RWMutex

	SecLayerConfInContext *SecLayerConfig
	WaitGroups            *sync.WaitGroup
	Read2HandleChan       chan *SecRequestWithCookie
	Handle2WriteChan      chan *SecResponse

	HistoryMap     map[string]*UserBuyHistory
	HistoryMapLock sync.Mutex
	//商品的计数
	ProductCountMgrIns *ProductCountMgr
}

type SecKillReq struct {
	Product     int    `json:"product"`
	Src         string `json:"src"`
	AuthCode    string `json:"auth_code"`
	Time        string `json:"time"`
	Nance       string `json:"nance"`
	AccessTime  int64  `json:"access_time"`
	ClientAddr  string `json:"client_addr"`
	ClientRefer string `json:"client_refer"`
}

type UserCookie struct {
	UserId         string `json:"user_id"`
	UserCookieAuth string `json:"user_cookie_auth"`
}

type SecRequestWithCookie struct {
	SecKillReq
	UserCookie
}

type SecResponse struct {
	ProductId int    `json:"product_id"`
	UserId    string `json:"user_id"`
	Token     string `json:"token"`
	TokenTime int64  `json:"token_time"`
	Code      int    `json:"code"` //用于标示是否抢到
}

func EnvInitLayer(confPath string) *SecLayerConfig {
	//InitLogger()
	Config := &SecLayerConfig{}
	confByte, err := ioutil.ReadFile(confPath)
	if err != nil {
		fmt.Printf("读取配置文件%s失败, err is:%v", confPath, err)
		panic(err)
	}
	errUnmarshal := json.Unmarshal(confByte, Config)
	if errUnmarshal != nil {
		fmt.Printf("json.Unmarshal失败, err is:%v", errUnmarshal)
		panic(errUnmarshal)
	}
	return Config
}

func init() {
	SecLayerConf = EnvInitLayer("./conf/seclayer.json")
	SecLayerConf.SecProductInfoMap = make(map[int]*SecProductInfoConf)
	SecLayerContextInstance.WaitGroups = &sync.WaitGroup{}
	SecLayerContextInstance.Read2HandleChan = make(chan *SecRequestWithCookie, SecLayerConf.Read2HandleChanSize)
	SecLayerContextInstance.Handle2WriteChan = make(chan *SecResponse, SecLayerConf.Handle2WriteChanSize)
	SecLayerContextInstance.HistoryMap = make(map[string]*UserBuyHistory, 100000)
	SecLayerContextInstance.ProductCountMgrIns = NewProductMgr()
}
