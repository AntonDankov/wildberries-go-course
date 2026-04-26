package main

import (
	"log"
	"time"
)

type RequestLoggerMessage struct {
	Method string
	Path   string
	Time   time.Time
}

var loggerChan = make(chan RequestLoggerMessage, 5)

func loggerLooop(done chan bool) {

	var message RequestLoggerMessage
	for {
		select {
		case message = <-loggerChan:
			log.Printf("[%s] %s %v", message.Method, message.Path, message.Time)
		case <-done:
			return
		}
	}

}

func LogRequest(method string, path string, t time.Time) {
	request := RequestLoggerMessage{
		Method: method,
		Path:   path,
		Time:   t,
	}

	loggerChan <- request
}
