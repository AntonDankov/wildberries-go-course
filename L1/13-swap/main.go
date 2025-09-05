package main

import "fmt"

func swap(a *int, b *int) {
	*a = *a + *b
	*b = *a - *b
	*a = *a - *b
}

func main() {
	a := 8
	b := 4

	fmt.Printf("%d | %d\n", a, b)
	swap(&a, &b)
	fmt.Printf("%d | %d\n", a, b)
}
