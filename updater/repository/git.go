package repository

import (
	"errors"
	"fmt"

	"github.com/jlmanriquez/dep-updater/updater/trace"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Repository is a simple wrapper for git.Repository
type Repository struct {
	g             *git.Repository
	auth          transport.AuthMethod
	url           string
	projectName   string
	workspacePath string
}

// Config RepositoryConfig is the required configuration to create a new Repository.
type Config struct {
	Auth          *AuthConfig
	RemoteURL     string
	WorkspacePath string
	ProjectName   string
}

type AuthConfig struct {
	Username string
	Password string
}

const (
	branchRefPath = "refs/heads/"
)

// New create a new Repository pointer, that wrapping a *git.Repository instance
func New(config Config) (*Repository, error) {
	repo, err := git.PlainOpenWithOptions(config.WorkspacePath, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, fmt.Errorf("error openning repository %s... %s", config.WorkspacePath, err.Error())
	}

	var authMethod transport.AuthMethod
	if config.Auth != nil {
		trace.Debugp(config.ProjectName, "using BasicAuth for repository")
		authMethod = &http.BasicAuth{
			Username: config.Auth.Username,
			Password: config.Auth.Password,
		}
	}

	return &Repository{
		g:           repo,
		auth:        authMethod,
		url:         config.RemoteURL,
		projectName: config.ProjectName,
	}, nil
}

// ChangeToBranch execute a git checkout to an existing branch
func (r *Repository) ChangeToBranch(branchName string) error {
	refPath := fmt.Sprintf("%s%s", branchRefPath, branchName)

	ref, err := r.g.Reference(plumbing.ReferenceName(refPath), true)
	if err != nil {
		return fmt.Errorf("change to branch fail... %s", err.Error())
	}

	options := &git.CheckoutOptions{
		Branch: ref.Name(),
	}
	w, _ := r.g.Worktree()

	// git checkout <branchName>
	if err := w.Checkout(options); err != nil {
		return fmt.Errorf("can't possible checkout to '%s' branch... %s", branchName, err.Error())
	}

	return nil
}

// CreateBranch create a new branch, similar to 'git checkout -b' command.
func (r *Repository) CreateBranch(fromBranchName, newBranchName string) error {
	if fromBranchName == "" || newBranchName == "" {
		return errors.New("the name of the new or originating branch is not defined")
	}

	newRefPath := fmt.Sprintf("%s%s", branchRefPath, newBranchName)
	fromRefPath := fmt.Sprintf("%s%s", branchRefPath, fromBranchName)

	fromRef, err := r.g.Reference(plumbing.ReferenceName(fromRefPath), true)
	if err != nil {
		return fmt.Errorf("error creating branch, %s not found... %s", fromRefPath, err.Error())
	}

	// git branch <new_branch>
	newBranchRef := plumbing.NewHashReference(plumbing.ReferenceName(newRefPath), fromRef.Hash())

	if err := r.g.Storer.SetReference(newBranchRef); err != nil {
		return fmt.Errorf("error saved reference in the storage... %s", err.Error())
	}

	w, _ := r.g.Worktree()

	// git checkout <new_branch>
	if err := w.Checkout(&git.CheckoutOptions{Branch: newBranchRef.Name()}); err != nil {
		return fmt.Errorf("checkout to new branch fail... %s", err.Error())
	}

	trace.Debugp(r.projectName, fmt.Sprintf("new local branch '%s' created successfully", newBranchName))
	return nil
}

// GetCurrentBranch returns a reference to branch or commit of the header
func (r *Repository) GetCurrentBranch() (*plumbing.Reference, error) {
	head, err := r.g.Head()
	if err != nil {
		return nil, fmt.Errorf("could not get a head... %s", err.Error())
	}
	return head, nil
}

// CommitAndPush realize a commit of all changes and remote push.
// The commit is executed over working tree.
func (r *Repository) CommitAndPush(comment string) error {
	w, _ := r.g.Worktree()

	_, err := w.Commit(comment, &git.CommitOptions{All: true})
	if err != nil {
		return fmt.Errorf("commit fail... %s", err.Error())
	}

	pushOptions := &git.PushOptions{
		RemoteName: "origin",
		Auth:       r.auth,
	}

	if err := r.g.Push(pushOptions); err != nil {
		return fmt.Errorf("could not do push... %s", err.Error())
	}

	trace.Debugp(r.projectName, "push successful")
	return nil
}

// BranchExist return true, if exist branch reference for name, otherwise return false.
func (r *Repository) BranchExist(name string) (bool, error) {
	pathRef := fmt.Sprintf("%s%s", branchRefPath, name)

	if _, err := r.g.Reference(plumbing.ReferenceName(pathRef), true); err != nil {
		if err == plumbing.ErrReferenceNotFound {
			return false, nil
		} else {
			return false, fmt.Errorf("could not get the branch reference %s...%s", name, err.Error())
		}
	}
	return true, nil
}

// Clone clone a remote repository.
func (r *Repository) Clone(url string) error {
	_, err := git.PlainClone(r.workspacePath, false, &git.CloneOptions{
		URL:  r.url,
		Auth: r.auth,
	})
	if err != nil {
		return fmt.Errorf("error cloning repo '%s'...%s", r.url, err.Error())
	}
	return nil
}
