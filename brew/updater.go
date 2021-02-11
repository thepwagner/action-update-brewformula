package brew

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/thepwagner/action-update/updater"
)

type Updater struct {
	root   string
	client *http.Client

	ghRepos *github.RepositoriesService
}

func NewUpdater(root string) *Updater {
	client := http.DefaultClient
	gh := github.NewClient(client)
	return &Updater{
		root:    root,
		client:  client,
		ghRepos: gh.Repositories,
	}
}

func (u Updater) Name() string {
	return "brew"
}

func (u Updater) Dependencies(context.Context) ([]updater.Dependency, error) {
	formulae, err := filepath.Glob(filepath.Join(u.root, "*.rb"))
	if err != nil {
		return nil, fmt.Errorf("globbing formulae: %w", err)
	}

	deps := make([]updater.Dependency, 0, len(formulae))
	for _, f := range formulae {
		formulaDeps, err := parseFormula(f)
		if err != nil {
			return nil, fmt.Errorf("parsing formula %s: %w", f, err)
		}
		deps = append(deps, formulaDeps...)
	}
	return deps, nil
}

func (u Updater) Check(ctx context.Context, dep updater.Dependency, filter func(string) bool) (*updater.Update, error) {
	switch {
	case strings.HasPrefix(dep.Path, "https://github.com/"):
		return checkGitHubRelease(ctx, u.ghRepos, dep)
	case strings.HasPrefix(dep.Path, "https://golang.org/dl/go"):
		return checkGolangRelease(ctx, u.client, dep)
	default:
		return checkApacheRelease(ctx, u.client, dep)
	}
}

func (u Updater) ApplyUpdate(ctx context.Context, update updater.Update) error {
	panic("implement me")
}
