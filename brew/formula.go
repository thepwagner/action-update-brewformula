package brew

import (
	"io/ioutil"
	"regexp"

	"github.com/thepwagner/action-update/updater"
)

var (
	urlRe           = regexp.MustCompile(`url ["'](.*)["']`)
	versionVarRe    = regexp.MustCompile(`(?i)version\s*=?\s+["'](.*)["']`)
	versionTemplate = regexp.MustCompile(`(?i)#{version}`)
	semverRe        = regexp.MustCompile(`\d+\.\d+\.\d+`)
)

func parseFormula(formulaPath string) ([]updater.Dependency, error) {
	formulaBytes, err := ioutil.ReadFile(formulaPath)
	if err != nil {
		return nil, err
	}
	formula := string(formulaBytes)

	// Find `VERSION =` variables and `url` string:
	versionVar := versionVarRe.FindStringSubmatch(formula)
	urlMatch := urlRe.FindAllStringSubmatch(formula, -1)
	if len(urlMatch) != 1 {
		return nil, nil
	}
	formulaURL := urlMatch[0][1]

	// https://foo.com/awesome-#{VERSION}.tar.gz
	if len(versionVar) > 0 && versionTemplate.MatchString(formulaURL) {
		return []updater.Dependency{{
			Path:    formulaURL,
			Version: versionVar[1],
		}}, nil
	}

	if version := semverRe.FindString(formulaURL); version != "" {
		return []updater.Dependency{{
			Path:    formulaURL,
			Version: version,
		}}, nil
	}

	return nil, nil
}
