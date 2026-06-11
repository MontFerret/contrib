package core

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"hash/fnv"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type keyFamily int

const (
	familyHMAC keyFamily = iota
	familyRSA
	familyECDSA
	familyEd25519
)

// HMACKey is an opaque HMAC secret exposed to Ferret.
type HMACKey struct {
	secret []byte
	id     uint64
}

// PublicKey is an opaque public key exposed to Ferret.
type PublicKey struct {
	rsaKey *rsa.PublicKey
	ecKey  *ecdsa.PublicKey
	edKey  ed25519.PublicKey
	family keyFamily
	id     uint64
}

// PrivateKey is an opaque private key exposed to Ferret.
type PrivateKey struct {
	rsaKey *rsa.PrivateKey
	ecKey  *ecdsa.PrivateKey
	edKey  ed25519.PrivateKey
	family keyFamily
	id     uint64
}

// NewHMACKey creates an opaque HMAC key from secret material.
func NewHMACKey(secret []byte) *HMACKey {
	return &HMACKey{
		secret: append([]byte(nil), secret...),
		id:     newResourceID(),
	}
}

// NewPublicKey parses a PEM-encoded public key.
func NewPublicKey(pemText string) (*PublicKey, error) {
	key, family, err := parsePublicKeyPEM(pemText)
	if err != nil {
		return nil, err
	}

	out := &PublicKey{
		family: family,
		id:     newResourceID(),
	}

	switch family {
	case familyRSA:
		out.rsaKey = key.(*rsa.PublicKey)
	case familyECDSA:
		out.ecKey = key.(*ecdsa.PublicKey)
	case familyEd25519:
		out.edKey = key.(ed25519.PublicKey)
	}

	return out, nil
}

// NewPrivateKey parses a PEM-encoded private key.
func NewPrivateKey(pemText string) (*PrivateKey, error) {
	key, family, err := parsePrivateKeyPEM(pemText)
	if err != nil {
		return nil, err
	}

	out := &PrivateKey{
		family: family,
		id:     newResourceID(),
	}

	switch family {
	case familyRSA:
		out.rsaKey = key.(*rsa.PrivateKey)
	case familyECDSA:
		out.ecKey = key.(*ecdsa.PrivateKey)
	case familyEd25519:
		out.edKey = key.(ed25519.PrivateKey)
	}

	return out, nil
}

func parsePublicKeyPEM(pemText string) (any, keyFamily, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(pemText)))
	if block == nil {
		return nil, 0, newError(ErrInvalidKey, "invalid PEM public key")
	}

	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		return classifyPublicKey(key)
	}

	if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		return classifyPublicKey(cert.PublicKey)
	}

	return nil, 0, newError(ErrInvalidKey, "invalid PEM public key")
}

func parsePrivateKeyPEM(pemText string) (any, keyFamily, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(pemText)))
	if block == nil {
		return nil, 0, newError(ErrInvalidKey, "invalid PEM private key")
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return classifyPrivateKey(key)
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, familyRSA, nil
	}

	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, familyECDSA, nil
	}

	return nil, 0, newError(ErrInvalidKey, "invalid PEM private key")
}

func classifyPublicKey(key any) (any, keyFamily, error) {
	switch typed := key.(type) {
	case *rsa.PublicKey:
		return typed, familyRSA, nil
	case *ecdsa.PublicKey:
		return typed, familyECDSA, nil
	case ed25519.PublicKey:
		return typed, familyEd25519, nil
	default:
		return nil, 0, newError(ErrInvalidKey, "unsupported public key type")
	}
}

func classifyPrivateKey(key any) (any, keyFamily, error) {
	switch typed := key.(type) {
	case *rsa.PrivateKey:
		return typed, familyRSA, nil
	case *ecdsa.PrivateKey:
		return typed, familyECDSA, nil
	case ed25519.PrivateKey:
		return typed, familyEd25519, nil
	default:
		return nil, 0, newError(ErrInvalidKey, "unsupported private key type")
	}
}

func algorithmFamily(alg string) (keyFamily, error) {
	switch {
	case strings.HasPrefix(alg, "HS"):
		return familyHMAC, nil
	case strings.HasPrefix(alg, "RS"), strings.HasPrefix(alg, "PS"):
		return familyRSA, nil
	case strings.HasPrefix(alg, "ES"):
		return familyECDSA, nil
	case alg == "EdDSA":
		return familyEd25519, nil
	default:
		return 0, newError(ErrUnsupportedAlgorithm, "unsupported JWT algorithm")
	}
}

func verifyKeyMatchesAlgorithm(key any, alg string) error {
	expected, err := algorithmFamily(alg)
	if err != nil {
		return err
	}

	switch expected {
	case familyHMAC:
		if _, ok := key.(*HMACKey); !ok {
			return newError(ErrInvalidKey, "algorithm requires an HMAC key")
		}
	case familyRSA:
		switch key.(type) {
		case *PublicKey, *PrivateKey:
			if pub, ok := key.(*PublicKey); ok && pub.family != familyRSA {
				return newError(ErrInvalidKey, "algorithm requires an RSA key")
			}
			if priv, ok := key.(*PrivateKey); ok && priv.family != familyRSA {
				return newError(ErrInvalidKey, "algorithm requires an RSA key")
			}
		default:
			return newError(ErrInvalidKey, "algorithm requires an RSA key")
		}
	case familyECDSA:
		switch key.(type) {
		case *PublicKey, *PrivateKey:
			if pub, ok := key.(*PublicKey); ok && pub.family != familyECDSA {
				return newError(ErrInvalidKey, "algorithm requires an ECDSA key")
			}
			if priv, ok := key.(*PrivateKey); ok && priv.family != familyECDSA {
				return newError(ErrInvalidKey, "algorithm requires an ECDSA key")
			}
		default:
			return newError(ErrInvalidKey, "algorithm requires an ECDSA key")
		}
	case familyEd25519:
		switch key.(type) {
		case *PublicKey, *PrivateKey:
			if pub, ok := key.(*PublicKey); ok && pub.family != familyEd25519 {
				return newError(ErrInvalidKey, "algorithm requires an Ed25519 key")
			}
			if priv, ok := key.(*PrivateKey); ok && priv.family != familyEd25519 {
				return newError(ErrInvalidKey, "algorithm requires an Ed25519 key")
			}
		default:
			return newError(ErrInvalidKey, "algorithm requires an Ed25519 key")
		}
	}

	return nil
}

func (k *HMACKey) Secret() []byte {
	return k.secret
}

func (k *HMACKey) ResourceID() uint64 { return k.id }
func (k *HMACKey) String() string     { return "<security.jwt.hmac_key>" }
func (k *HMACKey) Copy() runtime.Value {
	return k
}
func (k *HMACKey) MarshalJSON() ([]byte, error) {
	return []byte(`"<security.jwt.hmac_key>"`), nil
}
func (k *HMACKey) Hash() uint64 { return opaqueHash("security.jwt.hmac_key", k.id) }

func (k *PublicKey) ResourceID() uint64 { return k.id }
func (k *PublicKey) String() string     { return "<security.jwt.public_key>" }
func (k *PublicKey) Copy() runtime.Value {
	return k
}
func (k *PublicKey) MarshalJSON() ([]byte, error) {
	return []byte(`"<security.jwt.public_key>"`), nil
}
func (k *PublicKey) Hash() uint64 { return opaqueHash("security.jwt.public_key", k.id) }

func (k *PrivateKey) ResourceID() uint64 { return k.id }
func (k *PrivateKey) String() string     { return "<security.jwt.private_key>" }
func (k *PrivateKey) Copy() runtime.Value {
	return k
}
func (k *PrivateKey) MarshalJSON() ([]byte, error) {
	return []byte(`"<security.jwt.private_key>"`), nil
}
func (k *PrivateKey) Hash() uint64 { return opaqueHash("security.jwt.private_key", k.id) }

func opaqueHash(prefix string, id uint64) uint64 {
	h := fnv.New64a()
	h.Write([]byte(prefix + ":"))

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, id)
	h.Write(bytes)

	return h.Sum64()
}
