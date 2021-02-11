package brew

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/thepwagner/action-update/updater"
	"golang.org/x/mod/semver"
)

func checkApacheRelease(ctx context.Context, client *http.Client, dep updater.Dependency) (*updater.Update, error) {
	candidates, err := listApacheVersions(ctx, client, dep)
	if err != nil {
		return nil, err
	}

	depVersion := semverIsh(dep.Version)
	for _, version := range candidates {
		if semver.Compare(depVersion, semverIsh(version)) < 0 {
			return &updater.Update{
				Path:     dep.Path,
				Previous: dep.Version,
				Next:     version,
			}, nil
		}
	}
	return nil, nil
}

func listApacheVersions(ctx context.Context, client *http.Client, dep updater.Dependency) ([]string, error) {
	// Split the URL on the last component that includes the version:
	listingURL, nextPath, err := getListing(dep)
	if err != nil {
		return nil, err
	}

	// Fetch that page, and hope it's an basic HTML listing...
	req, err := http.NewRequest("GET", listingURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Search for links matching the previous last segment (e.g. my-dep-1.0.0)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	filter := regexp.MustCompile(semverRe.ReplaceAllString(nextPath, `(\d+\.\d+\.\d+)`) + "/?")

	var ret []string
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		match := filter.FindStringSubmatch(s.Text())
		if len(match) > 0 {
			ret = append(ret, match[1])
		}
	})
	semverSort(ret)
	return ret, nil
}

func getListing(dep updater.Dependency) (string, string, error) {
	parsed, err := url.Parse(versionTemplate.ReplaceAllString(dep.Path, dep.Version))
	if err != nil {
		return "", "", err
	}
	paths := strings.Split(parsed.Path, "/")

	for i, pathPart := range paths {
		if strings.Contains(pathPart, dep.Version) {
			parsed.Path = strings.Join(paths[:i], "/")
			return parsed.String(), pathPart, nil
		}
	}
	return "", "", fmt.Errorf("could not find version in URL %s", dep.Path)
}
