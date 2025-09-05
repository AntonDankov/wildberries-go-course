package main

type Human struct {
	Health int
}

func (human *Human) kill() {
	human.Health = 0
}

type Action struct {
	Human
}

func main() {
}
