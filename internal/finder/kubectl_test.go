package finder

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
)

func TestSortAsc(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

func TestSortUsesSemverNotStringOrder(t *testing.T) {
	t.Parallel()

	bin9 := KubectlBinary{Path: "b9", Version: semver.MustParse("1.9.0")}
	bin10 := KubectlBinary{Path: "b10", Version: semver.MustParse("1.10.0")}
	bin28 := KubectlBinary{Path: "b28", Version: semver.MustParse("1.28.0")}

	ascending := KubectlBinaries{bin28, bin9, bin10}
	SortKubectlByVersion(ascending, false)
	assert.Equal(t, KubectlBinaries{bin9, bin10, bin28}, ascending)

	descending := KubectlBinaries{bin9, bin28, bin10}
	SortKubectlByVersion(descending, true)
	assert.Equal(t, KubectlBinaries{bin28, bin10, bin9}, descending)
}

func TestSortBreaksVersionTiesByPath(t *testing.T) {
	t.Parallel()

	binA := KubectlBinary{Path: "/a/kubectl1.28.0", Version: semver.MustParse("1.28.0")}
	binB := KubectlBinary{Path: "/b/kubectl1.28.0", Version: semver.MustParse("1.28.0")}

	ascending := KubectlBinaries{binB, binA}
	SortKubectlByVersion(ascending, false)
	assert.Equal(t, KubectlBinaries{binA, binB}, ascending)

	descending := KubectlBinaries{binB, binA}
	SortKubectlByVersion(descending, true)
	assert.Equal(t, KubectlBinaries{binA, binB}, descending)
}
