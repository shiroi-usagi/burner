package version

import (
	"fmt"
	"github.com/shiroi-usagi/burner"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "print version",
	Long:  "Version prints the build information for Burner executables.",
	Run: func(_ *cobra.Command, _ []string) {
		info := burner.GetVersionInfo()
		fmt.Println("Burner", info.Version)
		fmt.Println("    built on", info.BuildDate)
		fmt.Println("    built by", info.BuiltBy)
	},
}
