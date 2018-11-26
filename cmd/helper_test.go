package cmd

import "testing"

var RepeatableFlagSet = defaultRepeatableFlagSet

func NormalizePath(path string) (repo, ver string, err error) {
	return normalizePath(path)
}

func AssertPath(t *testing.T, val, repo, ver, out string, actual *path) {
	if val != actual.val {
		t.Errorf("val is wrong: expected = %s, actual = %s", val, actual.val)
	}
	if repo != actual.repo {
		t.Errorf("repo is wrong: expected = %s, actual = %s", repo, actual.repo)
	}
	if ver != actual.ver {
		t.Errorf("ver is wrong: expected = %s, actual = %s", ver, actual.ver)
	}
	if out != actual.out {
		t.Errorf("out is wrong: expected = %s, actual = %s", out, actual.out)
	}
}
