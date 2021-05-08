package main

import (
	"flag"
	"github.com/shiroi-usagi/burner/cmd/burner/internal/burn"
	"github.com/shiroi-usagi/burner/cmd/burner/internal/prepare"
	"github.com/shiroi-usagi/burner/cmd/burner/internal/version"
	"github.com/shiroi-usagi/pkg/command"
	"os"
)

func main() {
	cmd := &command.Command{
		Name:      "burner",
		Arguments: "[arguments]",
		Long:      `Burner is a tool for transcoding video.`,
		Flag:      flag.CommandLine,
		Commands: []*command.Subcommand{
			burn.Cmd,
			version.Cmd,
			prepare.Cmd,
		},
		Run: command.BaseRun,
	}
	flag.Usage = func() {
		cmd.Usage(os.Stdout)
	}
	cmd.Flag.Parse(os.Args[1:])
	cmd.Run(cmd, cmd.Flag.Args())
}
