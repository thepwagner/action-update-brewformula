package brew

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/google/go-github/v33/github"
	"github.com/thepwagner/action-update/updater"
)

type Updater struct {
	root       string
	client     *http.Client
	gpg        bool
	pathFilter func(string) bool

	ghRepos *github.RepositoriesService
}

func NewUpdater(root string, opts ...UpdaterOpt) *Updater {
	client := http.DefaultClient
	gh := github.NewClient(client)
	u := &Updater{
		root:    root,
		client:  client,
		ghRepos: gh.Repositories,
	}
	for _, o := range opts {
		o(u)
	}
	return u
}

type UpdaterOpt func(*Updater)

func WithGPG(gpg bool) UpdaterOpt {
	return func(u *Updater) {
		u.gpg = gpg
	}
}

func (u Updater) Name() string {
	return "brew"
}

func (u Updater) Dependencies(context.Context) ([]updater.Dependency, error) {
	var deps []updater.Dependency
	err := u.eachFormula(func(_, formula string) error {
		formulaDeps, err := parseFormulaDeps(formula)
		if err == nil {
			deps = append(deps, formulaDeps...)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func (u Updater) Check(ctx context.Context, dep updater.Dependency, filter func(string) bool) (*updater.Update, error) {
	// FIXME: pass the filter function
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
	return u.eachFormula(func(path, formula string) error {
		oldnew := []string{
			update.Previous, update.Next,
		}
		if shasums := parseFormulaHashes(formula); len(shasums) == 1 {
			oldHash := shasums[0]
			newHash, err := u.updatedHash(ctx, update, oldHash)
			if err != nil {
				return fmt.Errorf("finding updated hash: %w", err)
			}
			if newHash != "" {
				oldnew = append(oldnew, oldHash, newHash)
			}
		}

		replaced := strings.NewReplacer(oldnew...).Replace(formula)
		return ioutil.WriteFile(path, []byte(replaced), 0600)
	})
}

func (u Updater) updatedHash(ctx context.Context, update updater.Update, oldHash string) (string, error) {
	switch {
	case strings.HasPrefix(update.Path, "https://github.com/"):
		return updatedGitHubHash(ctx, u.client, u.ghRepos, update, oldHash)
	case strings.HasPrefix(update.Path, "https://golang.org/dl/go"):
		return updatedGolangHash(ctx, u.client, update, oldHash)
	default:
		return updatedApacheHash(ctx, u.client, update, oldHash, u.gpg)
	}
}

func (u *Updater) eachFormula(process func(path, formula string) error) error {
	formulae, err := doublestar.Glob(filepath.Join(u.root, "**", "*.rb"))
	if err != nil {
		return fmt.Errorf("globbing formulae: %w", err)
	}

	for _, f := range formulae {
		if u.pathFilter != nil && u.pathFilter(f) {
			continue
		}

		formula, err := ioutil.ReadFile(f)
		if err != nil {
			return fmt.Errorf("reading formula %s: %w", f, err)
		}
		if err := process(f, string(formula)); err != nil {
			return fmt.Errorf("processing formula %s: %w", f, err)
		}
	}
	return nil
}
