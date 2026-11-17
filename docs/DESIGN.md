# Waycord design system v0

## Intent

Dense without feeling cramped. Crisp typography, neutral charcoal panels and a restrained sage accent. Message text sits directly on the canvas; rounded surfaces are reserved for controls and inputs. Fast response and an interface that still looks intentional when its content is empty.

## Tokens

| Role | Hex |
| --- | --- |
| Canvas | `#17191F` |
| Server rail | `#101217` |
| Channel rail | `#1C1F26` |
| Surface | `#242830` |
| Raised surface | `#303540` |
| Accent | `#91B9AD` |
| Accent surface | `#33443F` |
| Primary text | `#ECEEF2` |
| Muted text | `#B4BAC5` |
| Subtle text | `#89919F` |
| Positive | `#8CCAB0` |

Single source of truth: `internal/ui/theme/palette.go`.

## Desktop layout (first iteration)

- **Server rail:** 64dp, full window height, strong contrast with the rest.
- **Channel rail:** 216dp, grouped text/voice channels; bottom connection-state section.
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
