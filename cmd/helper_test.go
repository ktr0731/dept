package cmd

func NormalizePath(path string) (repo, ver string, err error) {
	return normalizePath(path)
}
