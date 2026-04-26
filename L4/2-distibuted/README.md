## Run code
go run .
After hello message wait for leader message or start of the election
After write the command in the input line:

### Input
# <input_file> -f <fields> [options]

A command-line tool for extracting selected fields from each line of a file, similar to the Unix `cut` utility.

## Usage

Build the program using command
```bash
go build -o app.exe
```

Run one or many of the copies of program in the terminal
```bash
.\app.exe
```
Then in the input run the commans
```bash
 <input_file> -f <fields> [options]
```
### Example of the command
Example file can be find in test_files/test.txt
#### Run Example

.\test_files\test.txt -f 1,3 -s -o .\test_files\result.txt

#### Generate test file
go test ./distributed/... -run TestGenerateInputFile -v

## Flags

| Flag | Default | Description |
|---|---|---|
| `-f` | *(required)* | Fields to extract. Accepts comma-separated field numbers and/or ranges (e.g. `1,3` or `1-3` or `1,3-5`) |
| `-d` | `\t` | Field delimiter character |
| `-s` | `false` | Only print lines that contain the delimiter (skip lines without it) |
| `-o` | *(stdout)* | Path to output file. If not specified, output is printed to the console. Can be used ONLY be leader node! |

## Task
Можно выбрать одну из утилит или реализовать несколько. Этот проект тренирует умение работать со строками, файлами, параметрами командной строки и приближает к разработке реальных инструментов. Реализуйте распределённую версию одной из классических CLI-утилит (grep, cut или sort) с поддержкой работы в многосерверном режиме.

Результат: исполняемый файл (например, mygrep), принимающий флаги, читающий входной поток и выводящий результат. 

Каждый экземпляр сервиса обрабатывает часть данных (например, файл или поток), передаёт/получает задания через сеть и обменивается результатами с остальными.

Итоговый результат вычисляется при достижении кворума (например, N/2+1 серверов успешно обработали свой фрагмент). Используйте каналы и горутины для параллелизма внутри каждого сервера и для обмена между ними. 

Обязательно должны быть подготовлены примеры использования и сравнительный тест с оригинальной утилитой.