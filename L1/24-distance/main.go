package main

import (
	"fmt"
	"math"
)

type Point struct {
	x float64
	y float64
}

func NewPoint(x float64, y float64) Point {
	return Point{
		x: x,
		y: y,
	}
}

func (point *Point) calculateDistance(other Point) float64 {
	dx := math.Pow(point.x-other.x, 2)
	dy := math.Pow(point.y-other.y, 2)

	return math.Sqrt(dx + dy)
}

func main() {
	point1 := NewPoint(3, 5)
	point2 := NewPoint(1, 10)

	distance := point1.calculateDistance(point2)
	fmt.Printf("distance: %.1f\n", distance)
}
