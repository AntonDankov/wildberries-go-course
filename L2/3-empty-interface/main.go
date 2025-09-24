package main

import (
	"fmt"
	"os"
)

func Foo() error {
	// Typed nil
	// тип *os.Patherror, не nil
	var err *os.PathError = nil
	return err
}

func main() {
	err := Foo()
	fmt.Println(err)
	// Интерфейс в го хранит в себе тип и знаниче.
	// Untyped nil имеет в себе и тип и значение nil
	//
	fmt.Println(err == nil)
}
