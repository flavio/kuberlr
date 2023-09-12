package finder

import (
	"testing"

	"github.com/blang/semver/v4"
)

func TestSortAsc(t *testing.T) {
	bin1 := KubectlBinary{
		Path:    "b1",
		Version: semver.MustParse("1.0.0"),
	}

	bin2 := KubectlBinary{
		Path:    "b2",
		Version: semver.MustParse("2.0.0"),
	}

	bin3 := KubectlBinary{
		Path:    "b3",
		Version: semver.MustParse("2.0.3"),
	}

	expected := KubectlBinaries{bin1, bin2, bin3}
	actual := KubectlBinaries{bin3, bin1, bin2}

	SortKubectlByVersion(actual, false)

	for i, e := range expected {
		if actual[i].Path != e.Path {
			t.Errorf("Got %+v instead of %+v", actual[i].Version, e.Version)
		}
	}
}

func TestSortDesc(t *testing.T) {
	bin1 := KubectlBinary{
		Path:    "b1",
		Version: semver.MustParse("1.0.0"),
	}

	bin2 := KubectlBinary{
		Path:    "b2",
		Version: semver.MustParse("2.0.0"),
	}

	bin3 := KubectlBinary{
		Path:    "b3",
		Version: semver.MustParse("2.0.3"),
	}

	expected := KubectlBinaries{bin3, bin2, bin1}
	actual := KubectlBinaries{bin3, bin1, bin2}

	SortKubectlByVersion(actual, true)

	for i, e := range expected {
		if actual[i].Path != e.Path {
			t.Errorf("Got %+v instead of %+v", actual[i].Version, e.Version)
		}
	}
}
