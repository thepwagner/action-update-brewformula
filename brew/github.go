package brew

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update/updater"
	"golang.org/x/mod/semver"
)

func semverIsh(s string) string {
	if semver.IsValid(s) {
		return s
	}

	if vt := fmt.Sprintf("v%s", s); semver.IsValid(vt) {
		return vt
	}
	return ""
}

func semverSort(versions []string) []string {
	sort.Slice(versions, func(i, j int) bool {
		// Prefer strict semver ordering:
		if c := semver.Compare(semverIsh(versions[i]), semverIsh(versions[j])); c > 0 {
			return true
		} else if c < 0 {
			return false
		}
		// Failing that, prefer the most specific version:
		return strings.Count(versions[i], ".") > strings.Count(versions[j], ".")
	})
	return versions
}

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
