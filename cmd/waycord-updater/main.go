// Waycord's standalone, non-GUI update tool and recovery launcher.
package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/angyedz/Waycord/internal/ota"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "waycord-updater:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: waycord-updater [status|check|install|launch] [flags]")
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	root := fs.String("dir", filepath.Join(configDir, "waycord", "install"), "per-user installation directory")
	manifestURL := fs.String("manifest", "", "HTTPS URL of release manifest")
	publicKeyHex := fs.String("public-key", "", "trusted Ed25519 public key in hex (64 hex characters)")
	channel := fs.String("channel", "stable", "update channel: stable, beta, nightly")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	store := ota.Store{Root: *root}
	switch args[0] {
	case "status":
		st, gen, _, err := store.Load()
		if err != nil {
			return err
		}
		fmt.Printf("current=%q previous=%q pending=%q generation=%d\n", st.Current, st.Previous, st.Pending, gen)
		return nil
	case "launch":
		return store.Launch(context.Background())
	case "check", "install":
		if *manifestURL == "" {
			return fmt.Errorf("-manifest must be supplied")
		}
		key, err := hex.DecodeString(*publicKeyHex)
		if err != nil || len(key) != ed25519.PublicKeySize {
			return fmt.Errorf("-public-key must contain a trusted 32-byte Ed25519 key")
		}
		if !ota.IsChannel(*channel) {
			return fmt.Errorf("invalid channel")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if args[0] == "install" {
			p, err := store.CheckAndInstall(ctx, *manifestURL, *channel, ed25519.PublicKey(key))
			if err != nil {
				return err
			}
			fmt.Printf("Staged %s; run 'waycord-updater launch' to verify and activate.\n", p.Version)
			return nil
		}
		p, err := ota.FetchManifest(ctx, *manifestURL, ed25519.PublicKey(key))
		if err != nil {
			return err
		}
		if p.Channel != *channel {
			return fmt.Errorf("manifest is for %s, expected %s", p.Channel, *channel)
		}
		if !p.PlatformMatches() {
			return fmt.Errorf("manifest is for %s/%s, not this platform", p.OS, p.Arch)
		}
		st, _, _, err := store.Load()
		if err != nil {
			return err
		}
		if st.Current != "" {
			newer, err := ota.IsNewer(p.Version, st.Current)
			if err != nil {
				return err
			}
			if !newer {
				fmt.Println("No newer release available.")
				return nil
			}
		}
		fmt.Printf("Available: %s (%s, %d bytes)\n", p.Version, p.Channel, p.Size)
		return nil
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}
