package vaultsdk

import "github.com/ByteDeskAI/bytedesk-sdk-dependencies/semver"

// APIAtLeast reports whether a Vault-advertised apiVersion satisfies need (e.g. MinAPI).
func APIAtLeast(have, need string) bool {
	return semver.AtLeast(have, need)
}
