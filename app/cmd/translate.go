package cmd

import (
	"context"
	"fmt"
	"github.com/andresbott/substrans/app/logger"
	"log/slog"
	"os"

	"github.com/andresbott/substrans/internal/llmtranslate"
	"github.com/andresbott/substrans/internal/subsedit"
	"github.com/asticode/go-astisub"
	"github.com/spf13/cobra"
)

func translateCmd() *cobra.Command {
	var inputFile string
	var outputFile string
	var targetLanguage string
	var model string

	cmd := &cobra.Command{
		Use:   "translate",
		Short: "Translate subtitles to another language",
		Long:  `Translate a video subtitle file to another language using the specified target language.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputFile == "" || targetLanguage == "" {
				return fmt.Errorf("input file and target language must be specified")

			}
			fmt.Printf("Translating %s to %s and saving to %s\n", inputFile, targetLanguage, outputFile)

			ollamaURL := os.Getenv("OLLAMA_HOST")

			if model == "" {
				model = llmtranslate.ModelLlama31
			}

			translator, err := llmtranslate.NewTranslator(model, ollamaURL, 0.3)
			if err != nil {
				return fmt.Errorf("failed to create translator: %v", err)

			}

			log, err := logger.GetDefault(slog.LevelInfo)
			if err != nil {
				return fmt.Errorf("failed to create logger: %v", err)
			}

			editor, err := subsedit.New(inputFile, log)
			if err != nil {
				return fmt.Errorf("failed to create subtitle editor: %v", err)
			}

			callback := func(prevItems []astisub.Item, actualItem astisub.Item, nextItems []astisub.Item) ([]astisub.Line, error) {
				return translateCallback(prevItems, actualItem, nextItems, translator, targetLanguage)
			}

			err = editor.IterateAndReplace(10, callback)
			if err != nil {
				return fmt.Errorf("failed to translate subtitles %v", err)
			}

			err = editor.Write(outputFile)
			if err != nil {
				return fmt.Errorf("failed to save translated subtitles: %v", err)
			}

			fmt.Println("Translation completed successfully.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input subtitle file")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output subtitle file")
	cmd.Flags().StringVarP(&targetLanguage, "language", "l", "", "Target language for translation")
	cmd.Flags().StringVarP(&model, "model", "m", "", "model to use")

	return cmd
}

func translateCallback(prevItems []astisub.Item, actualItem astisub.Item, nextItems []astisub.Item, translator *llmtranslate.Translator, targetLanguage string) ([]astisub.Line, error) {
	ctx := context.Background()
	prevContext := extractText(prevItems)
	postContext := extractText(nextItems)
	var translatedLines []astisub.Line

	for _, line := range actualItem.Lines {

		newLine := astisub.Line{}
		for _, item := range line.Items {
			// empty text lines might happen to add style to a line
			if item.Text == "" {
				newLine.Items = append(newLine.Items, astisub.LineItem{Text: ""})
				continue
			}
			translatedText, err := translator.Translate(ctx, prevContext, postContext, item.Text, targetLanguage)
			if err != nil {
				return nil, err
			}
			newLine.Items = append(newLine.Items, astisub.LineItem{Text: translatedText})
		}
		translatedLines = append(translatedLines, newLine)
	}
	return translatedLines, nil
}

// Helper function to extract text from subtitle items
func extractText(items []astisub.Item) []string {
	texts := []string{}
	for _, item := range items {
		for _, line := range item.Lines {
			for _, lineItem := range line.Items {
				texts = append(texts, lineItem.Text)
			}
		}
	}
	return texts
}
