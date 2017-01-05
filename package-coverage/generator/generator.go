package generator

import "regexp"

// UnknownPackage ...
const UnknownPackage = "unknown"

// Coverage will generate coverage for the supplied directory and any sub-directories that contain Go files
func Coverage(basePath string, exclusionsMatcher *regexp.Regexp, quiet bool, goTestArgs []string) {
	processAllDirs(basePath, exclusionsMatcher, "coverage", func(path string) {
		generateCoverage(path, exclusionsMatcher, quiet, goTestArgs)
	})
}

// CoverageSingle will generate coverage for the supplied directory (and ignore all sub directories)
func CoverageSingle(basePath string, exclusionsMatcher *regexp.Regexp, quiet bool, goTestArgs []string) {
	generateCoverage(basePath, exclusionsMatcher, quiet, goTestArgs)
}
