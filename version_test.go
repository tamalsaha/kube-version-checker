package main

import (
	"testing"

	"github.com/appscode/go-version"
)

func TestDefaultSupportedVersion(t *testing.T) {
	cases := []struct {
		version     string
		multiMaster bool
		err         bool
	}{
		{"1.10.0", true, false},
	}

	for _, tc := range cases {
		v, err := version.NewVersion(tc.version)
		if err != nil {
			t.Fatalf("failed parse version for input %s: %s", tc.version, err)
		}

		err = checkVersion(
			v,
			tc.multiMaster,
			defaultConstraint,
			defaultBlackListedVersions,
			defaultBlackListedMultiMasterVersions)
		if tc.err && err == nil {
			t.Fatalf("expected error for input: %s", tc.version)
		} else if !tc.err && err != nil {
			t.Fatalf("error for input %s: %s", tc.version, err)
		}
	}
}
