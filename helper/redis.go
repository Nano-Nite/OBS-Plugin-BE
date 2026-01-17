package helper

import (
	"encoding/json"
	"log"
	"main/model"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

const (
	REDIS_KEY_PRODUCT_TRIAL = "PRODUCT_TRIAL"
	REDIS_KEY_PRODUCT       = "PRODUCT"
)

func InitRedis() {
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(opt)

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	RDB = rdb

	log.Println("Redis connected ✅")
}

func RedisSet(key string, data interface{}, duration time.Duration) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = RDB.Set(ctx, key, b, duration).Err()
	if err != nil {
		return err
	}

	log.Printf("Success SET to Redis with Key : %s and data : \n%v\n", key, string(b))
	return nil
}

func GetCachedProduct(key string) (result []*model.Product, err error) {
	val, err := RDB.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var p []*model.Product
	if err := json.Unmarshal([]byte(val), &p); err != nil {
		return nil, err
	}

	log.Printf("Success GET Redis with Key: %s and data : \n%v\n", key, p)
	return p, nil
}
