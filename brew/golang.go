package brew

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update/updater"
	"golang.org/x/mod/semver"
)

const (
	golangIndexURL            = "https://golang.org/dl/?mode=json"
	golangIndexWithHistoryURL = "https://raw.githubusercontent.com/WillAbides/goreleases/main/releases.json"
)

type golangIndexedVersion struct {
	Version string `json:"version"`
	Files   []struct {
		Filename string `json:"filename"`
		Sha256   string `json:"sha256"`
	}
}

func checkGolangRelease(ctx context.Context, client *http.Client, dep updater.Dependency) (*updater.Update, error) {
	versions, err := fetchGolangIndex(ctx, client, golangIndexURL)
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

func updatedGolangHash(ctx context.Context, client *http.Client, update updater.Update, oldHash string) (string, error) {
	historic, err := historicVersion(ctx, client, oldHash)
	if err != nil {
		return "", err
	}
	if historic == "" {
		return "", nil
	}
	logrus.WithField("historic", historic).Debug("found old hash on artifact")

	versions, err := fetchGolangIndex(ctx, client, golangIndexURL)
	if err != nil {
		return "", err
	}
	targetFn := strings.ReplaceAll(historic, update.Previous, update.Next)
	for _, v := range versions {
		if v.Version[2:] != update.Next {
			continue
		}
		for _, f := range v.Files {
			if f.Filename == targetFn {
				logrus.WithField("updated", f.Filename).Debug("found updated file, updating hash")
				return f.Sha256, nil
			}
		}
	}

	return "", nil
}

func historicVersion(ctx context.Context, client *http.Client, oldHash string) (string, error) {
	historic, err := fetchGolangIndex(ctx, client, golangIndexWithHistoryURL)
	if err != nil {
		return "", err
	}
	for _, indexedVersion := range historic {
		for _, f := range indexedVersion.Files {
			if f.Sha256 == oldHash {
				return f.Filename, nil
			}
		}
	}

	return "", nil
}

func fetchGolangIndex(ctx context.Context, client *http.Client, indexURL string) ([]golangIndexedVersion, error) {
	req, err := http.NewRequest("GET", indexURL, nil)
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
