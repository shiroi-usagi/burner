package version

import (
	"flag"
	"fmt"
	"github.com/shiroi-usagi/burner"
	"github.com/shiroi-usagi/pkg/command"
)

var Cmd = &command.Subcommand{
	Name:  "version",
	Short: "print version",
	Long: `Version prints the build information for Burner executables.
`,
	Flag: flag.NewFlagSet("", flag.ExitOnError),
	Run: func(cmd *command.Subcommand, args []string) {
		info := burner.GetVersionInfo()
		fmt.Println("Burner", info.Version)
		fmt.Println("    built on", info.BuildDate)
		fmt.Println("    built by", info.BuiltBy)
	},
}
