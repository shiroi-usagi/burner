//go:build codegen
// +build codegen

// This program generates doctype.go.

package main

import (
	"bytes"
	"github.com/shiroi-usagi/burner/internal/burn"
	"io"
	"io/ioutil"
	"log"
)

var header = []byte(`# Burner

![logo](https://images.weserv.nl/?url=raw.githubusercontent.com/shiroi-usagi/burner/main/logo.png&w=64&mask=circle)

## Flags

`)

func main() {
	filename := "README.md"
	buf := bytes.NewBuffer(header)

	gen(buf)

	err := ioutil.WriteFile(filename, buf.Bytes(), 0666)
	if err != nil {
		log.Fatal(err)
	}
}

func gen(w io.Writer) {
	cmd := burn.Cmd
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(w)
	if flags.HasAvailableFlags() {
		io.WriteString(w, "```\n")
		flags.PrintDefaults()
		io.WriteString(w, "```\n\n")
	}
}
