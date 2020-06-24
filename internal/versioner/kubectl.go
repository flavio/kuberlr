package versioner

import (
	"sort"

	"github.com/blang/semver"
)

type KubectlBinary struct {
	Path    string
	Version semver.Version
}

type KubectlBinaries []KubectlBinary

func SortByVersion(binaries KubectlBinaries, reverse bool) {
	sort.Slice(binaries, func(i, j int) bool {
		if reverse {
			return binaries[i].Version.GT(binaries[j].Version)
		} else {
			return binaries[i].Version.LT(binaries[j].Version)
		}
	})
}
