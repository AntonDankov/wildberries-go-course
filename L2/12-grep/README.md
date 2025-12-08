## Flags

- `-A N` - Print N lines after each match
- `-B N` - Print N lines before each match
- `-C N` - Print N lines of context (both before and after)
- `-c` - Print only the count of matching lines
- `-i` - Ignore case
- `-v` - Invert match (print lines that do NOT match)
- `-F` - Fixed string match (exact substring, not regex)
- `-n` - Print line numbers

## How to Run

Basic run:

```bash
go run . <pattern> <*file>
```

Example:

```bash
go run . "test" .\test_files\test.txt
```
