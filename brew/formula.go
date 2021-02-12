package brew

import (
	"regexp"

	"github.com/thepwagner/action-update/updater"
)

var (
	urlRe           = regexp.MustCompile(`url ["'](.*)["']`)
	shasumRe        = regexp.MustCompile(`sha(1|256) ["'](.*)["']`)
	versionVarRe    = regexp.MustCompile(`(?i)version\s*=?\s+["'](.*)["']`)
	versionTemplate = regexp.MustCompile(`(?i)#{version}`)
	semverRe        = regexp.MustCompile(`\d+\.\d+\.\d+`)
)

func parseFormulaDeps(formula string) ([]updater.Dependency, error) {
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

func parseFormulaHashes(formula string) (sums []string) {
	for _, m := range shasumRe.FindAllStringSubmatch(formula, -1) {
		sums = append(sums, m[2])
	}
	return
}
