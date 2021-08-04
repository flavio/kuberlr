package finder

import (
	"testing"

	"github.com/blang/semver/v4"
)

func TestSortAsc(t *testing.T) {
	b1 := KubectlBinary{
		Path:    "b1",
		Version: semver.MustParse("1.0.0"),
	}

	b2 := KubectlBinary{
		Path:    "b2",
		Version: semver.MustParse("2.0.0"),
	}

	b3 := KubectlBinary{
		Path:    "b3",
		Version: semver.MustParse("2.0.3"),
	}

	expected := KubectlBinaries{b1, b2, b3}
	actual := KubectlBinaries{b3, b1, b2}

	SortKubectlByVersion(actual, false)

	for i, e := range expected {
		if actual[i].Path != e.Path {
			t.Errorf("Got %+v instead of %+v", actual[i].Version, e.Version)
		}
	}
}

func TestSortDesc(t *testing.T) {
	b1 := KubectlBinary{
		Path:    "b1",
		Version: semver.MustParse("1.0.0"),
	}

	b2 := KubectlBinary{
		Path:    "b2",
		Version: semver.MustParse("2.0.0"),
	}

	b3 := KubectlBinary{
		Path:    "b3",
		Version: semver.MustParse("2.0.3"),
	}

	expected := KubectlBinaries{b3, b2, b1}
	actual := KubectlBinaries{b3, b1, b2}

	SortKubectlByVersion(actual, true)

	for i, e := range expected {
		if actual[i].Path != e.Path {
			t.Errorf("Got %+v instead of %+v", actual[i].Version, e.Version)
		}
	}
}
