package crayfi

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type signingFixture struct {
	PrivateKeyPEM string `json:"private_key_pem"`
	PublicKeyPEM  string `json:"public_key_pem"`
	KeyID         string `json:"key_id"`
	T             string `json:"t"`
	N             string `json:"n"`
	Challenge     string `json:"challenge"`
	Cases         []struct {
		Name, Method, Path, Body string
		Query                    map[string]string `json:"query"`
		SigningString            string            `json:"signing_string"`
	} `json:"cases"`
}

func loadFixture(t *testing.T) signingFixture {
	t.Helper()
	bytes, err := os.ReadFile("testdata/signing-fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture signingFixture
	if err = json.Unmarshal(bytes, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}
func publicKey(t *testing.T, pemText string) *ecdsa.PublicKey {
	t.Helper()
	block, _ := pem.Decode([]byte(pemText))
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return parsed.(*ecdsa.PublicKey)
}
func TestSigningFixture(t *testing.T) {
	fixture := loadFixture(t)
	key := publicKey(t, fixture.PublicKeyPEM)
	for _, testCase := range fixture.Cases {
		path := strings.TrimRight(testCase.Path, "/")
		if path == "" {
			path = "/"
		}
		if len(testCase.Query) > 0 {
			path += "?" + canonicalQuery(testCase.Query)
		}
		bodyHash := sha256.Sum256([]byte(testCase.Body))
		expected := strings.Join([]string{strings.ToUpper(testCase.Method), path, fixture.T, fixture.N, hex.EncodeToString(bodyHash[:])}, "\n")
		if expected != testCase.SigningString {
			t.Fatalf("unexpected signing string for %s: %q", testCase.Name, expected)
		}
		header, err := BuildSignatureHeader(testCase.Method, path, []byte(testCase.Body), fixture.PrivateKeyPEM, fixture.KeyID, fixture.T, fixture.N)
		if err != nil {
			t.Fatal(err)
		}
		parts := strings.Split(header, "v1=")
		signature, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256([]byte(testCase.SigningString))
		if !ecdsa.VerifyASN1(key, hash[:], signature) {
			t.Fatalf("fixture signature failed for %s", testCase.Name)
		}
	}
	signature, err := SignChallenge(fixture.Challenge, fixture.PrivateKeyPEM)
	if err != nil {
		t.Fatal(err)
	}
	decoded, _ := base64.StdEncoding.DecodeString(signature)
	hash := sha256.Sum256([]byte(fixture.Challenge))
	if !ecdsa.VerifyASN1(key, hash[:], decoded) {
		t.Fatal("challenge signature failed")
	}
	if got := Fingerprint(fixture.PublicKeyPEM); got == "" {
		t.Fatal("missing fingerprint")
	}
}

func TestUnsignedRequestsDoNotSendSignatureHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Signature") != "" {
			t.Fatal("unsigned client sent X-Signature")
		}
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, err := New("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.Wallets.Balances(); err != nil {
		t.Fatal(err)
	}
}

func TestRetriesUseFreshSignatureNonces(t *testing.T) {
	fixture := loadFixture(t)
	var headers []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		headers = append(headers, request.Header.Get("X-Signature"))
		if len(headers) == 1 {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, err := New("test-key", WithBaseURL(server.URL), WithRetries(1), WithSigningKey(fixture.PrivateKeyPEM, fixture.KeyID))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.Wallets.Balances(); err != nil {
		t.Fatal(err)
	}
	if len(headers) != 2 {
		t.Fatalf("expected two attempts, got %d", len(headers))
	}
	if headers[0] == headers[1] {
		t.Fatal("retry reused the signature header")
	}
	if strings.Split(headers[0], ",")[1] == strings.Split(headers[1], ",")[1] {
		t.Fatal("retry reused the signature nonce")
	}
}
