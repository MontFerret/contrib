package core

import "time"

type (
	// Parameters contains provider-specific form fields. Each value is emitted as
	// a separate form field, preserving repeated parameters.
	Parameters map[string][]string

	// ClientCredentialsOptions configures a client credentials grant.
	ClientCredentialsOptions struct {
		Parameters Parameters
		Audience   string
		Scope      []string
		Timeout    time.Duration
	}

	// RefreshOptions configures a refresh token grant.
	RefreshOptions struct {
		Parameters   Parameters
		RefreshToken string
		Scope        []string
		Timeout      time.Duration
	}

	// TokenOptions configures an OAuth extension grant.
	TokenOptions struct {
		Parameters Parameters
		Timeout    time.Duration
	}
)
