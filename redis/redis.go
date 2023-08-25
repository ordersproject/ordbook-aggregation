package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"ordbook-aggregation/config"
	"ordbook-aggregation/major"
)

const (
	ex                           = "86400"
	CacheGetUtxo_                = "cache_get_utxo_"
	UtxoTypeDummy_               = "dummy_"
	UtxoTypeBidY_                = "bid_y_"
	UtxoTypeMultiSigInscription_ = "multi_sig_inscription_"
	CacheGetClaimOrder_          = "cache_get_claim_order_"
	CacheGetPoolClaimOrder_      = "cache_get_pool_claim_order_"
)

var addressEndpoint = ""
var redisPassword = ""

//var redisDb int = 3

var (
	redisDbUtxo int = 3
)

var (
	_redisManager *RedisManager
)

type RedisManager struct {
	pools map[int]*redis.Pool
}

func (r *RedisManager) InitRedis() {
	r.pools = newPools()
	if r.pools == nil {
		panic("redis init error. ")
	}
	major.Println("Init redis success")
}

func GetRedisManager() *RedisManager {
	if _redisManager == nil {
		_redisManager = new(RedisManager)
		_redisManager.InitRedis()
	}
	return _redisManager
}

func InitRedisManager() {
	if _redisManager == nil {
		_redisManager = new(RedisManager)
		_redisManager.InitRedis()
	}
}

func newPools() map[int]*redis.Pool {
	addressEndpoint = config.RedisEndpoint
	redisPassword = config.RedisPassword
	redisDbUtxo = config.RedisDbUtxo
	pools := make(map[int]*redis.Pool)
	pools[redisDbUtxo] = newOnePool(addressEndpoint, redisPassword, redisDbUtxo)
	return pools
}

func newOnePool(endpoint, password string, db int) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   100,
		MaxActive: 10000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", endpoint, redis.DialPassword(password))
			if err != nil {
				panic(err.Error())
				return nil, err
			}
			if _, err = c.Do("SELECT", db); err != nil {
				c.Close()
				panic(err.Error())
				return nil, err
			}
			return c, err
		},
	}
}

func (r *RedisManager) Set(db int, key string, value interface{}, times int) (interface{}, error) {
	c := r.pools[db].Get()
	if c == nil {
		major.Println("get redis Idle error")
		return nil, errors.New("get redis Idle error")
	}
	defer c.Close()

	data, err := json.Marshal(value)
	if err != nil {
		return nil, errors.New("decode error.")
	}
	if times <= 0 {
		v, err := c.Do("SET", key, data) // s
		if err != nil {
			major.Println(fmt.Sprintf("set redis error %s", err))
			return nil, err
		}
		return v, nil
	} else {
		v, err := c.Do("SET", key, data, "EX", times, "NX") // s
		if err != nil {
			major.Println(fmt.Sprintf("set redis error %s", err))
			return nil, err
		}
		return v, nil
	}
}

func (r *RedisManager) Get(db int, key string) (interface{}, error) {
	c := r.pools[db].Get()
	if c == nil {
		major.Println(fmt.Sprintf("get redis Idle error"))
		return nil, errors.New("get redis Idle error")
	}
	defer c.Close()
	value, err := c.Do("Get", key)
	//fmt.Println(value)
	if err != nil {
		return nil, err
	} else {
		if value != nil {
			return value, nil
		}
	}
	return nil, errors.New("error ")
}

func (r *RedisManager) GetList(db int, keyRegex string) (interface{}, error) {
	c := r.pools[db].Get()
	if c == nil {
		major.Println(fmt.Sprintf("get redis Idle error"))
		return nil, errors.New("get redis Idle error")
	}
	defer c.Close()
	value, err := c.Do("KEYS", keyRegex)
	if err != nil {
		return nil, err
	} else {
		if value != nil {
			return value, nil
		}
	}
	return nil, errors.New("error ")
}

func (r *RedisManager) Delete(db int, key string) (interface{}, error) {
	c := r.pools[db].Get()
	if c == nil {
		major.Println(fmt.Sprintf("get redis Idle error"))
		return nil, errors.New("get redis Idle error")
	}
	defer c.Close()
	v, err := c.Do("DEL", key)
	if err != nil {
		return nil, err
	}
	return v, nil
}
