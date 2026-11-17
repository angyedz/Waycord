# Waycord design system v0

## Intent

Dense without feeling cramped. Crisp typography, muted slate panels, one restrained violet accent, comfortable 12–16dp corners, no unnecessary gradients. Fast response and an interface that still looks intentional when its content is empty.

## Tokens

| Role | Hex |
| --- | --- |
| Canvas | `#10141D` |
| Server rail | `#0B0E15` |
| Channel rail | `#141925` |
| Surface | `#1B2230` |
| Raised surface | `#252D3E` |
| Accent | `#9585FA` |
| Accent surface | `#342E58` |
| Primary text | `#F0F3FA` |
| Muted text | `#A0AABD` |
| Subtle text | `#697487` |
| Positive | `#60D5A4` |

Single source of truth: `internal/ui/theme/palette.go`.

## Desktop layout (first iteration)

- **Server rail:** 66dp, full window height, strong contrast with the rest.
- **Channel rail:** 230dp, grouped text/voice channels; bottom connection-state section.
- **Workspace:** flexible width, 70dp header, virtualized message list and message composer.
- **Updates screen:** separate workspace route with channel picker and explicit explanation of signed OTA steps.
- **Minimum window:** 800 × 540dp; responsive reflow below this size is future work.

## Interaction policies

- Everything in the current build is labeled `offline preview` where appropriate. Clicking Send updates local fixture data only.
- Server/channel controls work with preview state, without network traffic.
- Update channel selectors change design-preview state only; they do not claim to check for actual releases.
- Actual OTA is through `waycord-updater` CLI until trusted signing setup and GUI integration exist.
- Focus states, keyboard navigation, high DPI, accessibility, empty/loading/error views, and icons are the next design passes.

## Performance goals (not measured results)

Aim for 60–120 MiB RSS in idle desktop use once actual functionality exists, near-zero idle CPU, lazy image decoding and virtualized message lists. Validate separately on Linux and Windows; don't advertise these as achieved.
