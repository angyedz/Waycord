# Waycord design system

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

## Desktop layout

- **Navigation:** server rail and channel list form one rounded surface, with a fine divider. The channel list can be collapsed or resized; its chosen width is retained while the app runs. Server supplied categories collapse by their own names; uncategorized channels stay at the top.
- **Workspace:** flexible width, 58dp header, virtualized message list and message composer.
- **Direct messages:** their own navigation entry and contact list; a compact profile for the selected contact appears only in the direct conversation when there is room.
- **Own profile:** clicking the user panel opens a floating card with local presence selection, editable display name and status; voice location appears only while connected.
- **Members:** the server member panel is optional and is only laid out while open.
- **Message tools:** search filters the current conversation. Pins filter the current channel; pin, unpin, reply and copy actions are available from the message menu.
- **Emoji:** a bundled color font supplies a fallback on both desktop platforms; the composer includes a small local picker.
- **Reactions:** hovering over a message shows quick reactions and a local emoji picker; selecting an existing reaction again removes it. Hover state does not change row height.
- **Settings:** Appearance offers dark, midnight and light themes plus compact messages; Performance can reduce motion; Audio controls microphone/deafen state; Video controls are only present in an active voice preview; Updates contains signed release discovery.
- **Motion:** controls, overlays, sidebars and workspace transitions animate by default. The Reduce motion setting disables movement and hover transitions.
- **Minimum window:** 800 × 540dp, with compact behavior for narrow widths.

## Interaction policies

- Buttons size to their label and intentional padding, rather than inheriting a header or panel's minimum dimensions.
- The application uses local fixture data. Send, reactions, reply, pins, profile edits, search, channel selection, voice controls and settings change local state only.
- Channel controls switch between local preview conversations.
- User settings are stored under the operating-system configuration directory with two recoverable generations.
- Update discovery is available in Settings when a trusted signing key is configured. Installation stays in the standalone updater until the trust key and release workflow are configured.
- Focus states, keyboard navigation, high DPI, accessibility, empty/loading/error views, and icons are the next design passes.

## Performance goals (not measured results)

Aim for 60–120 MiB RSS in idle desktop use once actual functionality exists, near-zero idle CPU, lazy image decoding and virtualized message lists. Validate separately on Linux and Windows; don't advertise these as achieved.
