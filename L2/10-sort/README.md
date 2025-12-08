## Flags

- `-k` - Sort by column N (1-based, tab separated, default: 1)
- `-n` - Sort numerically
- `-r` - Reverse order
- `-u` - Unique lines only
- `-M` - Sort by month name
- `-b` - Ignore trailing blanks
- `-c` - Check if already sorted
- `-h` - Human-readable number sort with suffixes (e.g., 1K, 2M)

## How to Run

Basic usage:

```bash
go run <flags> . <input-file>
```

Test command:

```
go run . -Mk 3 .\test_files\test1.txt
```
