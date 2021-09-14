package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/yurykabanov/throttler"
)

type redisClient interface {
	Keys(ctx context.Context, pattern string) *redis.StringSliceCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
}

type Storage struct {
	prefix string
	client redisClient
}

func New(prefix string, client *redis.Client) *Storage {
	return &Storage{
		prefix: prefix,
		client: client,
	}
}

func (s *Storage) CountLastExecuted(ctx context.Context, action throttler.Action, after time.Time) (int, error) {
	pattern := fmt.Sprintf("%s-%s-*", s.prefix, action.GroupID())

	list, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, err
	}

	return len(list), nil
}

func (s *Storage) SaveSuccessfulExecution(ctx context.Context, action throttler.Action, at time.Time, expiration time.Duration) error {
	// Special global counter to avoid rare possibility to lose actions executed at the same time
	counter, err := s.client.Incr(ctx, s.prefix).Result()
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s-%s-%d-%d", s.prefix, action.GroupID(), at.UnixNano(), counter)

	_, err = s.client.SetNX(ctx, key, at.UnixNano(), expiration).Result()
	if err != nil {
		return err
	}

	return nil
}
