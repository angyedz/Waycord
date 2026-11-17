# Waycord architecture (initial pass)

Waycord targets **Linux and Windows as first-class platforms**. The core and UI should stay native, lean, and independently testable. The design and update system are the only working subsystems in this first PR.

## Non-negotiable boundaries

1. `cmd/waycord` is the composition root; it starts Gio and reports the first healthy frame to the launcher.
2. `internal/ui` owns only presentation state, widgets and local preview fixtures. No sockets, tokens, or Discord endpoints here.
3. `internal/ota` owns all update trust checks, version slots, state and rollback. It cannot touch authentication secrets.
4. `cmd/waycord-sign` is an offline maintainer-only tool, never loaded into the application runtime.
5. New features each get a narrow interface and their own package; don't create a general-purpose `utils` package.

## Proposed future modules (NOT IMPLEMENTED)

```text
internal/core/        application state and event bus
internal/protocol/    authenticated transports and API boundaries
internal/security/    system keyring, authorization and secret redaction
internal/media/       voice, DAVE, camera, streaming
internal/platform/    Linux / Windows capture, notifications, integration
internal/storage/     database, bounded media cache, migrations
```

The future GUI must only depend on application service interfaces, never on protocol implementation details. Keep full-resolution video frames out of Go's long-lived heap when possible, and budget CPU/RAM by feature. Media libraries load on demand.

## Quality gates

- `go test ./...` on Linux and Windows.
- `go vet ./...` in CI.
- OTA signature and hash checks are never bypassed in production paths.
- No Discord account is needed to run the visual prototype or updater tests.
- No silent release installation: the standalone CLI is explicit; GUI integration will be opt-in when signed release infrastructure exists.

## Packaging considerations

On Linux, a user-scoped portable app can be installed under a private app directory. Package-manager-managed installations (RPM/DEB/Flatpak) must delegate updating to their package manager instead of using the standalone binary replacement path. Windows uses versioned `.exe` slots so that the running binary need not be replaced. Installer and publisher integration are future work.
