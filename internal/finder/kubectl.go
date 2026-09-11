package finder

import (
	"sort"

	"github.com/blang/semver/v4"
)

// KubectlBinary describes a kubectl binary.
type KubectlBinary struct {
	Path    string
	Version semver.Version
}

// KubectlBinaries is a list of KubectlBinary objects.
type KubectlBinaries []KubectlBinary

// SortKubectlByVersion sorts a list of KubectlBinary objects using their version
// attribute. By default objects are sorted ascendantly (from earlier to more
// recent versions); this can be changed via the `reverse` parameter. When two
// binaries have the same version, their path is used as a tiebreaker, so the
// order is always deterministic.
func SortKubectlByVersion(binaries KubectlBinaries, reverse bool) {
	sort.Slice(binaries, func(i, j int) bool {
		vi, vj := binaries[i].Version, binaries[j].Version
		if vi.Equals(vj) {
			return binaries[i].Path < binaries[j].Path
		}
		if reverse {
			return vi.GT(vj)
		}
		return vi.LT(vj)
	})
}
