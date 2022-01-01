package main

import (
	"fmt"
	"github.com/shiroi-usagi/burner/internal/burn"
	"github.com/shiroi-usagi/burner/internal/prepare"
	"github.com/shiroi-usagi/burner/internal/version"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	cmd := &cobra.Command{
		Use:   "burner",
		Short: "Burner is a tool for transcoding video.",

		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	cmd.AddCommand(
		burn.Cmd,
		version.Cmd,
		prepare.Cmd,
	)
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
