package main

import (
	"sync"
)

type WorkerPool struct {
	w         sync.WaitGroup
	freeNum   int
	workerNum int
	lock      sync.Mutex
}

func (pool *WorkerPool) Init(worker func(id int), workerNum int) {
	for id := 0; id < threadNum; id++ {
		go worker(id)
		pool.w.Add(1)
	}
	pool.workerNum = workerNum
	pool.freeNum = workerNum
}

func (pool *WorkerPool) Wait() {
	pool.w.Wait()
}

func (pool *WorkerPool) Close() {
	for i := 0; i < pool.workerNum; i++ {
		pool.w.Done()
	}
}

func (pool *WorkerPool) FreePop() {
	pool.lock.Lock()
	pool.freeNum--
	pool.lock.Unlock()
}

func (pool *WorkerPool) FreeAdd() {
	pool.lock.Lock()
	pool.freeNum++
	pool.lock.Unlock()
}
