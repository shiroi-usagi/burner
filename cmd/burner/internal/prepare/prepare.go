package prepare

import (
	"flag"
	"github.com/shiroi-usagi/pkg/command"
	"log"
	"os"
	"path/filepath"
)

var Cmd = &command.Subcommand{
	Name:  "prepare",
	Short: "prepare environment",
	Long: `Prepare environments by creating the needed 'in' and 'out' directories.
`,
	Flag: flag.NewFlagSet("", flag.ExitOnError),
	Run: func(cmd *command.Subcommand, args []string) {
		if err := os.Mkdir(filepath.Join(".", "in"), 0755); err != nil {
			log.Fatal("was not able to create input dir")
		}
		if err := os.Mkdir(filepath.Join(".", "out"), 0755); err != nil {
			log.Fatal("was not able to create output dir")
		}
	},
}
