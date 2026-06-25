package core

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const defaultPort = 5432

// OpenOptions configures Postgres database connections.
type OpenOptions struct {
	URI      *string `json:"uri"`
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Database *string `json:"database"`
	User     *string `json:"user"`
	Password *string `json:"password"`
	SSLMode  *string `json:"sslMode"`
}

func (o OpenOptions) dsn() (string, error) {
	uriProvided := o.uriProvided()
	fieldsProvided := o.structuredFieldsProvided()

	if uriProvided == fieldsProvided {
		return "", fmt.Errorf("exactly one of uri or structured connection fields must be provided")
	}

	if uriProvided {
		return strings.TrimSpace(*o.URI), nil
	}

	if !o.hostProvided() || !o.databaseProvided() || !o.userProvided() {
		return "", fmt.Errorf("host, database, and user are required when uri is not provided")
	}

	port := defaultPort
	if o.Port != nil {
		if *o.Port <= 0 {
			return "", fmt.Errorf("port must be greater than 0")
		}

		port = *o.Port
	}

	out := url.URL{
		Scheme: "postgres",
		User:   url.User(strings.TrimSpace(*o.User)),
		Host:   net.JoinHostPort(strings.TrimSpace(*o.Host), strconv.Itoa(port)),
		Path:   "/" + strings.TrimSpace(*o.Database),
	}

	if o.passwordProvided() {
		out.User = url.UserPassword(strings.TrimSpace(*o.User), *o.Password)
	}

	if o.sslModeProvided() {
		query := out.Query()
		query.Set("sslmode", strings.TrimSpace(*o.SSLMode))
		out.RawQuery = query.Encode()
	}

	return out.String(), nil
}

func (o OpenOptions) uriProvided() bool {
	return o.URI != nil && strings.TrimSpace(*o.URI) != ""
}

func (o OpenOptions) structuredFieldsProvided() bool {
	return o.hostProvided() ||
		o.databaseProvided() ||
		o.userProvided() ||
		o.Port != nil ||
		o.passwordProvided() ||
		o.sslModeProvided()
}

func (o OpenOptions) hostProvided() bool {
	return o.Host != nil && strings.TrimSpace(*o.Host) != ""
}

func (o OpenOptions) databaseProvided() bool {
	return o.Database != nil && strings.TrimSpace(*o.Database) != ""
}

func (o OpenOptions) userProvided() bool {
	return o.User != nil && strings.TrimSpace(*o.User) != ""
}

func (o OpenOptions) passwordProvided() bool {
	return o.Password != nil
}

func (o OpenOptions) sslModeProvided() bool {
	return o.SSLMode != nil && strings.TrimSpace(*o.SSLMode) != ""
}
