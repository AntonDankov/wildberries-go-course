package main

import "fmt"

// defer функция выполняется после выхода из функции, где она была определена,
// но после того как функция исполнится и return значение просчитано
// в этом случае в return не передается x, а делается через named return value
func test() (x int) {
	defer func() {
		x++
	}()
	x = 1
	return
}

// В этом тесте возвращаемое значение определено и оно будет помещено в регистр перед исполнением defer
func anotherTest() int {
	var x int
	defer func() {
		x++
	}()
	x = 1
	return x
}

func main() {
	fmt.Println(test())
	fmt.Println(anotherTest())
}
