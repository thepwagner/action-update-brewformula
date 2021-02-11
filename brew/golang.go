package brew

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/thepwagner/action-update/updater"
	"golang.org/x/mod/semver"
)

const golangIndexURL = "https://golang.org/dl/?mode=json"

type golangIndexedVersion struct {
	Version string `json:"version"`
	Files   []struct {
		Filename string `json:"filename"`
		Sha256   string `json:"sha256"`
	}
}

func checkGolangRelease(ctx context.Context, client *http.Client, dep updater.Dependency) (*updater.Update, error) {
	versions, err := fetchGolangIndex(ctx, client)
	if err != nil {
		return nil, err
	}

	depVersion := semverIsh(dep.Version)
	for _, v := range versions {
		version := semverIsh(v.Version[2:])
		if semver.Compare(version, depVersion) > 0 {
			return &updater.Update{
				Path:     dep.Path,
				Previous: dep.Version,
				Next:     v.Version[2:],
			}, nil
		}
	}
	return nil, nil
}

func fetchGolangIndex(ctx context.Context, client *http.Client) ([]golangIndexedVersion, error) {
	req, err := http.NewRequest("GET", golangIndexURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var versions []golangIndexedVersion
	if err := json.NewDecoder(res.Body).Decode(&versions); err != nil {
		return nil, err
	}
	return versions, nil
}
