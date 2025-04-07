package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func translateCmd() *cobra.Command {
	var inputFile string
	var outputFile string
	var targetLanguage string

	cmd := &cobra.Command{
		Use:   "translate",
		Short: "Translate subtitles to another language",
		Long:  `Translate a video subtitle file to another language using the specified target language.`,
		Run: func(cmd *cobra.Command, args []string) {
			if inputFile == "" || targetLanguage == "" {
				fmt.Println("Input file and target language must be specified.")
				os.Exit(1)
			}
			fmt.Printf("Translating %s to %s and saving to %s\n", inputFile, targetLanguage, outputFile)
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input subtitle file")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output subtitle file")
	cmd.Flags().StringVarP(&targetLanguage, "language", "l", "", "Target language for translation")

	return cmd
}
