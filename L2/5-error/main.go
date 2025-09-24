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
// как и в случае с 3 задачей передается не Untyped nil
// а typed nil, интерфейс с типом *customerror
// а для проверки == nil, он также должен иметь тип nil
func main() {
	var err error
	err = test()
	if err != nil {
		println("error")
		return
	}
	println("ok")
}
