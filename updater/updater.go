package updater

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/jlmanriquez/dep-updater/updater/cfg"
	"github.com/jlmanriquez/dep-updater/updater/repository"
	"github.com/jlmanriquez/dep-updater/updater/trace"
)

type Updater struct {
	config cfg.Config
}

func New(config cfg.Config) *Updater {
	return &Updater{config: config}
}

// UpdateProjectDependencies update the project dependencies, set the working git branch,
// modify the package.json file with the new ones the dependency versions described and
// execute the commit/push if this is configured in the configuration file.
func (u *Updater) UpdateProjectDependencies(prj cfg.Project, wg *sync.WaitGroup) error {
	defer wg.Done()
	trace.Infop(prj.Name, "init process")

	projectPath := fmt.Sprintf("%s/%s", u.config.WorkspaceHome, prj.Name)

	// create a new repository for each project to process
	repo, err := newRepository(prj, u.config)
	if err != nil {
		return err
	}

	if err := u.manageRepository(prj.Name, repo); err != nil {
		return err
	}

	if err := u.updateVersionsInFile(prj.Name, projectPath); err != nil {
		return err
	}

	if prj.Push {
		trace.Infop(prj.Name, "committing and pushing branch")
		if err := repo.CommitAndPush("update dependencies"); err != nil {
			return err
		}
		trace.Infop(prj.Name, "commit and push branch, done")
	}

	trace.Infop(prj.Name, "process completed successfully")
	return nil
}

// CommitAndPush makes the commit and push of the work_branch branch the project.
// If the branch does not exist, it returns the error.
func (u *Updater) CommitAndPush(prj cfg.Project, wg *sync.WaitGroup) error {
	defer wg.Done()
	trace.Infop(prj.Name, "initializing commit and push of the branch")

	// create a new repository for each project to process
	repo, err := newRepository(prj, u.config)
	if err != nil {
		return err
	}

	exist, err := repo.BranchExist(u.config.WorkingBranch)
	if err != nil {
		return err
	}

	if !exist {
		return fmt.Errorf("working_branch %s doesn't exist", u.config.WorkingBranch)
	}

	if err := repo.ChangeToBranch(u.config.WorkingBranch); err != nil {
		return err
	}

	if err := repo.CommitAndPush("commit from dep-updater"); err != nil {
		return err
	}

	trace.Infop(prj.Name, "the commit and push, done")
	return nil
}

func (u *Updater) updateVersionsInFile(projectName, projectPath string) error {
	trace.Infop(projectName, "updating package.json...")

	newLibs := u.config.Libraries
	if len(newLibs) == 0 {
		return fmt.Errorf("libraries configuration does not exists")
	}

	packageJsonFilePath := fmt.Sprintf("%s/package.json", projectPath)
	fileData, err := ioutil.ReadFile(packageJsonFilePath)
	if err != nil {
		return fmt.Errorf("error reading package.json file in %s", projectPath)
	}

	changes := false
	data := string(fileData)
	for libName := range newLibs {
		// finds the start position of the library name to update
		startCurrentLib := strings.Index(data, libName)
		exist := startCurrentLib >= 0

		if exist {
			// finds the end line in the current line, with the library to update
			endLine := strings.Index(data[startCurrentLib:], ",")
			lineToReplace := data[startCurrentLib : startCurrentLib+endLine]
			// the new line should not start with double quotes because the string.Index() does not consider it
			newLine := fmt.Sprintf("%s\": \"%s\"", libName, newLibs[libName])
			data = strings.Replace(data, lineToReplace, newLine, 1)
			changes = true
		}
	}

	if changes {
		ioutil.WriteFile(packageJsonFilePath, []byte(data), 0644)
		trace.Infop(projectName, "update package.json file, done")
	} else {
		trace.Infop(projectName, "no dependencies found in package.json")
	}

	return nil
}

func (u *Updater) manageRepository(projectName string, repo *repository.Repository) error {
	workingBranchName := u.config.WorkingBranch
	trace.Infop(projectName, "checking branches...")

	exist, err := repo.BranchExist(workingBranchName)
	if err != nil {
		return err
	}

	if exist {
		trace.Infop(projectName, fmt.Sprintf("branch '%s' exist", workingBranchName))

		currentBranch, err := repo.GetCurrentBranch()
		if err != nil {
			return err
		}

		// directory is in the working branch
		if string(currentBranch.Name()) == workingBranchName {
			trace.Infop(projectName, "project is in the correct branch")
			return nil
		}

		// if current directory is not in the working branch, change to working branch
		if err := repo.ChangeToBranch(workingBranchName); err != nil {
			return err
		}

		trace.Infop(projectName, fmt.Sprintf("checkout to '%s' branch, done", workingBranchName))
	} else {
		trace.Infop(projectName, fmt.Sprintf("branch '%s' doesn't exist... will try to create", workingBranchName))

		fromBranchName := u.config.CreateFrom
		if strings.TrimSpace(fromBranchName) == "" {
			return fmt.Errorf("it's not possible create the branch, create_from is not set")
		}

		if err := repo.CreateBranch(fromBranchName, workingBranchName); err != nil {
			return err
		}

		trace.Infop(
			projectName,
			fmt.Sprintf(
				"creation of the '%s' branch from '%s' branch, done",
				workingBranchName,
				fromBranchName))
	}

	return nil
}

func newRepository(prj cfg.Project, config cfg.Config) (*repository.Repository, error) {
	projectPath := fmt.Sprintf("%s/%s", config.WorkspaceHome, prj.Name)

	var auth *repository.AuthConfig
	// if repository require authentication
	if config.Repository.Username != "" && config.Repository.Password != "" {
		auth = &repository.AuthConfig{
			Username: strings.TrimSpace(config.Repository.Username),
			Password: strings.TrimSpace(config.Repository.Password),
		}
	}

	// create a new repository for each project to process
	repo, err := repository.New(repository.Config{
		Auth:          auth,
		RemoteURL:     config.Repository.URL,
		WorkspacePath: projectPath,
		ProjectName:   prj.Name,
	})

	if err != nil {
		return nil, err
	}

	return repo, nil
}
