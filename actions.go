package main

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jlmanriquez/dep-updater/updater"
	"github.com/jlmanriquez/dep-updater/updater/cfg"
	"github.com/jlmanriquez/dep-updater/updater/trace"

	"github.com/urfave/cli/v2"
)

type UpdaterProcess func(prj cfg.Project, wg *sync.WaitGroup) error

func ConfigAppAction(c *cli.Context) error {
	level := trace.InfoLevel
	trace.Init(level)
	return nil
}

// RunUpdaterAction execute the complete process following the json configuration file.
// First, create o change to branch working_branch according to json configuration.
// Second, updates the described dependencies into projects section of the json configuration.
func RunUpdaterAction(c *cli.Context) error {
	configFile := c.String("config")
	var config cfg.Config
	if err := ReadJsonFile(configFile, &config); err != nil {
		return err
	}

	updater := updater.New(config)

	return executeAction(config, updater.UpdateProjectDependencies)
}

// CommitAndPushAction only run commit and push. This command considers the porjects
// with the flag push, in true only.
func CommitAndPushAction(c *cli.Context) error {
	configFile := c.String("config")
	var config cfg.Config
	if err := ReadJsonFile(configFile, &config); err != nil {
		return err
	}

	updater := updater.New(config)

	return executeAction(config, updater.CommitAndPush)
}

func executeAction(config cfg.Config, updaterProcess UpdaterProcess) error {
	trace.InfoConsole("", fmt.Sprintf("🛠️  init working into: '%s'", config.WorkspaceHome))

	projectsToUpdate := config.Projects
	if len(projectsToUpdate) == 0 {
		err := errors.New("project names to update, not configured")
		trace.ErrorConsole("", err.Error())
		return err
	}

	var wg sync.WaitGroup
	counter := struct {
		Ok       int
		Fail     int
		Disabled int
	}{}

	for _, p := range projectsToUpdate {
		if !p.Enabled {
			counter.Disabled += 1
			trace.InfoConsole(p.Name, "ℹ️ Not considered")
			continue
		}

		wg.Add(1)

		go func(project cfg.Project) {
			if err := updaterProcess(project, &wg); err != nil {
				counter.Fail += 1
				trace.ErrorConsole(project.Name, fmt.Sprintf("❌ %s", err.Error()))
			} else {
				counter.Ok += 1
				trace.InfoConsole(project.Name, "✅ updated successfully")
			}
		}(p)
	}

	wg.Wait()
	trace.InfoConsole(
		"",
		fmt.Sprintf("🏁 done, %d projects. OK: %d, Fail: %d, Disabled: %d",
			counter.Ok+counter.Disabled+counter.Fail,
			counter.Ok,
			counter.Fail,
			counter.Disabled))
	return nil
}
