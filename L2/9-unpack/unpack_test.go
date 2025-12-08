package main

import (
	"errors"
	"testing"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
		expectedError  error
	}{
		{
			name:           "With repeats",
			input:          "a4bc2d5e",
			expectedOutput: "aaaabccddddde",
			expectedError:  nil,
		},
		{
			name:           "No numbers",
			input:          "abcd",
			expectedOutput: "abcd",
			expectedError:  nil,
		},
		{
			name:           "Only digits",
			input:          "45",
			expectedOutput: "",
			expectedError:  errors.New("no character was provided to unpack"),
		},
		{
			name:           "Empty string",
			input:          "",
			expectedOutput: "",
			expectedError:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := unpack(test.input)

			if test.expectedError != nil {
				if err == nil || err.Error() != test.expectedError.Error() {
					t.Errorf("unpack(%q) error = %v, want %v", test.input, err, test.expectedError)
				}
				if result != test.expectedOutput {
					t.Errorf("unpack(%q) output = %q, want %q", test.input, result, test.expectedOutput)
				}
				return
			}

			if err != nil {
				t.Errorf("unpack(%q) unexpected error: %v", test.input, err)
			}

			if result != test.expectedOutput {
				t.Errorf("unpack(%q) = %q, want %q", test.input, result, test.expectedOutput)
			}
		})
	}
}
