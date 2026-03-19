package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	CachePrefix = "radio:"
)

type RedisCacheConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func DefaultRedisCacheConfig() *RedisCacheConfig {
	return &RedisCacheConfig{
		Addr:         getEnv("TOUGHRADIUS_REDIS_HOST", "localhost:6379"),
		Password:     getEnv("TOUGHRADIUS_REDIS_PASSWORD", ""),
		DB:           getEnvInt("TOUGHRADIUS_REDIS_DB", 0),
		PoolSize:     getEnvInt("TOUGHRADIUS_REDIS_POOL_SIZE", 100),
		MinIdleConns: getEnvInt("TOUGHRADIUS_REDIS_MIN_IDLE", 10),
		DialTimeout:  getEnvDuration("TOUGHRADIUS_REDIS_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:  getEnvDuration("TOUGHRADIUS_REDIS_READ_TIMEOUT", 3*time.Second),
		WriteTimeout: getEnvDuration("TOUGHRADIUS_REDIS_WRITE_TIMEOUT", 3*time.Second),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return time.Duration(intVal) * time.Second
		}
	}
	return defaultVal
}

type RedisCache struct {
	client *redis.Client
	mu     sync.RWMutex
	prefix string
}

func NewRedisCache(config *RedisCacheConfig) (*RedisCache, error) {
	if config == nil {
		config = DefaultRedisCacheConfig()
	}

	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisCache{
		client: client,
		prefix: CachePrefix,
	}, nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) Client() *redis.Client {
	return c.client
}

func (c *RedisCache) Key(tenantID int64, resource, key string) string {
	return fmt.Sprintf("%stenant:%d:%s:%s", c.prefix, tenantID, resource, key)
}

func (c *RedisCache) UserKey(tenantID int64, username string) string {
	return c.Key(tenantID, "user", username)
}

func (c *RedisCache) NasKey(tenantID int64, nasIP string) string {
	return c.Key(tenantID, "nas", nasIP)
}

func (c *RedisCache) SessionKey(tenantID int64, sessionID string) string {
	return c.Key(tenantID, "session", sessionID)
}

func (c *RedisCache) VoucherKey(tenantID int64, code string) string {
	return c.Key(tenantID, "voucher", code)
}

func (c *RedisCache) ProfileKey(tenantID int64, profileID int64) string {
	return c.Key(tenantID, "profile", fmt.Sprintf("%d", profileID))
}

func (c *RedisCache) ProductKey(tenantID int64, productID int64) string {
	return c.Key(tenantID, "product", fmt.Sprintf("%d", productID))
}

func (c *RedisCache) TenantStatsKey(tenantID int64) string {
	return c.Key(tenantID, "stats", "summary")
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

func (c *RedisCache) DeleteTenantKeys(ctx context.Context, tenantID int64) error {
	pattern := fmt.Sprintf("%stenant:%d:*", c.prefix, tenantID)
	return c.DeletePattern(ctx, pattern)
}

func (c *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *RedisCache) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

func (c *RedisCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, key, ttl).Err()
}

func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

type CachedUser struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Status       string `json:"status"`
	ProfileId    int64  `json:"profile_id"`
	AddrPool     string `json:"addr_pool"`
	UpRate       int64  `json:"up_rate"`
	DownRate     int64  `json:"down_rate"`
	BindMac      int    `json:"bind_mac"`
	BindVlan     int    `json:"bind_vlan"`
	ExpireTime   int64  `json:"expire_time"`
	ActiveNum    int    `json:"active_num"`
	InputOctets  int64  `json:"input_octets"`
	OutputOctets int64  `json:"output_octets"`
	InputPackets int64  `json:"input_packets"`
	OutputPackets int64 `json:"output_packets"`
	SessionTime  int64  `json:"session_time"`
	MACAddr      string `json:"mac_addr"`
}

type CachedNAS struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Ipaddr    string `json:"ipaddr"`
	Identifier string `json:"identifier"`
	Secret     string `json:"secret"`
	Status    string `json:"status"`
}

type CachedSession struct {
	SessionID   string `json:"session_id"`
	Username    string `json:"username"`
	NasAddr     string `json:"nas_addr"`
	FramedIP    string `json:"framed_ip"`
	StartTime   int64  `json:"start_time"`
	InputOctets int64  `json:"input_octets"`
	OutputOctets int64 `json:"output_octets"`
}

func (c *RedisCache) SetUser(ctx context.Context, tenantID int64, username string, user *CachedUser, ttl time.Duration) error {
	return c.Set(ctx, c.UserKey(tenantID, username), user, ttl)
}

func (c *RedisCache) GetUser(ctx context.Context, tenantID int64, username string) (*CachedUser, error) {
	var user CachedUser
	err := c.Get(ctx, c.UserKey(tenantID, username), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *RedisCache) DeleteUser(ctx context.Context, tenantID int64, username string) error {
	return c.Delete(ctx, c.UserKey(tenantID, username))
}

func (c *RedisCache) SetNAS(ctx context.Context, tenantID int64, nasIP string, nas *CachedNAS, ttl time.Duration) error {
	return c.Set(ctx, c.NasKey(tenantID, nasIP), nas, ttl)
}

func (c *RedisCache) GetNAS(ctx context.Context, tenantID int64, nasIP string) (*CachedNAS, error) {
	var nas CachedNAS
	err := c.Get(ctx, c.NasKey(tenantID, nasIP), &nas)
	if err != nil {
		return nil, err
	}
	return &nas, nil
}

func (c *RedisCache) DeleteNAS(ctx context.Context, tenantID int64, nasIP string) error {
	return c.Delete(ctx, c.NasKey(tenantID, nasIP))
}

func (c *RedisCache) IncrementSessionCount(ctx context.Context, tenantID int64, username string) (int64, error) {
	key := fmt.Sprintf("%stenant:%d:session_count:%s", c.prefix, tenantID, username)
	return c.client.Incr(ctx, key).Result()
}

func (c *RedisCache) DecrementSessionCount(ctx context.Context, tenantID int64, username string) (int64, error) {
	key := fmt.Sprintf("%stenant:%d:session_count:%s", c.prefix, tenantID, username)
	result, err := c.client.Decr(ctx, key).Result()
	if result < 0 {
		c.client.Set(ctx, key, 0, 0)
		return 0, err
	}
	return result, err
}

func (c *RedisCache) GetSessionCount(ctx context.Context, tenantID int64, username string) (int64, error) {
	key := fmt.Sprintf("%stenant:%d:session_count:%s", c.prefix, tenantID, username)
	return c.client.Get(ctx, key).Int64()
}
