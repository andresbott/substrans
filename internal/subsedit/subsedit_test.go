package subsedit

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/asticode/go-astisub"
	"github.com/google/go-cmp/cmp"
)

func TestLoadAndReadSubtitles(t *testing.T) {
	filePath := "testData/overlord.ass"

	// Create a new Editor instance
	translator, err := New(filePath, silentLogger())
	if err != nil {
		t.Fatalf("Failed to create Editor: %v", err)
	}

	// Get the first subtitle item
	firstItem, err := translator.GetNthItem(0)
	if err != nil {
		t.Fatalf("Failed to get the first subtitle item: %v", err)
	}

	expectedText := "The Roble Sacred Kingdom, lying on a peninsula"

	// Compare the actual text with the expected text using cmp.Diff
	if len(firstItem.Lines) > 0 && len(firstItem.Lines[0].Items) > 0 {
		actualText := firstItem.Lines[0].Items[0].Text
		if diff := cmp.Diff(expectedText, actualText); diff != "" {
			t.Errorf("Mismatch (-expected +actual):\n%s", diff)
		}
	} else {
		t.Error("The first subtitle item does not contain any text lines")
	}
}

func TestReplaceLineWithCallback(t *testing.T) {
	tcs := []struct {
		name          string
		filePath      string
		lineNumber    int
		constextSize  int
		expectedText  []astisub.Line
		previousItems []astisub.Item
		actualItem    astisub.Item
		nextItems     []astisub.Item
	}{
		{
			name:         "Replace the first line of an .ass file",
			filePath:     "testData/overlord.ass",
			lineNumber:   0,
			constextSize: 3,
			expectedText: []astisub.Line{
				{Items: []astisub.LineItem{{Text: "[[The Roble Sacred Kingdom, lying on a peninsula]]"}}},
				{Items: []astisub.LineItem{{Text: "[[to the southwest of the Re-Estize Kingdom.]]"}}},
			},
			previousItems: []astisub.Item{}, // expect empty here

			actualItem: astisub.Item{
				Lines: []astisub.Line{
					{Items: []astisub.LineItem{{Text: "The Roble Sacred Kingdom, lying on a peninsula"}}},
					{Items: []astisub.LineItem{{Text: "to the southwest of the Re-Estize Kingdom."}}},
				},
			},

			nextItems: []astisub.Item{
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "Its territory is divided into northern"}}},
						{Items: []astisub.LineItem{{Text: "and southern halves by a great gulf."}}},
					},
				},
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "The borders of the Sacred Kingdom are guarded by a massive wall"}}},
						{Items: []astisub.LineItem{{Text: "to protect against demihuman tribes in the bordering Abelion Hills."}}},
					},
				},
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "An area also located near the Slain Theocracy."}}},
					},
				},
			},
		},

		{
			name:         "Replace the second line of an .ass file",
			filePath:     "testData/overlord.ass",
			lineNumber:   1,
			constextSize: 3,
			expectedText: []astisub.Line{
				{Items: []astisub.LineItem{{Text: "[[Its territory is divided into northern]]"}}},
				{Items: []astisub.LineItem{{Text: "[[and southern halves by a great gulf.]]"}}},
			},
			previousItems: []astisub.Item{
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "The Roble Sacred Kingdom, lying on a peninsula"}}},
						{Items: []astisub.LineItem{{Text: "to the southwest of the Re-Estize Kingdom."}}},
					},
				},
			},

			actualItem: astisub.Item{
				Lines: []astisub.Line{
					{Items: []astisub.LineItem{{Text: "Its territory is divided into northern"}}},
					{Items: []astisub.LineItem{{Text: "and southern halves by a great gulf."}}},
				},
			},

			nextItems: []astisub.Item{
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "The borders of the Sacred Kingdom are guarded by a massive wall"}}},
						{Items: []astisub.LineItem{{Text: "to protect against demihuman tribes in the bordering Abelion Hills."}}},
					},
				},
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "An area also located near the Slain Theocracy."}}},
					},
				},
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "It's a symbol of strength as well as the suffering"}}},
						{Items: []astisub.LineItem{{Text: "they face at the hands of their neighbors."}}},
					},
				},
			},
		},
		{
			name:         "Replace the one to last line of an .ass file",
			filePath:     "testData/overlord.ass",
			lineNumber:   2006,
			constextSize: 3,
			expectedText: []astisub.Line{
				{Items: []astisub.LineItem{{Text: "[[Enjoy your happiness while you can, my dear subjects.]]"}}},
			},
			previousItems: []astisub.Item{
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "Very well, my lord."}}},
					},
				},
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "Good then."}}},
					},
				},
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "I'll leave you to it."}}},
					},
				},
			},

			actualItem: astisub.Item{
				Lines: []astisub.Line{
					{Items: []astisub.LineItem{{Text: "Enjoy your happiness while you can, my dear subjects."}}},
				},
			},
			nextItems: []astisub.Item{
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "You may find it is quite fleeting."}}},
					},
				},
			},
		},
		{
			name:         "Replace line with special position",
			filePath:     "testData/withPos.ass",
			lineNumber:   1,
			constextSize: 1,
			expectedText: []astisub.Line{
				{Items: []astisub.LineItem{{Text: "[[]]"}, {Text: "[[Syr]]"}}},
			},
			previousItems: []astisub.Item{
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "I need to validate myself. And prove who I want to be."}}},
					},
				},
			},
			actualItem: astisub.Item{
				Lines: []astisub.Line{
					{Items: []astisub.LineItem{{Text: ""}, {Text: "Syr"}}},
				},
			},
			nextItems: []astisub.Item{
				{
					Lines: []astisub.Line{
						{Items: []astisub.LineItem{{Text: "Congratulations, Little Miss Supporter!"}}},
					},
				},
			},
		},
	}

	ignoreItemProps := cmpopts.IgnoreFields(astisub.Item{}, "Comments", "Index", "EndAt", "InlineStyle", "Region", "StartAt", "Style", "InlineStyle")
	ignoreLineProps := cmpopts.IgnoreFields(astisub.LineItem{}, "InlineStyle")

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new Editor instance
			translator, err := New(tc.filePath, silentLogger())
			if err != nil {
				t.Fatalf("Failed to create Editor: %v", err)
			}

			// Capture the parameters passed to the callback
			var capturedPrevItems []astisub.Item
			var capturedActualItem astisub.Item
			var capturedNextItems []astisub.Item

			callback := func(prevItems []astisub.Item, actualItem astisub.Item, nextItems []astisub.Item) ([]astisub.Line, error) {
				capturedPrevItems = prevItems
				capturedActualItem = actualItem
				capturedNextItems = nextItems
				out := []astisub.Line{}
				for _, line := range actualItem.Lines {
					newLine := astisub.Line{Items: []astisub.LineItem{}}
					for _, item := range line.Items {
						newLine.Items = append(newLine.Items, astisub.LineItem{Text: "[[" + item.Text + "]]"})
					}
					out = append(out, newLine)
				}
				return out, nil
			}

			// Replace the line using the callback
			err = translator.ReplaceLineWithCallback(tc.lineNumber, tc.constextSize, callback)
			if err != nil {
				t.Fatalf("Failed to replace line: %v", err)
			}

			// Verify the captured parameters
			if diff := cmp.Diff(tc.previousItems, capturedPrevItems, ignoreItemProps); diff != "" {
				t.Errorf("Mismatch in prevItems (-expected +actual):\n%s", diff)
			}

			if diff := cmp.Diff(tc.actualItem, capturedActualItem, ignoreItemProps); diff != "" {
				t.Errorf("Mismatch in actualItem (-expected +actual):\n%s", diff)
			}

			if diff := cmp.Diff(tc.nextItems, capturedNextItems, ignoreItemProps); diff != "" {
				t.Errorf("Mismatch in nextItems (-expected +actual):\n%s", diff)
			}

			// Get the first subtitle item
			got, err := translator.GetNthItem(tc.lineNumber)
			if err != nil {
				t.Fatalf("Failed to get the first subtitle item: %v", err)
			}

			// Check if the line was replaced correctly
			if diff := cmp.Diff(tc.expectedText, got.Lines, ignoreLineProps); diff != "" {
				t.Errorf("Mismatch (-expected +actual):\n%s", diff)
			}

		})
	}
}

func TestIterateAndReplace(t *testing.T) {
	filePath := "testData/overlord.ass"

	// Create a new Editor instance
	translator, err := New(filePath, silentLogger())
	if err != nil {
		t.Fatalf("Failed to create Editor: %v", err)
	}

	// Define the callback function
	callback := func(prevItems []astisub.Item, actualItem astisub.Item, nextItems []astisub.Item) ([]astisub.Line, error) {
		out := []astisub.Line{}
		for _, line := range actualItem.Lines {
			newLine := astisub.Line{Items: []astisub.LineItem{}}
			for _, item := range line.Items {
				newLine.Items = append(newLine.Items, astisub.LineItem{Text: "[[" + item.Text + "]]"})
			}
			out = append(out, newLine)
		}
		return out, nil
	}

	// Call IterateAndReplace with context size 1
	err = translator.IterateAndReplace(1, callback)
	if err != nil {
		t.Fatalf("Failed to iterate and replace: %v", err)
	}

	// Write the modified subtitles to a buffer
	var buf strings.Builder
	err = translator.subtitles.WriteToSSA(&buf)
	if err != nil {
		t.Fatalf("Failed to write subtitles to buffer: %v", err)
	}
	//translator.Write("testData/overlord_modified.ass")

	modifiedFile := "testData/overlord_modified.ass"
	originalContent, err := os.ReadFile(modifiedFile)
	if err != nil {
		t.Fatalf("Failed to read original file: %v", err)
	}

	// Compare the buffer content with the original file content
	if diff := cmp.Diff(string(originalContent), buf.String()); diff != "" {
		t.Errorf("Mismatch (-expected +actual):\n%s", diff)
	}
}

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}
