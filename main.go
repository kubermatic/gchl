/*
Copyright 2020 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"k8c.io/gchl/pkg/action"
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
	app.Version = version
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
