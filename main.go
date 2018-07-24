package main

import (
	"fmt"
	"github.com/kubermatic/gchl/pkg/action"
	"github.com/urfave/cli"
	"log"
	"os"
)

var (
	version = "v0.1"
)

func main() {

	action := action.New()

	app := cli.NewApp()
	app.Author = "Christian Bargmann"
	app.Email = "christian@loodse.com"
	app.Name = "gchl - A Go-written Changelog Generator"
	app.Usage = "Generate Changelogs, based on GitHub PRs"
	app.Version = fmt.Sprintf("%s", version)
	app.Commands = getCommands(action, app)
	app.Flags = getGlobalFlags()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getCurrentWorkDir() string {
	wd, _ := os.Getwd()
	return wd
}
