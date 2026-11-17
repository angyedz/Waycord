package ota

import (
	"fmt"
	"strconv"
	"strings"
)

type parts struct {
	numbers [3]uint64
	stage   string
	serial  uint64
}

func parseVersion(s string) (parts, error) {
	if !validVersion.MatchString(s) {
		return parts{}, fmt.Errorf("invalid version %q", s)
	}
	var v parts
	segments := strings.SplitN(strings.TrimPrefix(s, "v"), "-", 2)
	for i, n := range strings.Split(segments[0], ".") {
		value, err := strconv.ParseUint(n, 10, 64)
		if err != nil {
			return parts{}, err
		}
		v.numbers[i] = value
	}
	if len(segments) == 2 {
		suffix := strings.Split(segments[1], ".")
		v.stage = suffix[0]
		var err error
		v.serial, err = strconv.ParseUint(suffix[1], 10, 64)
		if err != nil {
			return parts{}, fmt.Errorf("invalid prerelease serial: %w", err)
		}
	}
	return v, nil
}

func rank(s string) int {
	switch s {
	case "nightly":
		return 0
	case "beta":
		return 1
	default:
		return 2
	}
}

// IsNewer compares the supported version format without lexicographic mistakes.
func IsNewer(candidate, installed string) (bool, error) {
	a, err := parseVersion(candidate)
	if err != nil {
		return false, err
	}
	b, err := parseVersion(installed)
	if err != nil {
		return false, err
	}
	for i := range a.numbers {
		if a.numbers[i] != b.numbers[i] {
			return a.numbers[i] > b.numbers[i], nil
		}
	}
	if rank(a.stage) != rank(b.stage) {
		return rank(a.stage) > rank(b.stage), nil
	}
	return a.serial > b.serial, nil
}
