package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const issueCodeScript = `
if redis.call("EXISTS", KEYS[2]) == 1 then
	return 0
end
redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
redis.call("SET", KEYS[2], "1", "PX", ARGV[3])
return 1
`

const consumeValueScript = `
local value = redis.call("GET", KEYS[1])
if not value then
	return false
end
redis.call("DEL", KEYS[1])
return value
`

type RedisCodeStore struct {
	client redis.UniversalClient
}

func NewRedisCodeStore(client redis.UniversalClient) *RedisCodeStore {
	return &RedisCodeStore{client: client}
}

func (s *RedisCodeStore) Issue(ctx context.Context, phone, code string, ttl, cooldown time.Duration) error {
	result, err := s.client.Eval(
		ctx,
		issueCodeScript,
		[]string{"auth:sms:code:" + phone, "auth:sms:cooldown:" + phone},
		code,
		ttl.Milliseconds(),
		cooldown.Milliseconds(),
	).Int()
	if err != nil {
		return err
	}
	if result != 1 {
		return ErrCodeRateLimited
	}
	return nil
}

func (s *RedisCodeStore) Verify(ctx context.Context, phone, code string) error {
	key := "auth:sms:code:" + phone
	stored, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) || stored != code {
		return ErrInvalidCode
	}
	if err != nil {
		return err
	}
	deleted, err := s.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if deleted != 1 {
		return ErrInvalidCode
	}
	return nil
}

type RedisSessionStore struct {
	client redis.UniversalClient
}

func NewRedisSessionStore(client redis.UniversalClient) *RedisSessionStore {
	return &RedisSessionStore{client: client}
}

func (s *RedisSessionStore) Create(ctx context.Context, userID int64, ttl time.Duration) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	if err := s.client.Set(ctx, refreshKey(token), strconv.FormatInt(userID, 10), ttl).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func (s *RedisSessionStore) Consume(ctx context.Context, token string) (int64, error) {
	if token == "" {
		return 0, ErrInvalidRefreshToken
	}
	value, err := s.client.Eval(ctx, consumeValueScript, []string{refreshKey(token)}).Text()
	if errors.Is(err, redis.Nil) || value == "" {
		return 0, ErrInvalidRefreshToken
	}
	if err != nil {
		return 0, err
	}
	userID, err := strconv.ParseInt(value, 10, 64)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidRefreshToken
	}
	return userID, nil
}

func (s *RedisSessionStore) Delete(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.client.Del(ctx, refreshKey(token)).Err()
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func refreshKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "auth:refresh:" + base64.RawURLEncoding.EncodeToString(sum[:])
}
