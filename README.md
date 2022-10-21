# go-threadpool

## Import
```bash
go get github.com/bradfordwagner/go-threadpool
```

## Examples
### Simple Worker Pool
```go
package main

import (
	"fmt"
	"github.com/bradfordwagner/go-threadpool"
)

func main() {
	work := make(chan int)
	tp := threadpool.New(func(workerIndex int) {
		for i := range work {
			fmt.Printf("worker=%d, i=%d\n", workerIndex, i)
		}
	}, threadpool.OptionWorkerRoutines(5))
	done := tp.Start()

	// setup work to be done
	for i := 0; i < 10; i++ {
		work <- i
	}
	close(work)
	<-done
}
```

### Worker Pool With Progress
```go
package main

import (
	"fmt"
	"github.com/bradfordwagner/go-threadpool"
	"github.com/dustin/go-humanize"
	"go.uber.org/atomic"
	"time"
)

func main() {
	work := make(chan int)
	workItems, progress := 100, atomic.NewInt32(0)
	tp := threadpool.New(func(workerIndex int) {
		for range work {
			progress.Inc()
			time.Sleep(time.Millisecond * 500)
		}
	},
		threadpool.OptionWorkerRoutines(5),
		threadpool.OptionTickFunction(func() {
			progress := humanize.FormatFloat("###.##", float64(progress.Load())/float64(workItems))
			fmt.Printf("progress=%s\n", progress)
		}), threadpool.OptionTick(time.Second))
	done := tp.Start()

	// setup work to be done
	for i := 0; i < workItems; i++ {
		work <- i
	}
	close(work)
	<-done
	fmt.Println("Completed")
}
```
#### Results
```
❯ go run .
progress=0.10
progress=0.20
progress=0.30
progress=0.40
progress=0.50
progress=0.60
progress=0.70
progress=0.80
progress=0.90
progress=1.00
Completed
```
