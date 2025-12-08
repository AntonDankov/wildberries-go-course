## Flags

- `-f "fields"` - Specify field numbers to output (required)
  - Single fields: `-f 1` or `-f 1,3,5`
  - Ranges: `-f 1-3` or `-f 1-3,5-7`
  - Mixed: `-f 1,3-5,7`
- `-d "delimiter"` - Set field delimiter (default: tab `\t`)
- `-s` - Only output lines containing the delimiter

## How to run

Basic usage:

```bash
<input> | go run . -f 1-3
```

Example:

```bash
cat .\test_files\test.txt | go run . -f 1,3 -s
```
