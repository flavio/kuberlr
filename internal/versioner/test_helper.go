package versioner

import (
	"os"
	"path/filepath"

	"github.com/blang/semver"
)

type kubectlNamer interface {
	ID() string
	Compute(semver.Version) string
}

type localKubectlNamer struct {
}

func (n *localKubectlNamer) ID() string {
	return "local"
}

func (n *localKubectlNamer) Compute(v semver.Version) string {
	return buildKubectlNameForLocalBin(v)
}

type systemKubectlNamer struct {
}

func (n *systemKubectlNamer) ID() string {
	return "system"
}

func (n *systemKubectlNamer) Compute(v semver.Version) string {
	return buildKubectlNameForSystemBin(v)
}
func fakeKubectlBinaries(path string, versions []string, nameBuilder kubectlNamer) KubectlBinaries {
	bins := KubectlBinaries{}

	for _, v := range versions {
		version := semver.MustParse(v)

		if nameBuilder.ID() == "system" {
			// system wide kubectl binaries do not provide the patch version
			// hence we consider them to be at path level 0
			version.Patch = 0
		}

		bin := filepath.Join(path, nameBuilder.Compute(version))

		bins = append(
			bins,
			KubectlBinary{
				Version: version,
				Path:    bin,
			})
	}

	return bins
}

func createFakeKubectlBinaries(bins KubectlBinaries) error {
	for _, bin := range bins {
		dir := filepath.Dir(bin.Path)
		_, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		file, err := os.Create(bin.Path)
		if err != nil {
			return err
		}
		if err = file.Close(); err != nil {
			return err
		}
	}
	return nil
}
