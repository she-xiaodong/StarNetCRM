package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/starnet/crm/internal/config"
)

var Client *redis.Client

// Init 初始化Redis连接
func Init(cfg config.RedisConfig) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}

	return nil
}

// Close 关闭连接
func Close() {
	if Client != nil {
		_ = Client.Close()
	}
}

// Set 设置缓存
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return Client.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存
func Get(ctx context.Context, key string, dest interface{}) error {
	data, err := Client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Delete 删除缓存
func Delete(ctx context.Context, keys ...string) error {
	return Client.Del(ctx, keys...).Err()
}

// PathCacheKey 路径查询缓存键
func PathCacheKey(startID, endID string) string {
	return "path:" + startID + ":" + endID
}
