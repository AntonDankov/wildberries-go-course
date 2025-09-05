package main

import "fmt"

func setBitToZero(number *int64, index int) {
	*number = *number & (^(1 << index))
}

func setBitToOne(number *int64, index int) {
	*number = *number | (1 << index)
}

func main() {
	number := int64(5)
	setBitToZero(&number, 0)
	fmt.Printf("%d = 4\n", number)

	setBitToOne(&number, 1)

	fmt.Printf("%d = 6\n", number)
}
