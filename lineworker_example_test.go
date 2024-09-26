package lineworker_test

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/codesoap/lineworker"
)

func ExampleWorkerPool() {
	slowSprint := func(a int) (string, error) {
		delay := rand.Int()
		time.Sleep(time.Duration(delay%6) * time.Millisecond)
		return fmt.Sprint(a), nil
	}
	pool := lineworker.NewWorkerPool(runtime.NumCPU(), slowSprint)
	go func() {
		for i := 0; i < 10; i++ {
			pool.Process(i)
		}
		pool.Stop()
	}()
	for {
		res, err := pool.Next()
		if err == lineworker.EOS {
			break
		} else if err != nil {
			panic(err)
		}
		fmt.Println(res)
	}
	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
	// 7
	// 8
	// 9
}
