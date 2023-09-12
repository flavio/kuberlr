package config

import (
	"os"
	"path/filepath"
	"testing"
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
		0600)
}

func TestOnlySystemConfigExists(t *testing.T) {
	//nolint: varnamelen
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
		Paths: []string{td.FakeUsrEtc, td.FakeEtc, td.FakeHome},
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
	//nolint: varnamelen
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
		Paths: []string{td.FakeUsrEtc, td.FakeEtc, td.FakeHome},
	}

	v, err := c.Load()
	if err != nil {
		t.Errorf("Unexpected error loading config: %v", err)
	}
	if v.GetBool("AllowDownload") != true {
		t.Error("Expected configuration value wasn't found")
	}
}

func TestMergeConfigs(t *testing.T) {
	//nolint: varnamelen
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
		Paths: []string{td.FakeUsrEtc, td.FakeEtc, td.FakeHome},
	}

	//nolint: varnamelen
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
