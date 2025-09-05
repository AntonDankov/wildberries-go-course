package main

import (
	"context"
	"fmt"
	"time"
)

func Sleep(duration time.Duration) {
	context, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	<-context.Done()
}

func main() {
	fmt.Println("Started")
	Sleep(5 * time.Second)
	fmt.Println("Finished")
}
