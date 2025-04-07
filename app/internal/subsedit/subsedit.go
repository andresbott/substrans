package subsedit

import (
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/asticode/go-astisub"
)

type Translator struct {
	subtitles *astisub.Subtitles
	logger    *slog.Logger
}

type slogWriter struct {
	logger *slog.Logger
}

func (w slogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimRight(string(p), "\n") // Remove trailing newline
	w.logger.Debug(msg)
	return len(p), nil
}

// New creates a new Translator instance
func New(filePath string, logger *slog.Logger) (*Translator, error) {
	// Set the global logger to use slog
	log.SetOutput(slogWriter{logger: logger})
	log.SetFlags(0) // Disable default log flags

	subtitles, err := astisub.OpenFile(filePath)
	if err != nil {
		logger.Debug(fmt.Sprintf("error loading subtitle file: %v", err))
		return nil, fmt.Errorf("error loading subtitle file: %v", err)
	}
	logger.Debug("Subtitle file loaded successfully")
	return &Translator{subtitles: subtitles, logger: logger}, nil
}

// GetTotalItems returns the total number of subtitle items
func (t *Translator) GetTotalItems() int {
	return len(t.subtitles.Items)
}

// GetNthItem returns the Nth subtitle item
func (t *Translator) GetNthItem(n int) (*astisub.Item, error) {
	if n < 0 || n >= len(t.subtitles.Items) {
		return nil, fmt.Errorf("index out of range")
	}
	return t.subtitles.Items[n], nil
}

type TextReplace func([]astisub.Item, astisub.Item, []astisub.Item) ([]astisub.Line, error)

// ReplaceLineWithCallback replaces a single line with the string value returned by the callback
// accepts two parameters: slices of previous and next lines of size constextSize
func (t *Translator) ReplaceLineWithCallback(index int, contextSize int, callback TextReplace) error {
	if index < 0 || index >= len(t.subtitles.Items) {
		return fmt.Errorf("index out of range")
	}

	// Collect previous items
	tempPrevItems := []astisub.Item{}
	if index > 0 { // Ensure there are previous items
		itemCount := 0
		for i := index - 1; i >= 0 && itemCount < contextSize; i-- {
			tempPrevItems = append(tempPrevItems, *t.subtitles.Items[i])
			itemCount++
		}
	}

	// Reverse the collected items to maintain original order
	for i, j := 0, len(tempPrevItems)-1; i < j; i, j = i+1, j-1 {
		tempPrevItems[i], tempPrevItems[j] = tempPrevItems[j], tempPrevItems[i]
	}

	prevItems := tempPrevItems

	// Collect next items
	nextItems := []astisub.Item{}
	if index < len(t.subtitles.Items)-1 { // Ensure there are next items
		itemCount := 0
		for i := index + 1; i < len(t.subtitles.Items) && itemCount < contextSize; i++ {
			nextItems = append(nextItems, *t.subtitles.Items[i])
			itemCount++
		}
	}

	newLines, err := callback(prevItems, DeepCopyItem(t.subtitles.Items[index]), nextItems)
	if err != nil {
		return err
	}

	if len(newLines) != len(t.subtitles.Items[index].Lines) {
		return fmt.Errorf("callback returned unexpected amount of lines")
	}

	for i, line := range t.subtitles.Items[index].Lines {
		if i < len(newLines) {
			for j := range line.Items {
				if j < len(newLines[i].Items) {
					line.Items[j].Text = newLines[i].Items[j].Text
				}
			}
		}
	}
	return nil
}

// IterateAndReplace processes each item and logs the progress
func (t *Translator) IterateAndReplace(contextSize int, callback TextReplace) error {
	totalItems := len(t.subtitles.Items)
	var totalDuration time.Duration

	for i := 0; i < totalItems; i++ {
		start := time.Now()

		err := t.ReplaceLineWithCallback(i, contextSize, callback)
		if err != nil {
			return fmt.Errorf("error processing item %d: %w", i, err)
		}

		duration := time.Since(start)
		totalDuration += duration

		// Calculate estimated remaining time
		averageDuration := totalDuration / time.Duration(i+1)
		estimatedRemaining := averageDuration * time.Duration(totalItems-i-1)

		t.logger.Info("Processing line",
			"line", i+1,
			"total", totalItems,
			"duration", duration,
			"remaining", estimatedRemaining,
		)
	}
	return nil
}

func (t *Translator) Write(p string) error {
	return t.subtitles.Write(p)
}

// DeepCopyItem creates a deep copy of an astisub.Item
func DeepCopyItem(item *astisub.Item) astisub.Item {
	itemCopy := astisub.Item{
		Lines: make([]astisub.Line, len(item.Lines)),
		// Copy other fields if necessary
	}

	for i, line := range item.Lines {
		newLine := astisub.Line{
			Items: make([]astisub.LineItem, len(line.Items)),
		}
		for j, item := range line.Items {
			newLine.Items[j] = astisub.LineItem{
				Text: item.Text,
			}
		}
		itemCopy.Lines[i] = newLine
	}

	return itemCopy
}
