package main

import "fmt"

type TypeString string

const (
	String  TypeString = "string"
	Int     TypeString = "int"
	Bool    TypeString = "bool"
	Channel TypeString = "channel"
	Unknown TypeString = "unknown"
)

func getType(t interface{}) TypeString {
	switch t.(type) {
	case int:
		return Int
	case string:
		return String
	case bool:
		return Bool
	case chan bool, chan int, chan string:
		return Channel
	default:
		return Unknown
	}
}

func main() {
	values := []any{
		0,
		"test",
		false,
		make(chan string),
		3.14,
		nil,
	}
	for i, value := range values {
		t := getType(value)
		fmt.Printf("%d: %v %s\n", i, value, t)

	}
}
