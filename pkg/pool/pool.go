package pool

import (
	"github.com/panjf2000/ants/v2"
	"time"
	"timer/common/conf"
)

type WorkerPool interface {
	Submit(func()) error
}

type GoWorkerPool struct {
	pool *ants.Pool
}

func NewGoWorkerPool(config conf.PoolConfig) *GoWorkerPool {
	pool, err := ants.NewPool(config.Size, ants.WithExpiryDuration(time.Duration(config.ExpireSeconds)*time.Second))
	if err != nil {
		panic(err)
	}
	return &GoWorkerPool{
		pool: pool,
	}
}

func (gPool *GoWorkerPool) Submit(f func()) error {
	return gPool.pool.Submit(f)
}