package brew

import (
	"github.com/thepwagner/action-update/actions/updateaction"
	"github.com/thepwagner/action-update/updater"
)

type Environment struct {
	updateaction.Environment
	GPG bool `env:"INPUT_GPG" envDefault:"false"`
}

func (e *Environment) NewUpdater(root string) updater.Updater {
	u := NewUpdater(root, WithGPG(e.GPG))
	u.pathFilter = e.Ignored
	return u
}
