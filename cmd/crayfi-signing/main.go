package main

import (
	"fmt"
	"os"

	crayfi "github.com/noibilism/crayfi-go"
)

func main() {
	if len(os.Args) < 2 {
		fail("Usage: crayfi-signing <keygen|sign-challenge> [challenge]")
	}

	switch os.Args[1] {
	case "keygen":
		keys, err := crayfi.GenerateSigningKeyPair()
		if err != nil {
			fail(err.Error())
		}
		fmt.Print(keys.PrivateKeyPEM)
		fmt.Print(keys.PublicKeyPEM)
		fmt.Printf("Fingerprint: %s\n", keys.Fingerprint)
	case "sign-challenge":
		if len(os.Args) != 3 {
			fail("Usage: crayfi-signing sign-challenge <challenge>")
		}
		privateKey := os.Getenv("CRAY_SIGNING_PRIVATE_KEY")
		if privateKey == "" && os.Getenv("CRAY_SIGNING_PRIVATE_KEY_PATH") != "" {
			bytes, err := os.ReadFile(os.Getenv("CRAY_SIGNING_PRIVATE_KEY_PATH"))
			if err != nil {
				fail(err.Error())
			}
			privateKey = string(bytes)
		}
		if privateKey == "" {
			fail("Set CRAY_SIGNING_PRIVATE_KEY or CRAY_SIGNING_PRIVATE_KEY_PATH")
		}
		signature, err := crayfi.SignChallenge(os.Args[2], privateKey)
		if err != nil {
			fail(err.Error())
		}
		fmt.Println(signature)
	default:
		fail("Usage: crayfi-signing <keygen|sign-challenge> [challenge]")
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
