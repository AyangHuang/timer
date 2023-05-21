package redis

import (
	"context"
	"errors"
	"github.com/gomodule/redigo/redis"
	"timer/common/utils"
)

const ftimerLockKeyPrefix = "FTIMER_LOCK_PREFIX_"

type DistributeLocker interface {
	Lock(context.Context, int64) error
	ExpireLock(ctx context.Context, expireSeconds int64) error
}

// DistributeLock 分布式锁
type DistributeLock struct {
	key string
	// 线程ID_协程ID，防止错误删锁。
	// 存 redis 就是 value
	token  string
	client *Client
}

func NewReentrantDistributeLock(key string, client *Client) *DistributeLock {
	return &DistributeLock{
		key:    key,
		token:  utils.GetProcessAndGoroutineIDStr(),
		client: client,
	}
}

// Lock 加锁.
func (r *DistributeLock) Lock(ctx context.Context, expireSeconds int64) error {
	// key:app,value:线程id+协程id，Get 获取，有返回，没有返回nil
	res, err := r.client.Get(ctx, r.key)
	if err != nil && !errors.Is(err, redis.ErrNil) {
		return err
	}

	// 锁已存在，并且是自己的，此时是重入
	if res == r.token {
		return nil
	}

	// 创建分布式锁，setnx 如果没有则创建，有则不干任何事情
	reply, err := r.client.SetNX(ctx, r.getLockKey(), r.token, expireSeconds)
	if err != nil {
		return err
	}

	re, _ := reply.(int64)
	// 1 设置成功，0 没有设置成功（别人已经设置成功）
	if re != 1 {
		return errors.New("lock is acquired by others")
	}

	return nil
}

// ExpireLock 更新锁的过期时间，基于 lua 脚本实现操作原子性
func (r *DistributeLock) ExpireLock(ctx context.Context, expireSeconds int64) error {
	keysAndArgs := []interface{}{r.getLockKey(), r.token, expireSeconds}
	reply, err := r.client.Eval(ctx, LuaCheckAndExpireDistributionLock, 1, keysAndArgs)
	if err != nil {
		return err
	}

	if ret, _ := reply.(int64); ret != 1 {
		return errors.New("can not expire lock without ownership of lock")
	}

	return nil
}

func (r *DistributeLock) getLockKey() string {
	return ftimerLockKeyPrefix + r.key
}

func (c *Client) GetDistributionLock(key string) DistributeLocker {
	return NewReentrantDistributeLock(key, c)
}
