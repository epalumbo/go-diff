package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// RedisRepository is the Redis-backed implementation of the DiffRepository contract
type RedisRepository struct {
	client redis.Cmdable
}

// NewRedisRepository creates a new instance of the RedisRepository implementation
func NewRedisRepository(opt *redis.Options) *RedisRepository {
	return &RedisRepository{redis.NewClient(opt)}
}

// SaveDataSide saves data sides to Redis
func (r *RedisRepository) SaveDataSide(ID string, side string, data []byte) error {
	if len(ID) == 0 {
		return errors.New("cannot save diff side data without ID")
	}
	if len(side) == 0 {
		return errors.New("cannot save diff side data without side")
	}
	k := keyOf(ID)
	return r.client.HSet(ctx, k, side, string(data)).Err()
}

// GetDataSidesByID gets data sides by ID from Redis
func (r *RedisRepository) GetDataSidesByID(ID string) (map[string][]byte, error) {
	h, err := r.client.HGetAll(ctx, keyOf(ID)).Result()
	if err != nil {
		return nil, err
	}
	ds := make(map[string][]byte, len(h))
	for k, v := range h {
		ds[k] = []byte(v)
	}
	return ds, nil
}

func keyOf(ID string) string {
	return fmt.Sprintf("diff:%s", ID)
}
