package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"github.com/vanclief/coderunner/cmd"
	"github.com/vanclief/coderunner/files"
	"github.com/vanclief/ez"
)

func main() {
	app := cli.NewApp()
	app.Name = "coderunner"
	app.Usage = "A context-aware code extraction tool to run LLMs in your codebase."
	app.Version = "0.2.3"

	app.Commands = []*cli.Command{
		cmd.ScopeCmd(),
		cmd.LLMCmd(),
	}

	err := files.Init()
	if err != nil {
		color.Red(ez.ErrorMessage(err))
		return
	}

	err = app.Run(os.Args)
	if err != nil {
		errorCode := ez.ErrorCode(err)
		if errorCode == ez.EINTERNAL {
			color.Red(err.Error())
		} else {
			color.Red(ez.ErrorMessage(err))
		}
	}
}
