package brew

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update/updater"
	"golang.org/x/mod/semver"
)

func checkGitHubRelease(ctx context.Context, repos *github.RepositoriesService, dep updater.Dependency) (*updater.Update, error) {
	releases, err := listGitHubReleases(ctx, repos, dep)
	if err != nil {
		return nil, fmt.Errorf("listing releases: %w", err)
	}

	depVer := semverIsh(dep.Version)
	for _, release := range releases {
		switch semver.Compare(depVer, release) {
		case -1:
			return &updater.Update{
				Path:     dep.Path,
				Previous: dep.Version,
				Next:     release,
			}, nil
		case 1:
			return nil, nil
		}
	}
	return nil, nil
}

func listGitHubReleases(ctx context.Context, repos *github.RepositoriesService, dependency updater.Dependency) ([]string, error) {
	owner, name := parseGitHubRelease(dependency.Path)
	releases, _, err := repos.ListReleases(ctx, owner, name, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("querying for releases: %w", err)
	}
	log := logrus.WithFields(logrus.Fields{
		"owner": owner,
		"repo":  name,
	})
	log.WithField("releases", len(releases)).Debug("fetched releases")

	ret := make([]string, 0, len(releases))
	for _, release := range releases {
		ret = append(ret, release.GetTagName())
	}

	sort.SliceStable(ret, func(i, j int) bool {
		return semver.Compare(ret[i], ret[j]) > 0
	})
	return ret, nil
}

func parseGitHubRelease(path string) (owner, repoName string) {
	parsed, _ := url.Parse(path)
	pathSplit := strings.SplitN(parsed.Path, "/", 4)
	return pathSplit[1], pathSplit[2]
}

func updatedGitHubHash(ctx context.Context, client *http.Client, repos *github.RepositoriesService, update updater.Update, oldHash string) (string, error) {
	// Fetch the previous release:
	owner, repoName := parseGitHubRelease(update.Path)
	prevRelease, err := getReleaseByTag(ctx, repos, owner, repoName, update.Previous)
	if err != nil {
		return "", err
	}

	// First pass, does the project release a SHASUMS etc file we can grab?
	for _, prevAsset := range prevRelease.Assets {
		log := logrus.WithField("name", prevAsset.GetName())
		oldAsset, err := isShasumAsset(ctx, client, prevAsset, oldHash)
		if err != nil {
			log.WithError(err).Warn("inspecting potential hash asset")
			continue
		} else if len(oldAsset) == 0 {
			log.Debug("old shasum asset not found")
			continue
		}
		log.Debug("identified shasum asset in previous release")

		// The previous release contained a shasum file that contained the previous hash
		// Does the new release have the same file?
		newHash, err := updatedHashFromShasumAsset(ctx, client, prevAsset, oldAsset, oldHash, update)
		if err != nil {
			log.WithError(err).Warn("fetching updated hash asset")
			continue
		}
		if newHash != "" {
			log.Debug("fetched corresponding shasum asset from new release")
			return newHash, nil
		}
	}

	// There are no shasum files - get downloading
	logrus.Debug("shasum file not found, searching files from previous release")
	for _, prevAsset := range prevRelease.Assets {
		log := logrus.WithField("name", prevAsset.GetName())
		h, err := isHashAsset(ctx, client, prevAsset.GetBrowserDownloadURL(), oldHash)
		if err != nil {
			log.WithError(err).Warn("checking hash of previous assets")
			continue
		} else if !h {
			continue
		}
		log.Debug("identified hashed asset in previous release")

		// This asset from a previous release matched the previous hash
		// Does the new release have the same file?
		newHash, err := updatedHashFromAsset(ctx, client, prevAsset.GetBrowserDownloadURL(), update, oldHash)
		if err != nil {
			return "", err
		}
		if newHash != "" {
			log.Debug("fetched corresponding asset from new release")
			return newHash, nil
		}
	}

	logrus.Debug("not found in release assets, checking source archives...")
	for _, sourceURL := range sourceURLs(prevRelease) {
		ok, err := isHashAsset(ctx, client, sourceURL, oldHash)
		if err != nil {
			return "", err
		}
		if !ok {
			continue
		}
		logrus.WithField("source_url", sourceURL).Debug("found as source archive")
		return updatedHashFromAsset(ctx, client, sourceURL, update, oldHash)
	}

	return "", nil
}

func sourceURLs(prevRelease *github.RepositoryRelease) []string {
	archiveByTagRoot := strings.ReplaceAll(prevRelease.GetHTMLURL(), "releases/tag", "archive")
	return []string{
		prevRelease.GetTarballURL(),
		prevRelease.GetZipballURL(),
		fmt.Sprintf("%s.tar.gz", archiveByTagRoot),
		fmt.Sprintf("%s.zip", archiveByTagRoot),
	}
}

func getReleaseByTag(ctx context.Context, repos *github.RepositoriesService, owner, repoName, version string) (*github.RepositoryRelease, error) {
	release, _, err := repos.GetReleaseByTag(ctx, owner, repoName, version)
	if err == nil {
		return release, nil
	}

	// If 1.2.3 is not found, try v1.2.3
	if asSemver := semverIsh(version); asSemver != version {
		var githubErr *github.ErrorResponse
		if errors.As(err, &githubErr) && githubErr.Message == "Not Found" {
			release, _, err := repos.GetReleaseByTag(ctx, owner, repoName, asSemver)
			if err == nil {
				return release, nil
			}
		}
	}

	return nil, err
}

// isShasumAsset returns true if the release asset is a SHASUMS file containing the previous hash
func isShasumAsset(ctx context.Context, client *http.Client, asset *github.ReleaseAsset, oldHash string) ([]string, error) {
	if asset.GetSize() > 1024 {
		return nil, nil
	}

	req, err := http.NewRequest("GET", asset.GetBrowserDownloadURL(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	s := string(b)
	if !strings.Contains(s, oldHash) {
		return nil, nil
	}
	return strings.Split(s, "\n"), nil
}

func updatedHashFromShasumAsset(ctx context.Context, client *http.Client, asset *github.ReleaseAsset, oldContents []string, oldHash string, update updater.Update) (string, error) {
	res, err := getUpdatedAsset(ctx, client, asset.GetBrowserDownloadURL(), update)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	s := string(b)

	// If there's one line, extract the checksum and return it:
	if len(oldContents) == 1 {
		return strings.SplitN(s, " ", 2)[0], nil
	}

	// If there's multiple lines, find the file corresponding the old hash:
	var hashedFile string
	for _, oldLine := range oldContents {
		split := strings.SplitN(oldLine, " ", 2)
		if split[0] == oldHash {
			hashedFile = split[1]
		}
	}
	if hashedFile == "" {
		return "", nil
	}

	logrus.WithField("fn", hashedFile).Debug("identified hashed file in shasum asset")
	for _, newLine := range strings.Split(s, "\n") {
		split := strings.SplitN(newLine, " ", 2)
		if len(split) == 1 {
			continue
		}
		if split[1] == hashedFile {
			return split[0], nil
		}
	}

	return "", nil
}

func getUpdatedAsset(ctx context.Context, client *http.Client, oldURL string, update updater.Update) (*http.Response, error) {
	newURL := updatedURL(oldURL, update)
	req, err := http.NewRequest("GET", newURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return client.Do(req)
}

func updatedURL(oldURL string, update updater.Update) string {
	newURL := strings.ReplaceAll(oldURL, update.Previous, update.Next)
	newURL = strings.ReplaceAll(newURL, update.Previous[1:], update.Next[1:])
	return newURL
}

func isHashAsset(ctx context.Context, client *http.Client, assetURL string, oldHash string) (bool, error) {
	h, ok := hasher(oldHash)
	if !ok {
		return false, nil
	}

	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		return false, err
	}
	req = req.WithContext(ctx)

	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if _, err := io.Copy(h, res.Body); err != nil {
		return false, err
	}
	sum := h.Sum(nil)
	newHash := fmt.Sprintf("%x", sum)
	logrus.WithFields(logrus.Fields{
		"url":  assetURL,
		"hash": newHash,
	}).Debug("downloaded asset")
	return newHash == oldHash, nil
}

func hasher(oldHash string) (hash.Hash, bool) {
	switch len(oldHash) {
	case 40:
		logrus.Warn("consider upgrading formula from sha1")
		return sha1.New(), true
	case 64:
		return sha256.New(), true
	case 128:
		return sha512.New(), true
	default:
		return nil, false
	}
}

func updatedHashFromAsset(ctx context.Context, client *http.Client, assetURL string, update updater.Update, oldHash string) (string, error) {
	res, err := getUpdatedAsset(ctx, client, assetURL, update)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	h, _ := hasher(oldHash)
	if _, err := io.Copy(h, res.Body); err != nil {
		return "", err
	}
	sum := h.Sum(nil)
	newHash := fmt.Sprintf("%x", sum)
	logrus.WithFields(logrus.Fields{
		"url":  assetURL,
		"hash": newHash,
	}).Debug("downloaded updated asset")
	return newHash, nil
}
