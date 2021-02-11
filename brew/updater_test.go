package brew_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-brewformula/brew"
	"github.com/thepwagner/action-update/updater"
	"github.com/thepwagner/action-update/updatertest"
)

type testFactory struct{}

func (t testFactory) NewUpdater(root string) updater.Updater { return brew.NewUpdater(root) }

var (
	azCopy1070 = updater.Dependency{Path: "https://github.com/Azure/azure-storage-azcopy/archive/v#{version}.tar.gz", Version: "10.7.0"}
	hadoop260  = updater.Dependency{Path: "https://archive.apache.org/dist/hadoop/core/hadoop-2.6.0/hadoop-2.6.0.tar.gz", Version: "2.6.0"}
	golang1156 = updater.Dependency{Path: "https://golang.org/dl/go#{VERSION}.linux-amd64.tar.gz", Version: "1.15.6"}
	libvirt102 = updater.Dependency{Path: "https://libvirt.org/sources/libvirt-#{VERSION}.tar.gz", Version: "1.0.2"}
)

func TestUpdater_Dependencies(t *testing.T) {
	updatertest.DependenciesFixtures(t, &testFactory{}, map[string][]updater.Dependency{
		"azcopy": {azCopy1070},
		"debian": {
			{Path: "https://libvirt.org/sources/libvirt-1.0.2.tar.gz", Version: "1.0.2"},
		},
		"hadoop":     {hadoop260},
		"versionvar": {libvirt102},
		"go":         {golang1156},
	})
}

func TestUpdater_Check_GitHubRelease(t *testing.T) {
	update := updatertest.CheckInFixture(t, "azcopy", &testFactory{}, azCopy1070, nil)
	require.NotNil(t, update)
	t.Log(update.Next)
}

func TestUpdater_Check_Golang(t *testing.T) {
	update := updatertest.CheckInFixture(t, "go", &testFactory{}, golang1156, nil)
	require.NotNil(t, update)
	t.Log(update.Next)
}

func TestUpdater_Check_Apache(t *testing.T) {
	update := updatertest.CheckInFixture(t, "hadoop", &testFactory{}, hadoop260, nil)
	require.NotNil(t, update)
	t.Log(update.Next)
}

func TestUpdater_Check_Libvirt(t *testing.T) {
	update := updatertest.CheckInFixture(t, "versionvar", &testFactory{}, libvirt102, nil)
	require.NotNil(t, update)
	t.Log(update.Next)
}
