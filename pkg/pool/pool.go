package pool

import (
	"github.com/panjf2000/ants/v2"
)

type WorkerPool interface {
	Submit(func()) error
}

type GoWorkerPool struct {
	pool *ants.Pool
}

func NewGoWorkerPool(size int) *GoWorkerPool {
	pool, err := ants.NewPool(size)
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
