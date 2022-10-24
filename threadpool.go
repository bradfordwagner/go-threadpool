package threadpool

import (
	"sync"
	"time"
)

// ThreadPool - public interface for mocking
type ThreadPool interface {
	Start() (done <-chan struct{})
}

// ThreadPoolIntervalInterface - interface to implement Worker and Tick functions
type ThreadPoolIntervalInterface interface {
	CreateWorker(workerIndex int)
	OnTick()
}

// ThreadPoolInterface - interface to implement Worker Function
type ThreadPoolInterface interface {
	CreateWorker(workerIndex int)
}

type WorkerFunc func(workerIndex int)
type TickFunc func()

type threadPool struct {
	config     *Config
	wg         *sync.WaitGroup
	workerFunc WorkerFunc
}

// enforce interface
var _ ThreadPool = (*threadPool)(nil)

func New(workerFunction WorkerFunc, opts ...Option) ThreadPool {
	// initialize configs
	c := newDefaultConfig()
	for _, opt := range opts {
		opt(c)
	}

	return &threadPool{
		workerFunc: workerFunction,
		config:     c,
		wg:         new(sync.WaitGroup),
	}
}

func (t *threadPool) Start() <-chan struct{} {
	t.wg.Add(t.config.numWorkers)
	for i := 0; i < t.config.numWorkers; i++ {
		go t.startWorker(i)
	}

	// if we are using ticked functions
	var ticker *time.Ticker
	if t.config.tickFunction != nil {
		ticker = t.startTicker()
	}

	// allow continuation
	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
		if ticker != nil {
			ticker.Stop()
		}
	}()
	return done
}

func (t *threadPool) startTicker() (tick *time.Ticker) {
	tick = time.NewTicker(t.config.tick)
	go func() {
		for {
			_, ok := <-tick.C
			t.config.tickFunction()
			if !ok {
				return
			}
		}
	}()
	return
}

func (t *threadPool) startWorker(w int) {
	defer t.wg.Done()
	t.workerFunc(w)
}
