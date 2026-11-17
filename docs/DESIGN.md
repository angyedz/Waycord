# Waycord design system v0

## Intent

Dense without feeling cramped. Crisp typography, tonal dark surfaces, rounded controls and a restrained violet accent inspired by Material design. Message text sits directly on the canvas; rounded surfaces are reserved for controls and inputs. Fast response and an interface that still looks intentional when its content is empty.

## Tokens

| Role | Hex |
| --- | --- |
| Canvas | `#141218` |
| Server rail | `#0F0D13` |
| Channel rail | `#1D1B20` |
| Surface | `#2B2930` |
| Raised surface | `#36343B` |
| Accent | `#D0BCFF` |
| Accent surface | `#4A4458` |
| Primary text | `#E7E0E9` |
| Muted text | `#CAC4D0` |
| Subtle text | `#AAA4B1` |
| Positive | `#A8DAB5` |

Single source of truth: `internal/ui/theme/palette.go`.

## Desktop layout (first iteration)

- **Navigation:** server rail and channel list form one rounded surface. The channel list can be collapsed or resized; its chosen width is retained while the app runs.
- **Workspace:** flexible width, 58dp header, virtualized message list and message composer.
- **Direct messages:** their own navigation entry and contact list; a compact profile for the selected contact appears only in the direct conversation when there is room.
- **Emoji:** a bundled color font supplies a fallback on both desktop platforms; the composer includes a small local picker.
- **Updates screen:** separate workspace route with channel picker and explicit explanation of signed OTA steps.
- **Minimum window:** 800 × 540dp; responsive reflow below this size is future work.

## Interaction policies

- Buttons size to their label and intentional padding, rather than inheriting a header or panel's minimum dimensions.
- Everything in the current build is labeled `offline preview` where appropriate. Clicking Send updates local fixture data only.
- Channel controls switch between local preview conversations.
- The updates screen reports that checks are unavailable in the preview.
- Actual OTA is through `waycord-updater` CLI until trusted signing setup and GUI integration exist.
- Focus states, keyboard navigation, high DPI, accessibility, empty/loading/error views, and icons are the next design passes.

## Performance goals (not measured results)

Aim for 60–120 MiB RSS in idle desktop use once actual functionality exists, near-zero idle CPU, lazy image decoding and virtualized message lists. Validate separately on Linux and Windows; don't advertise these as achieved.
