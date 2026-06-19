package main

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func test() *customError {
	// ... do something
	return nil
}

// Выводит error
// а для проверки == nil, он также должен иметь тип nil
// Проблема в топ что мы присваваем наш поинтер к интерфейсу error и получаем уже boxed значение c nil, вместо правильного nil
// проверить корректнее было бы принять также поинтер и его проверить на nil, и если он не нил приравнять к error
func main() {
	var err error
	err = test()
	if err != nil {
		println("error")
		return
	}
	println("ok")
}
