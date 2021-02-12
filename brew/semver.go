package brew

import (
	"fmt"
	"sort"
	"strings"

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
