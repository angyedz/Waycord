# Server workspace: reference screen inventory

The prototype uses local fixture data. The reference describes presentation and interaction; it does not imply that Discord, voice, streaming, or device access are connected.

| Region | Element | Visibility and interaction |
| --- | --- | --- |
| Server rail | Home and servers, unread marks, badges | Home and joined servers only; unread marks and counts only when provided. Selecting an entry changes the workspace. |
| Server sidebar | Name, banner, menu, invite | Show the banner only when configured; show invite and management actions only when permitted. The menu opens server actions. |
| Server sidebar | Boost status, guide, events, roles, member directory | Each row depends on the server's features, data and user's permissions. Hide boost progress when no boost data exists. In the offline preview, guide steps and event interest update locally. |
| Server sidebar | Categories and channels | Preserve server-defined category names and order; allow categories to collapse. Render uncategorized channels directly. Show unread, selected and permission states as appropriate. |
| Server sidebar | Channel actions | Show relevant actions on hover or focus and only when permitted; text and voice channels have different actions. |
| Workspace header | Channel title and channel actions | Reflect the selected channel. Pins filters pinned messages and search filters the active conversation. The members control toggles the right panel. Notification controls are not shown until they can affect a real notification service. |
| Messages | Authors, timestamps, date separators, mentions | Group consecutive messages from the same author in the same minute; keep each message separately selectable and actionable. Date separators and mention highlights depend on message data. |
| Messages | Markdown and attachments | Format supported markup; display image previews and file metadata, with bounded file reads. A local preview must never imply a network upload. |
| Messages | Hover actions and context menu | Show actions on hover or focus without changing row dimensions. Include only actions supported for the message and user; reactions, reply, copy and menu are independently addressable. |
| Composer | Text, file picker, emoji, GIF and stickers | Keep pending files visible and removable. Emoji, GIF and sticker choices add local sample content to the active conversation; no live media search or upload occurs. |
| Members | Role groups, status, bot badges, activities | Show only when the members panel is enabled, on a server conversation. Group by supplied roles, and show activities or bot badges only when present. |
| Voice | Channel, connection status, disconnect | Show the connection card only while in a voice channel. Switching servers or opening direct messages must not hide an active connection. |
| Voice | Microphone, headphones, camera, screen share | Mute and deafen reflect their own states; camera and screen controls appear while connected. Stop-sharing appears only during a share; disconnect clears camera and sharing state. |
| User panel | Avatar, name, presence, settings | Always show the current local user. Microphone and headphone states remain visible independently of the current server. |
| Own profile | Avatar, display name, custom status, presence, voice location | Open from the user panel. Presence choices and profile edits update local state. Voice location is hidden unless connected. |
| Settings | Appearance, Performance, Audio, Video, Updates | Appearance changes persisted theme and message density. Performance can reduce motion. Audio controls the local voice state. Video controls appear only in an active voice preview. Updates owns release checks and their preferences. |

## Implemented so far

The current prototype includes a floating combined navigation panel with a divider, server supplied categories, conditional guide/event/role/boost screens, grouped messages, reaction, copy, reply and pin actions, channel search, local file selection, a local emoji/GIF/sticker picker, a virtualized hideable member panel grouped by supplied roles, optional bot/activity labels, a clickable self profile, persistent appearance and motion settings, and local voice state controls. Live media search and upload, server permissions, real audio/video, notification delivery, cross-server search, and production update installation remain to be implemented. Do not add inert controls for these features.
