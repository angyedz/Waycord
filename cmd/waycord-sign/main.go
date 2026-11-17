// An offline maintainer tool. Never commit generated private keys.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/angyedz/Waycord/internal/ota"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "waycord-sign:", err)
		os.Exit(1)
	}
}
func run() error {
	mode := flag.String("mode", "", "keygen or sign")
	keyPath := flag.String("key", "", "private key file path")
	payloadPath := flag.String("payload", "", "input payload JSON")
	out := flag.String("out", "", "output manifest JSON")
	flag.Parse()
	if *keyPath == "" {
		return errors.New("-key is required")
	}
	switch *mode {
	case "keygen":
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}
		f, err := os.OpenFile(*keyPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		if _, err = f.WriteString(base64.StdEncoding.EncodeToString(priv)); err != nil {
			f.Close()
			return err
		}
		if err = f.Close(); err != nil {
			return err
		}
		fmt.Printf("PUBLIC_KEY_HEX=%x\n", pub)
		return nil
	case "sign":
		if *payloadPath == "" || *out == "" {
			return errors.New("-payload and -out required")
		}
		encoded, err := os.ReadFile(*keyPath)
		if err != nil {
			return err
		}
		priv, err := base64.StdEncoding.DecodeString(string(encoded))
		if err != nil || len(priv) != ed25519.PrivateKeySize {
			return errors.New("invalid private key")
		}
		b, err := os.ReadFile(*payloadPath)
		if err != nil {
			return err
		}
		var p ota.Payload
		if err = json.Unmarshal(b, &p); err != nil {
			return err
		}
		body, err := p.Bytes()
		if err != nil {
			return err
		}
		m := ota.Manifest{Payload: p, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(ed25519.PrivateKey(priv), body))}
		result, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return err
		}
		// Avoid overwriting an existing signed release accidentally.
		return os.WriteFile(*out, append(result, '\n'), 0600)
	default:
		return errors.New("choose -mode=keygen or -mode=sign")
	}
}
