# Waycord

**Experimental native Discord client shell** — Go + [Gio](https://gioui.org), targeting **Linux and Windows**.

> **Development status:** UI is currently an **offline design preview** with local fixture messages. No Discord login, messaging, calls, or screen sharing are implemented. Update settings can inspect signed GitHub Releases when a trusted public key is supplied to the build; this repository has no release signing key or published releases. Waycord is not affiliated with Discord Inc. and unofficial user clients may violate Discord's terms and carry account restrictions.

The current Markdown renderer does not yet support Discord spoilers or embeds.

## Current scope

- Native, GPU-rendered UI with server rail, server-defined categories, conditional guide/events/roles/boost sections, local chat, channel search and pins, replies, an optional member panel, editable local profile, voice controls, and local emoji/GIF/sticker choices.
- Appearance and performance settings for dark/midnight/light themes, compact messages and reduced motion; preferences are stored in recoverable user-scoped files.
- Rich, selectable message text with bold, italic, code, headings, quotes, lists, links, underline and strikethrough formatting. Links open only after a click and use the OS handler.
- Independent updater with Ed25519 signed manifests, SHA-256 verification, versioned binary slots, and startup-health rollback.
- Update settings for stable, beta and nightly channels, with persisted automatic-check preference and signed GitHub release discovery.
- Separate Go packages for UI, theme tokens, test fixtures, update engine, signing utility, and launcher.
- CI builds/tests for Linux and Windows.

## Build and run

Requires **Go 1.24+** (Gio v0.10.3). On Linux, install Gio's [development libraries](https://gioui.org/doc/install/linux).

```sh
go run ./cmd/waycord             # offline GUI preview
go run ./cmd/waycord-updater status
go test ./...
```

To build both tools:

```sh
go build -o waycord ./cmd/waycord
go build -o waycord-updater ./cmd/waycord-updater
```

On Windows, use `.exe` file extensions in the output names. No browser/WebView/Electron dependency is used in the app.

## Project organization

```text
cmd/
  waycord/              # Gio application / first-frame health signal
  waycord-updater/      # standalone update checker, installer and launcher
  waycord-sign/         # offline release signing utility
internal/
  ui/                   # presentation layer (no Discord network code)
    theme/              # palette design tokens
    mock/               # explicitly offline demo data
  ota/                  # signed release verification, download, slots, rollback
  version/              # build metadata
.github/workflows/      # Linux and Windows CI
docs/
  ARCHITECTURE.md       # module boundaries and future roadmap
  DESIGN.md             # visual direction and interaction conventions
  OTA.md                # release format, security model, setup guide
```

## OTA is opt-in until release signing is configured

The updater **fails closed** unless a trusted Ed25519 public key is injected into a release build. The public key is a verifier and is not secret; the private signing key must remain offline and outside the repository. Automatic checks are enabled by default but make no network request while the trusted key is absent. Start with [docs/OTA.md](docs/OTA.md).

## Next

Implement the real application state layer and audited protocol adapter **separately**. Calls, streams, and screen sharing are planned, not present. Test experimental integrations off the primary account. Follow Discord policies and don't automate account actions.
