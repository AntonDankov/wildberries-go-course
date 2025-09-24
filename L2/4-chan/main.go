package main

func main() {
	ch := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
		}
		// Выведет deadlock, потому что канал не был закрыт и будет бесконечное ожидание
		// нужно передачи всех значений в канал, стоит закрыть его
		// close(ch)
	}()
	for n := range ch {
		println(n)
	}
}
