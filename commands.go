package main

import (
	"github.com/kubermatic/gchl/pkg/action"
	"github.com/urfave/cli"
)

func getCommands(action *action.Action, app *cli.App) []cli.Command {
	return []cli.Command{
		{
			Name:        "between",
			Usage:       "Create a changelog for changes between to references.",
			Description: "Create a changelog for changes between to references.",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "release-notes",
					Usage: "use pull request titles instead of release-notes tag in the pull request message",
				},
				cli.BoolFlag{
					Name:  "release-notes-none",
					Usage: "list PRs that contain empty release-notes",
				},
			},
			Action: func(c *cli.Context) error {
				return action.GenerateChangelogBetween(c)
			},
		},
		{
			Name:        "since",
			Usage:       "Create a changelog for changes since reference.",
			Description: "Create a changelog for changes between to reference.",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "release-notes",
					Usage: "use pull request titles instead of release-notes tag in the pull request message",
				},
				cli.BoolFlag{
					Name:  "release-notes-none",
					Usage: "list PRs that contain empty release-notes",
				},
			},
			Action: func(c *cli.Context) error {
				return action.GenerateChangelogSince(c)
			},
		},
	}
}

func getGlobalFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "for-version, f",
			Usage: "Specify a version name that will be shown in changelog output",
			Value: "v0.0.0",
		},
		cli.StringFlag{
			Name:  "repository, repo",
			Usage: "The file path to the directory containing the git repository to be used",
			Value: getCurrentWorkDir(),
		},
		cli.StringFlag{
			Name:  "remote, r",
			Usage: "The remote github repository url",
			Value: "",
		},
		cli.StringFlag{
			Name:   "token, t",
			Usage:  "Your personal access token provided by GitHub Inc. See: https://github.com/settings/tokens",
			EnvVar: "GCHL_GITHUB_TOKEN",
			Value:  "",
		},
	}
}
