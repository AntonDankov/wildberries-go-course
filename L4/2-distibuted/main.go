package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"widlberries-go-course/L4-2/distribute"

	"github.com/chzyer/readline"
)

func main() {

	nodeState := distribute.SetNode()

	rl, _ := readline.New("> ")
	defer rl.Close()
	log.SetOutput(rl.Stdout())

	go func() {
		for {
			line, err := rl.Readline()
			if err != nil {
				log.Fatalf("input failed with error: %v", err)
			}
			if len(line) == 0 {
				continue
			}
			workMessage, err := distribute.NewWorkMessageFromInput(line)
			if err != nil {
				log.Printf("Failed to work with input, error: %v", err)
				continue
			}
			nodeState.WorkChan <- workMessage
		}
	}()

	addr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:9999")
	if err != nil {
		log.Fatal(err)
	}

	writeMulticastConnection, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	listeningMulticastConnection, err := distribute.CreateListeningUDPBroadcastConnection(9999)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancelContext := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		distribute.OperateBroadcastState(listeningMulticastConnection, writeMulticastConnection, ctx)
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	cancelContext()
	wg.Wait()
	log.Printf("shutdown complete")

}
