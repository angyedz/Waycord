# Waycord OTA v1: signed, versioned, recoverable

**Status:** working standalone updater primitives + CLI. No public releases or production signing key are configured, and UI update checks are *not* connected. Only manually invoked `check`/`install` commands with an explicitly trusted public key can download releases.

## Security design

1. Fetch a JSON manifest over HTTPS (bounded to 128 KiB). Reject non-HTTPS redirects.
2. Validate schema, project, channel, version, platform and size bounds.
3. Canonicalize the fixed Go `Payload` struct using `encoding/json.Marshal` and verify its Ed25519 signature against a **trusted** 32-byte public key. Never trust a key fetched from the same manifest.
4. Download the named raw executable over HTTPS (bounded to the signed size, max 300 MiB), checking the exact byte length and SHA-256 hash before installation.
5. Copy into `<install>/versions/vX.Y.Z/waycord[.exe]` and stage the new version in a small redundant state store. Existing builds are never overwritten.
6. On `waycord-updater launch`, the candidate must signal first-frame readiness via a random launcher challenge within 15 seconds. Otherwise the pending version is discarded and the previous build launches. After successful readiness, the new version becomes current.

Two checksummed state slots are alternated (`state-0.json`, `state-1.json`), so a torn write doesn't destroy both records. This is **crash recovery**, not anti-tamper protection against another process running as the same user. Never run the updater with elevated permissions. The launcher should eventually use an interprocess lock for concurrent executions and package-manager integrations.

**Limitations:** successful first-frame health does not guarantee long-term stability; updater self-updates, post-launch crash rollback windows, transaction locks, installation UI, Windows code-signing and full end-to-end signed-release CI are future work. Retaining previous versions consumes some disk. Do not use this to overwrite files managed by a distro package manager.

## Release manifest format

A new manifest is generated **per channel, OS and architecture**. The example below uses placeholders and is not a real release:

```json
{
  "payload": {
    "schema_version": 1,
    "project": "Waycord",
    "channel": "stable",
    "version": "v0.1.0",
    "os": "linux",
    "arch": "amd64",
    "url": "https://github.com/angyedz/Waycord/releases/download/v0.1.0/waycord-linux-amd64",
    "size": 12345678,
    "sha256": "<64 lowercase hex digits>"
  },
  "signature": "<base64 Ed25519 signature>"
}
```

Supported channel/version pairs:

- `stable`: `v1.2.3`
- `beta`: `v1.2.3-beta.4`
- `nightly`: `v1.2.3-nightly.8`

Versions must be monotonically newer; downgrades are rejected. Payloads carry `project: Waycord`, schema version `1`, a single target OS (`linux`/`windows`) and arch (`amd64`/`arm64`). A future schema version must not silently reuse the v1 signing format.

## Signing (maintainer, on a trusted offline machine)

Generate a release key **outside the repository**:

```sh
go run ./cmd/waycord-sign -mode keygen -key /secure/location/waycord-release.key
# Save the printed PUBLIC_KEY_HEX in your trusted release configuration.
```

Generate a payload JSON with the actual release URL, byte size and SHA-256 computed from the final packaged executable. Sign it:

```sh
go run ./cmd/waycord-sign -mode sign \
  -key /secure/location/waycord-release.key \
  -payload /tmp/waycord-payload.json \
  -out /tmp/waycord-manifest.json
```

The private key is **not** committed, uploaded to an Actions artifact, or bundled into the app. Upload the binary and signed manifest to GitHub Releases as versioned assets after testing. `-out` currently replaces the named output file; keep signed release assets immutable once published.

## Client commands

Download release metadata and verify without installing:

```sh
waycord-updater check \
  -manifest https://github.com/angyedz/Waycord/releases/download/v0.1.0/waycord-linux-amd64-stable.json \
  -public-key YOUR_PINNED_PUBLIC_KEY_HEX \
  -channel stable
```

Stage a signed release (not yet activated):

```sh
waycord-updater install -manifest https://.../signed-manifest.json \
  -public-key YOUR_PINNED_PUBLIC_KEY_HEX -channel stable
```

Launch and activate, falling back to the previous working build on early startup failure:

```sh
waycord-updater launch
waycord-updater status
```

The default user-scoped installation is under `os.UserConfigDir()/waycord/install`. `-dir` can override it for testing. Windows and Linux have separate signed executables; the launcher never treats a zip archive or installer script as an executable.

## Release rollout checklist

1. Compile and test the precise Go version pinned in `go.mod` on both target OSes.
2. Build reproducible release artifacts and review their dependencies.
3. Sign each immutable release payload with the offline Ed25519 key.
4. Smoke-test `check → install → launch` from a clean installation and upgrade from an old version.
5. Simulate a candidate that exits before health acknowledgment; verify fallback to the previous build.
6. Integrate opt-in GUI update checks only after the public key and distribution channel are established.
