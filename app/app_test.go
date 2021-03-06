package app

import (
	"os/exec"
	"strings"
	"testing"

	version "github.com/hashicorp/go-version"
)

func Test_version(t *testing.T) {
	out, err := exec.Command("git", "tag").Output()
	if err != nil {
		t.Fatalf("failed to get Git tags: %s", err)
	}
	sp := strings.Split(strings.TrimSpace(string(out)), "\n")
	cmdVer := sp[len(sp)-1]
	appVer := "v" + appVersion
	if appVer != cmdVer {
		t.Errorf("app: %s, but Git tag: %s", appVer, cmdVer)
	}
}

func Test_isCompatibleGoVersion(t *testing.T) {
	cases := []struct {
		version    string
		compatible bool
	}{
		{version: "1.10.0", compatible: false},
		{version: "1.11.0", compatible: true},
		{version: "1.20.0", compatible: true},
	}

	for _, c := range cases {
		ver := version.Must(version.NewVersion(c.version))
		if c.compatible != isCompatibleGoVersion(ver) {
			t.Error("isCompatibleGoVersion must be return same boolean value as c.compatible")
		}
	}
}

func Test_isLimitedGoModSupport(t *testing.T) {
	cases := []struct {
		version   string
		limited   bool
		willPanic bool
	}{
		{version: "1.10.0", willPanic: true},

		{version: "1.11.0", limited: true},
		{version: "1.11.1", limited: true},
		{version: "1.12.0", limited: false},
	}

	for _, c := range cases {
		t.Run(c.version, func(t *testing.T) {
			ver := version.Must(version.NewVersion(c.version))
			if c.willPanic {
				defer func() {
					err := recover()
					if err == nil {
						t.Error("must panic because incompatible version passed")
					}
				}()
			}

			if c.limited != isLimitedGoModSupport(ver) {
				t.Error("isLimitedGoModSupport must be return same boolean value as c.limited")
			}
		})
	}
}
