package core

// ClientAuthMethod identifies how an OAuth client authenticates at the token
// endpoint.
type ClientAuthMethod string

const (
	// ClientAuthMethodNone sends no client secret.
	ClientAuthMethodNone ClientAuthMethod = "none"
	// ClientAuthMethodBasic sends the client credentials in an Authorization
	// header using the client_secret_basic method.
	ClientAuthMethodBasic ClientAuthMethod = "client_secret_basic"
	// ClientAuthMethodPost sends the client credentials in the form body using
	// the client_secret_post method.
	ClientAuthMethodPost ClientAuthMethod = "client_secret_post"
)

func (m ClientAuthMethod) supported() bool {
	switch m {
	case ClientAuthMethodNone, ClientAuthMethodBasic, ClientAuthMethodPost:
		return true
	default:
		return false
	}
}
