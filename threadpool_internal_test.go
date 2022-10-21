package threadpool

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/atomic"
	"sync"
	"time"
)

var _ = Describe("Threadpool", func() {

	It("all workers started", func() {
		numWorkers := 5
		counter := atomic.NewInt32(0)

		// create thread pool
		tp := New(func(w int) {
			// increment the counter to check the number of workers started
			counter.Inc()
		}, OptionWorkerRoutines(numWorkers))

		// start and await completion
		<-tp.Start()

		Expect(counter.Load()).To(Equal(int32(numWorkers)))
	})

	It("uses unbuffered channel for work", func() {
		workItems, workRoutines, work := 20, 5, make(chan int)
		counter := atomic.NewInt32(0)

		// create thread pool with set number of routines
		tp := New(func(w int) {
			for range work {
				counter.Inc()
			}
		}, OptionWorkerRoutines(workRoutines))
		done := tp.Start()

		// enter work items for worker threads
		for i := 0; i < workItems; i++ {
			work <- i
		}
		close(work)
		<-done

		Expect(counter.Load()).To(Equal(int32(workItems)))
	})

	It("uses buffered channel with delays", func() {
		workItems, workRoutines, work := 20, 5, make(chan int, 20)
		counter := atomic.NewInt32(0)

		// create thread pool with set number of routines
		tp := New(func(w int) {
			for range work {
				counter.Inc()
				time.Sleep(time.Millisecond * 17)
			}
		}, OptionWorkerRoutines(workRoutines))
		done := tp.Start()

		// enter work items for worker threads
		for i := 0; i < workItems; i++ {
			work <- i
		}
		close(work)
		<-done

		// all work has completed successfully
		Expect(counter.Load()).To(Equal(int32(workItems)))
	})

	It("uses buffered channel with delays, and progress function", func() {
		workItems, workRoutines, work := 20, 5, make(chan int, 20)
		completedTasks := atomic.NewInt32(0)
		done := make(chan struct{})

		// create thread pool with set number of routines
		tp := New(func(w int) {
			for range work {
				time.Sleep(time.Millisecond * 17)
				completedTasks.Inc()
			}
		}, OptionWorkerRoutines(workRoutines), OptionTickFunction(func() {
			// use progress as the end condition
			// when everything is completed on a tick it will stop the test
			progress := int(completedTasks.Load())
			if progress == workItems {
				close(done)
			}
		}), OptionTick(time.Millisecond*5))
		tp.Start()

		// enter work items for worker threads
		for i := 0; i < workItems; i++ {
			work <- i
		}

		// wait for ticker to indicate we have seen all work items
		// then kill the work channel
		<-done
		close(work)
		Expect(completedTasks.Load()).To(Equal(int32(workItems)))
	})

	It("has two routines listening for completion", func() {
		work := make(chan int)
		wg := new(sync.WaitGroup)
		wg.Add(2)

		tp := New(func(workerIndex int) {
			for range work {
				//do nothing
			}
		}, OptionWorkerRoutines(5))
		done := tp.Start()

		// setup two listeners
		f := func() {
			select {
			case _, ok := <-done:
				if !ok {
					wg.Done()
				}
			}
		}
		go f()
		go f()

		for i := 0; i < 20; i++ {
			work <- i
		}
		close(work)

		// completed
		<-done
		wg.Wait()
	})
})
