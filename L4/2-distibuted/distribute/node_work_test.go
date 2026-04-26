package distribute

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"testing"
	"widlberries-go-course/L4-2/cut"
)

func TestSplitWorkOnlyOneLine(t *testing.T) {
	// Given
	msg := WorkMessage{Text: "just one line\n"}
	workers := uint8(3)

	// When
	indexes := splitWork(msg, workers)

	// Then
	if len(indexes) != 2 {
		t.Fatalf("Expected slice to stop early with length 2, got %d", len(indexes))
	}

	if indexes[0] != 0 || indexes[1] != 14 {
		t.Errorf("Expected indexes [0, 14], got %v", indexes)
	}
}

func TestSplitWorkWithSplitOnlyForFirstWorker(t *testing.T) {
	// Given
	// Length 15. Target size is 7 for 2 workers.
	// but \n only after split, so work only for the first one
	msg := WorkMessage{Text: "line1\nlongtail\n"}
	workers := uint8(2)

	// When
	indexes := splitWork(msg, workers)

	// Then
	if len(indexes) != 2 {
		t.Fatalf("Expected 2 indexes, got %d", len(indexes))
	}
	if indexes[0] != 0 || indexes[1] != 15 {
		t.Errorf("Expected indexes [0, 15], got %v", indexes)
	}

}

func TestSplitWorkWithSplitForBothWorkers(t *testing.T) {
	// Given
	// Length 15. Target size is 7 for 2 workers, so it should split at 8 between first and second worker.
	msg := WorkMessage{Text: "012\n456\n890123\n"}
	workers := uint8(2)

	// When
	indexes := splitWork(msg, workers)

	log.Printf("nice indexes: %v", indexes)
	// Then
	if len(indexes) != 3 {
		t.Fatalf("Expected 3 indexes, got %d", len(indexes))
	}
	if indexes[0] != 0 || indexes[1] != 8 || indexes[2] != 15 {
		t.Errorf("Expected indexes [0, 8, 15], got %v", indexes)
	}
}

func TestSplitWorkNormalBatch(t *testing.T) {
	// Given
	msg := WorkMessage{Text: "0\n2\n4\n6\n8\n"} // Len 10
	workers := uint8(3)

	// When
	indexes := splitWork(msg, workers)

	// Then
	if len(indexes) != 4 {
		t.Fatalf("Expected 4 indexes, got %d", len(indexes))
	}

	if indexes[0] != 0 || indexes[1] != 4 || indexes[2] != 8 || indexes[3] != 10 {
		t.Errorf("Expected [0, 4, 8, 10], got %v", indexes)
	}
}

// func TestProcessWork(t *testing.T) {
// 	// Given
// 	msg, err := NewWorkMessageFromInput("..\\test_files\\test.txt -f 1,3 -s")
// 	if err != nil {
// 		t.Fatalf("Failed to create WorkMessage: %v", err)
// 	}
// 	// When
// 	result := processWork(msg)
// 	// Then
// 	t.Errorf("result : %s", result)
// }

func TestProcessWorkNoMatchingSeparator(t *testing.T) {
	// Given
	msg := WorkMessage{
		Text: "just one line\nno separator here\n",
		Flags: cut.Flags{
			FieldRanges: []cut.Range{{Start: 0, End: 1}},
			Delimeter:   "\t",
			Separated:   true,
		},
	}
	// When
	result := processWork(msg)
	// Then
	if result != "" {
		t.Errorf("Expected empty string with -s flag and no delimiter, got %q", result)
	}
}

// Generating input file for the actual run
func TestGenerateInputFile(t *testing.T) {
	targetFileSize := int64(1 * 1024 * 1024) // 1 MB
	noSepInterval := 1000                    // every 1000th line has no separator
	testInputFile := "../test_files/input_1mb.txt"
	if err := os.MkdirAll("testdata", 0755); err != nil {
		t.Fatalf("mkdir testdata: %v", err)
	}

	f, err := os.Create(testInputFile)
	if err != nil {
		t.Fatalf("create %s: %v", testInputFile, err)
	}
	defer f.Close()

	w := bufio.NewWriterSize(f, 1<<20) // 1 MB write buffer

	var written int64
	row := 1

	for written < targetFileSize {
		var line string
		if row%noSepInterval == 0 {
			line = generateNoSepLine(row)
		} else {
			line = generateLine(row)
		}
		n, err := fmt.Fprintln(w, line)
		if err != nil {
			t.Fatalf("write row %d: %v", row, err)
		}
		written += int64(n)
		row++
	}

	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	info, _ := f.Stat()
	t.Logf("Generated %s: %.2f MB, %d rows", testInputFile,
		float64(info.Size())/1024/1024, row-1)
}

func generateLine(row int) string {
	return fmt.Sprintf(
		"row_%08d\tcol2_%08d\tcol3_%08d\tcol4_%08d\tcol5_%08d",
		row, row, row, row, row,
	)
}

// generateNoSepLine returns a line without any delimiter (for -s flag testing).
func generateNoSepLine(row int) string {
	return fmt.Sprintf("NOSEP_%08d", row)
}
