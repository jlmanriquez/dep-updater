package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "update",
				Usage:  "will update dependencies versions in package.json file in the projects",
				Action: ConfigAppAction,
				After:  RunUpdaterAction,
			},
			{
				Name:   "push",
				Usage:  "will commit and push the configured projects",
				Action: ConfigAppAction,
				After:  CommitAndPushAction,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "dependencies.json",
				Usage:   "file with migration configuration. This could be a complete path",
			},
		},
		Name:  "dep-updater",
		Usage: "Updates dependencies version for Angular projects",
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func ReadJsonFile(path string, out interface{}) error {
	contentFile, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file '%s'... %s", path, err.Error())
	}
	if err := json.Unmarshal(contentFile, out); err != nil {
		return fmt.Errorf("error creating a json object... %s", err.Error())
	}
	return nil
}
