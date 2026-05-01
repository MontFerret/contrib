package useragent

import "github.com/corpix/uarand"

const Random = "*"

func Resolve(value string) string {
	if value == "" {
		return value
	}

	if value != Random {
		return value
	}

	// TODO: Replace the dependency with a repository-local policy.
	return uarand.GetRandom()
}
