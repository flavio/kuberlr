package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flavio/kuberlr/internal/logger"
)

type testData struct {
	FakeUsrEtc string
	FakeEtc    string
	FakeHome   string
}

func setup() (testData, error) {
	fakeUsrEtc, err := os.MkdirTemp("", "fake-usr-etc")
	if err != nil {
		return testData{}, err
	}

	fakeEtc, err := os.MkdirTemp("", "fake-etc")
	if err != nil {
		return testData{}, err
	}

	fakeHome, err := os.MkdirTemp("", "fake-home")
	if err != nil {
		return testData{}, err
	}

	return testData{FakeUsrEtc: fakeUsrEtc, FakeEtc: fakeEtc, FakeHome: fakeHome}, nil
}

func teardown(td testData) {
	os.RemoveAll(td.FakeUsrEtc)
	os.RemoveAll(td.FakeEtc)
	os.RemoveAll(td.FakeHome)
}

func writeConfig(path, data string) error {
	return os.WriteFile(
		filepath.Join(path, "kuberlr.conf"),
		[]byte(data),
		0o600)
}

func TestOnlySystemConfigExists(t *testing.T) {
	t.Parallel()

	td, err := setup()
	if err != nil {
		t.Error(err)
	}
	defer teardown(td)

	err = writeConfig(td.FakeUsrEtc, "AllowDownload = false")
	if err != nil {
		t.Error(err)
	}

	c := Cfg{
		Paths: []string{
			filepath.Join(td.FakeUsrEtc, "kuberlr.conf"),
			filepath.Join(td.FakeEtc, "kuberlr.conf"),
			filepath.Join(td.FakeHome, "kuberlr.conf"),
		},
	}

	v, err := c.Load()
	if err != nil {
		t.Errorf("Unexpected error loading config: %v", err)
	}
	if v.GetBool("AllowDownload") != false {
		t.Error("Expected configuration value wasn't found")
	}
}

func TestHomeConfigOverridesSystemOne(t *testing.T) {
	t.Parallel()

	td, err := setup()
	if err != nil {
		t.Error(err)
	}
	defer teardown(td)

	err = writeConfig(td.FakeUsrEtc, "AllowDownload = false")
	if err != nil {
		t.Error(err)
	}
	err = writeConfig(td.FakeHome, "AllowDownload = true")
	if err != nil {
		t.Error(err)
	}

	c := Cfg{
		Paths: []string{
			filepath.Join(td.FakeUsrEtc, "kuberlr.conf"),
			filepath.Join(td.FakeEtc, "kuberlr.conf"),
			filepath.Join(td.FakeHome, "kuberlr.conf"),
		},
	}

	v, err := c.Load()
	if err != nil {
		t.Errorf("Unexpected error loading config: %v", err)
	}
	if v.GetBool("AllowDownload") != true {
		t.Error("Expected configuration value wasn't found")
	}
}

func TestEnvironmentVariables(t *testing.T) {
	td, err := setup()
	if err != nil {
		t.Error(err)
	}
	defer teardown(td)

	err = writeConfig(td.FakeHome, "AllowDownload = false")
	if err != nil {
		t.Error(err)
	}

	t.Setenv("KUBERLR_ALLOWDOWNLOAD", "true")

	c := Cfg{
		Paths: []string{},
	}

	v, err := c.Load()
	if err != nil {
		t.Errorf("Unexpected error loading config: %v", err)
	}
	if v.GetBool("AllowDownload") != true {
		t.Error("env var wasn't taken into account")
	}
}

func TestMergeConfigs(t *testing.T) {
	t.Parallel()

	td, err := setup()
	if err != nil {
		t.Error(err)
	}
	defer teardown(td)

	usrEtcCfg := `
AllowDownload = false
SystemPath = "global"
Timeout = 2
`
	err = writeConfig(td.FakeUsrEtc, usrEtcCfg)
	if err != nil {
		t.Error(err)
	}

	etcCfg := `
Timeout = 200
`
	err = writeConfig(td.FakeEtc, etcCfg)
	if err != nil {
		t.Error(err)
	}

	homeCfg := `
AllowDownload = true
`
	err = writeConfig(td.FakeHome, homeCfg)
	if err != nil {
		t.Error(err)
	}

	c := Cfg{
		Paths: []string{
			filepath.Join(td.FakeUsrEtc, "kuberlr.conf"),
			filepath.Join(td.FakeEtc, "kuberlr.conf"),
			filepath.Join(td.FakeHome, "kuberlr.conf"),
			os.Getenv("__FAKE_ENV__"),
		},
	}

	v, err := c.Load()
	if err != nil {
		t.Errorf("Unexpected error loading config: %v", err)
	}

	if v.GetBool("AllowDownload") != true {
		t.Errorf(
			"Wrong value for AllowDownload: got %v instead of %v",
			v.GetBool("AllowDownload"), true)
	}

	if v.GetInt64("Timeout") != 200 {
		t.Errorf(
			"Wrong value for Timeout: got %v instead of %v",
			v.GetInt64("Timeout"), 200)
	}

	if v.GetString("SystemPath") != "global" {
		t.Errorf(
			"Wrong value for Timeout: got %v instead of %v",
			v.GetString("SystemPath"), "global")
	}
}

func TestLoggerOptionsDefaults(t *testing.T) {
	t.Parallel()

	c := Cfg{Paths: []string{}}
	v, err := c.Load()
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}

	opts, err := LoggerOptions(v)
	if err != nil {
		t.Fatalf("Unexpected error building logger options: %v", err)
	}

	want := logger.Options{Verbosity: logger.VerbosityDefault, Quiet: false, Color: logger.ColorAuto}
	if opts != want {
		t.Errorf("Wrong UI options: got %+v instead of %+v", opts, want)
	}
}

func TestLoggerOptionsFromFile(t *testing.T) {
	t.Parallel()

	td, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer teardown(td)

	err = writeConfig(td.FakeHome, `
Verbosity = 2
Quiet = true
Color = "never"
`)
	if err != nil {
		t.Fatal(err)
	}

	c := Cfg{Paths: []string{filepath.Join(td.FakeHome, "kuberlr.conf")}}
	v, err := c.Load()
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}

	opts, err := LoggerOptions(v)
	if err != nil {
		t.Fatalf("Unexpected error building logger options: %v", err)
	}

	want := logger.Options{Verbosity: 2, Quiet: true, Color: logger.ColorNever}
	if opts != want {
		t.Errorf("Wrong UI options: got %+v instead of %+v", opts, want)
	}
}

func TestLoggerOptionsFromEnvironment(t *testing.T) {
	t.Setenv("KUBERLR_VERBOSITY", "1")
	t.Setenv("KUBERLR_QUIET", "true")
	t.Setenv("KUBERLR_COLOR", "always")

	c := Cfg{Paths: []string{}}
	v, err := c.Load()
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}

	opts, err := LoggerOptions(v)
	if err != nil {
		t.Fatalf("Unexpected error building logger options: %v", err)
	}

	want := logger.Options{Verbosity: 1, Quiet: true, Color: logger.ColorAlways}
	if opts != want {
		t.Errorf("Wrong UI options: got %+v instead of %+v", opts, want)
	}
}

func TestLoggerOptionsInvalidColor(t *testing.T) {
	t.Setenv("KUBERLR_COLOR", "rainbow")

	c := Cfg{Paths: []string{}}
	v, err := c.Load()
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}

	opts, err := LoggerOptions(v)
	if err == nil {
		t.Fatal("Expected an error for an invalid Color value")
	}
	if opts.Color != logger.ColorAuto {
		t.Errorf("Expected fallback to %q, got %q", logger.ColorAuto, opts.Color)
	}
}
