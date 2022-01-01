package prepare

import (
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var Cmd = &cobra.Command{
	Use:   "prepare",
	Short: "prepare environment",
	Long:  "Prepare environments by creating the needed 'in' and 'out' directories.",
	Run: func(_ *cobra.Command, _ []string) {
		if err := os.Mkdir(filepath.Join(".", "in"), 0755); err != nil {
			log.Fatal("was not able to create input dir")
		}
		if err := os.Mkdir(filepath.Join(".", "out"), 0755); err != nil {
			log.Fatal("was not able to create output dir")
		}
	},
}
