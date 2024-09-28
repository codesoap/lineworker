lineworker provides worker pools that perform work in parallel, but
output the work results in the order the work was given.

Take a look at the documentation for more info: https://godocs.io/github.com/codesoap/lineworker

# Example
```go
slowSprint := func(a int) (string, error) {
	delay := rand.Int()
	time.Sleep(time.Duration(delay%6) * time.Millisecond)
	return fmt.Sprint(a), nil
}

// Start the worker goroutines:
pool := lineworker.NewWorkerPool(runtime.NumCPU(), slowSprint)

// Put in work:
go func() {
	for i := 0; i < 10; i++ {
		workAccepted := pool.Process(i)
		if !workAccepted {
			// Cannot happen in this example, because pool.Stop is not called
			// outside this goroutine, but is handled for demonstration
			// purposes.
			return
		}
	}
	pool.Stop()
}()

// Retrieve the results:
for {
	res, err := pool.Next()
	if err == lineworker.EOS {
		break
	} else if err != nil {
		panic(err)
	}
	fmt.Println(res)
}
```
