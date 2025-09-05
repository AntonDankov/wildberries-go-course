package main

import (
	"fmt"
	"math"
	"math/big"
)

func addBig(a int64, b int64) string {
	bigA := big.NewInt(a)
	bigB := big.NewInt(b)
	result := big.NewInt(0).Add(bigA, bigB)
	return result.String()
}

func subBig(a int64, b int64) string {
	bigA := big.NewInt(a)
	bigB := big.NewInt(b)
	result := big.NewInt(0).Sub(bigA, bigB)
	return result.String()
}

func mulBig(a int64, b int64) string {
	bigA := big.NewInt(a)
	bigB := big.NewInt(b)
	result := big.NewInt(0).Mul(bigA, bigB)
	return result.String()
}

func divBig(a int64, b int64) string {
	bigA := big.NewInt(a)
	bigB := big.NewInt(b)
	result := big.NewInt(0).Div(bigA, bigB)
	return result.String()
}

func main() {
	a := int64(math.MaxInt64)

	fmt.Println(addBig(a, 2))
	fmt.Println(subBig(-1, a))
	fmt.Println(mulBig(a, 2))
	fmt.Println(divBig(a, 3))
}
