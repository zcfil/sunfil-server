package redisClient

import (
	"github.com/go-redis/redis"
	"log"
	"time"
	"zcfil-server/global"
)

var RedisClient = new(redis.Client)

func init() {
	RedisNewClient(global.ZC_CONFIG.Redis.Addr, global.ZC_CONFIG.Redis.Password, global.ZC_CONFIG.Redis.DB)
}

func RedisNewClient(addr string, password string, db int) {
	//timeout := time.Duration(readTimeout)
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       db,       // use default DB
	})
	if err := RedisClient.Ping().Err(); err != nil {
		panic(err.Error())
	}
}

func NewDefaultRedisStore() *RedisStore {
	return &RedisStore{
		Expiration: time.Minute * 3,
		PreKey:     "CONTRACTFORMAL_",
	}
}

type RedisStore struct {
	Expiration time.Duration
	PreKey     string
}

func (rs *RedisStore) Set(key string, value string) error {
	err := RedisClient.Set(rs.PreKey+key, value, rs.Expiration).Err()
	if err != nil {
		log.Println("key ", key, " RedisStoreSetError!", err.Error())
		return err
	}
	return nil
}

func (rs *RedisStore) Get(key string) string {
	data, err := RedisClient.Get(rs.PreKey + key).Result()
	if err != nil {
		log.Println("key ", key, " RedisStoreGetError", err.Error())
		return ""
	}
	return data
}
