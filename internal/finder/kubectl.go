package finder

import (
	"sort"

	"github.com/blang/semver"
)

// KubectlBinary describes a kubectl binary
type KubectlBinary struct {
	Path    string
	Version semver.Version
}

// KubectlBinaries is a list of KubectlBinary objects
type KubectlBinaries []KubectlBinary

// SortKubectlByVersion sorts a list of KubectlBinary objects using their version
// attribute. By default objects are sorted ascendantly (from earlier to more
// recent versions); this can be changed via the `reverse` parameter
func SortKubectlByVersion(binaries KubectlBinaries, reverse bool) {
	sort.Slice(binaries, func(i, j int) bool {
		if reverse {
			return binaries[i].Version.GT(binaries[j].Version)
		}
		return binaries[i].Version.LT(binaries[j].Version)
	})
}
