## Usage

Run the program directly using `go run .` and provide the required flags.
The downloaded files will be saved in a `scraped/` directory created automatically in the project's root folder.

### Flags

- `-u`: **(Required)** Start URL for crawling.
- `-d`: The recursion depth (default: 1). Minimum value is 1.
- `-N`: The number of concurrent worker threads (default: 1). Minimum value is 1.

### How to run

Example:

```bash
go run . -u "https://books.toscrape.com/catalogue/a-light-in-the-attic_1000/index.html" -d 3 -N 5
