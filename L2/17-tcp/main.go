package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"golang.org/x/term"
)

func main() {
	flags := GetFlags()

	address := net.JoinHostPort(flags.host, flags.port)
	timeout := time.Duration(flags.timeout) * time.Second

	connection, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	fmt.Printf("Connected to %v\n", connection.RemoteAddr())

	doneChan := make(chan struct{})

	go func() {
		_, err := io.Copy(os.Stdout, connection)
		if err != nil {
			os.Exit(0)
		}
		os.Exit(0)
	}()

	go func() {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		var line []byte
		buf := make([]byte, 1)
		writer := bufio.NewWriter(connection)

		for {
			select {
			default:
				n, err := os.Stdin.Read(buf)
				if err != nil {
					fmt.Println("\nFailed to read input")
					return
				}
				if n == 0 {
					continue
				}
				switch buf[0] {
				case 0x04:
					close(doneChan)
					return
				case 0x03:
					close(doneChan)
					return

				case '\r', '\n':
					fmt.Print("\r\n")
					if len(line) > 0 {
						writer.Write(line)
					}
					writer.WriteString("\r\n")
					writer.Flush()
					line = nil

				case 0x7F, 0x08:
					if len(line) > 0 {
						line = line[:len(line)-1]
						fmt.Print("\b \b")
					}
				default:
					if buf[0] >= 32 && buf[0] < 127 {
						line = append(line, buf[0])
						fmt.Printf("%c", buf[0])
					}

				}
			}
		}
	}()

	<-doneChan
}
