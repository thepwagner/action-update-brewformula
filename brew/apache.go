package brew

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
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

func updatedApacheHash(ctx context.Context, client *http.Client, update updater.Update, oldHash string, gpg bool) (string, error) {
	oldURL := versionTemplate.ReplaceAllString(update.Path, update.Previous)
	if ok, err := isHashAsset(ctx, client, oldURL, oldHash); err != nil {
		return "", err
	} else if !ok {
		return "", nil
	}

	var signature string
	if gpg {
		var err error
		signature, err = getUpdatedSignature(ctx, client, update, oldURL)
		if err != nil {
			logrus.WithError(err).Warn("error fetching updated signature, ignoring...")
		} else if signature == "" {
			logrus.Debug("no signature file detected")
		}
	}

	res, err := getUpdatedAsset(ctx, client, oldURL, update)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	updatedFn := filepath.Base(updatedURL(oldURL, update))
	h, _ := hasher(oldHash)
	var assetOut io.Writer
	var sigDir string
	if signature == "" {
		assetOut = h
	} else {
		sigDir, err = ioutil.TempDir("", "signature-*")
		if err != nil {
			return "", err
		}
		defer os.RemoveAll(sigDir)

		assetFile, err := os.OpenFile(filepath.Join(sigDir, updatedFn), os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return "", err
		}
		assetOut = io.MultiWriter(h, assetFile)
	}
	if _, err := io.Copy(assetOut, res.Body); err != nil {
		return "", err
	}

	if signature != "" {
		sigFn := fmt.Sprintf("%s.asc", updatedFn)
		if err := ioutil.WriteFile(filepath.Join(sigDir, sigFn), []byte(signature), 0600); err != nil {
			return "", err
		}
		cmd := exec.CommandContext(ctx, "gpg", "--verify", sigFn)
		cmd.Dir = sigDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", err
		}
	}

	sum := h.Sum(nil)
	return fmt.Sprintf("%x", sum), nil
}

func getUpdatedSignature(ctx context.Context, client *http.Client, update updater.Update, oldURL string) (string, error) {
	signatureRes, err := getUpdatedAsset(ctx, client, fmt.Sprintf("%s.asc", oldURL), update)
	if err != nil {
		return "", err
	}
	defer signatureRes.Body.Close()
	signatureBody, err := ioutil.ReadAll(signatureRes.Body)
	if err != nil {
		return "", err
	}

	signature := string(signatureBody)
	if !strings.Contains(signature, "-----BEGIN PGP SIGNATURE-----") {
		return "", nil
	}
	return signature, nil
}
