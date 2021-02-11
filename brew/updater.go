package brew

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/thepwagner/action-update/updater"
)

type Updater struct {
	root string
}

func NewUpdater(root string) *Updater {
	return &Updater{root: root}
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
	panic("implement me")
}

func (u Updater) ApplyUpdate(ctx context.Context, update updater.Update) error {
	panic("implement me")
}
