package cmd

import (
	"fmt"
	"github.com/andresbott/substrans/app/metainfo"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

// Execute is the entry point for the command line
func Execute() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "substrans",
		Short: "substrans LLM subtitle translation tool",
	}

	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		_ = cmd.Help()
		return nil
	})

	cmd.AddCommand(
		versionCmd(),
		translateCmd(),
	)

	return cmd
}

func versionCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "version",
		Short: "version ",
		Long:  `version long`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", metainfo.Version)
			fmt.Printf("Build date: %s\n", metainfo.BuildTime)
			fmt.Printf("Commit sha: %s\n", metainfo.ShaVer)
			fmt.Printf("Compiler: %s\n", runtime.Version())
		},
	}

	// hide persistent flag on this command
	cmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		_ = command.Flags().MarkHidden("pers")
		command.Parent().HelpFunc()(command, strings)
	})

	return &cmd
}
