package database

import "time"

func (r *RedisClient) Set(key string, value interface{}) error {
	return r.Client.Set(r.Ctx, key, value, 0).Err()
}

func (r *RedisClient) Get(key string) (string, error) {
	return r.Client.Get(r.Ctx, key).Result()
}

func (r *RedisClient) Del(key string) error {
	return r.Client.Del(r.Ctx, key).Err()
}
func (r *RedisClient) Ping() error {
	return r.Client.Ping(r.Ctx).Err()
}
func (r *RedisClient) Close() error {
	return r.Client.Close()
}
func (r *RedisClient) FlushAll() error {
	return r.Client.FlushAll(r.Ctx).Err()
}

func (r *RedisClient) SetWithExpiration(key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(r.Ctx, key, value, expiration).Err()
}

func (r *RedisClient) GetWithExpiration(key string) (string, time.Duration, error) {
	data, err := r.Client.Get(r.Ctx, key).Result()
	if err != nil {
		return "", 0, err
	}
	ttl := r.Client.TTL(r.Ctx, key).Val()
	return data, ttl, nil
}
func (r *RedisClient) GetSpaceLeft() (int64, error) {
	return r.Client.DBSize(r.Ctx).Result()
}
func (r *RedisClient) TotalSpace() (any, error) {
	result, err := r.Client.Info(r.Ctx).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}
