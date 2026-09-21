package crayfi

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

type SigningKeyPair struct{ PrivateKeyPEM, PublicKeyPEM, Fingerprint string }

func GenerateSigningKeyPair() (*SigningKeyPair, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	privateBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	publicBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}
	privatePEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateBytes}))
	publicPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicBytes}))
	return &SigningKeyPair{privatePEM, publicPEM, Fingerprint(publicPEM)}, nil
}

func Fingerprint(publicKeyPEM string) string {
	hash := sha256.Sum256([]byte(publicKeyPEM))
	return hex.EncodeToString(hash[:])
}

func parseSigningPrivateKey(value string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		return nil, fmt.Errorf("request signing private key must be PEM")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, parseErr := x509.ParseECPrivateKey(block.Bytes)
		if parseErr != nil {
			return nil, parseErr
		}
		if key.Curve != elliptic.P256() {
			return nil, fmt.Errorf("request signing private key must be an ECDSA P-256 key")
		}
		return key, nil
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok || key.Curve != elliptic.P256() {
		return nil, fmt.Errorf("request signing private key must be an ECDSA P-256 key")
	}
	return key, nil
}

func SignChallenge(challenge, privateKeyPEM string) (string, error) {
	key, err := parseSigningPrivateKey(privateKeyPEM)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(challenge))
	signature, err := ecdsa.SignASN1(rand.Reader, key, hash[:])
	return base64.StdEncoding.EncodeToString(signature), err
}
func BuildSignatureHeader(method, path string, body []byte, privateKeyPEM, keyID, timestamp, nonce string) (string, error) {
	if timestamp == "" {
		timestamp = fmt.Sprintf("%d", time.Now().UnixMilli())
	}
	if nonce == "" {
		nonce = uuidV4()
	}
	key, err := parseSigningPrivateKey(privateKeyPEM)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(body)
	path = strings.TrimRight(path, "/")
	if path == "" {
		path = "/"
	}
	message := strings.Join([]string{strings.ToUpper(method), path, timestamp, nonce, hex.EncodeToString(hash[:])}, "\n")
	digest := sha256.Sum256([]byte(message))
	signature, err := ecdsa.SignASN1(rand.Reader, key, digest[:])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("t=%s,n=%s,kid=%s,v1=%s", timestamp, nonce, keyID, base64.StdEncoding.EncodeToString(signature)), nil
}
func uuidV4() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}
func canonicalQuery(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, rfc3986(key)+"="+rfc3986(params[key]))
	}
	return strings.Join(parts, "&")
}
func rfc3986(value string) string { return strings.ReplaceAll(url.QueryEscape(value), "+", "%20") }
