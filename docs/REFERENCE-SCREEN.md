# Server workspace: reference screen inventory

The prototype uses local fixture data. The reference describes presentation and interaction; it does not imply that Discord, voice, streaming, or device access are connected.

| Region | Element | Visibility and interaction |
| --- | --- | --- |
| Server rail | Home and servers, unread marks, badges | Home and joined servers only; unread marks and counts only when provided. Selecting an entry changes the workspace. |
| Server sidebar | Name, banner, menu, invite | Show the banner only when configured; show invite and management actions only when permitted. The menu opens server actions. |
| Server sidebar | Boost status, guide, events, roles, member directory | Each row depends on the server's features, data and user's permissions. Hide boost progress when no boost data exists. |
| Server sidebar | Categories and channels | Preserve server-defined category names and order; allow categories to collapse. Render uncategorized channels directly. Show unread, selected and permission states as appropriate. |
| Server sidebar | Channel actions | Show relevant actions on hover or focus and only when permitted; text and voice channels have different actions. |
| Workspace header | Channel title and channel actions | Reflect the selected channel; show notifications, pins, members and search only when supported. The members control toggles the right panel. |
| Messages | Authors, timestamps, date separators, mentions | Group consecutive messages from the same author in the same minute; keep each message separately selectable and actionable. Date separators and mention highlights depend on message data. |
| Messages | Markdown and attachments | Format supported markup; display image previews and file metadata, with bounded file reads. A local preview must never imply a network upload. |
| Messages | Hover actions and context menu | Show actions on hover or focus without changing row dimensions. Include only actions supported for the message and user; reactions, reply, copy and menu are independently addressable. |
| Composer | Text, file picker, emoji, GIF and stickers | Enable supported input; keep pending attachments visible and removable. GIF and sticker controls require a working local or connected source. |
| Members | Role groups, status, bot badges, activities | Show only when the members panel is enabled, on a server conversation. Group by supplied roles, and show activities or bot badges only when present. |
| Voice | Channel, connection status, disconnect | Show the connection card only while in a voice channel. Switching servers or opening direct messages must not hide an active connection. |
| Voice | Microphone, headphones, camera, screen share | Mute and deafen reflect their own states; camera and screen controls appear while connected. Stop-sharing appears only during a share; disconnect clears camera and sharing state. |
| User panel | Avatar, name, presence, settings | Always show the current local user. Microphone and headphone states remain visible independently of the current server. |

## Implemented so far

The current prototype includes a floating combined navigation panel with a divider, grouped messages, reaction and copy actions, local file selection, a hideable member panel, and a local voice state preview. Server permissions, server-specific features, rich role and activity data, reply, GIF/sticker selection, real audio/video, search, pins, and production update installation remain to be implemented. Do not add inert controls for these features.
