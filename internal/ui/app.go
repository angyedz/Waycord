package ui

import (
	"context"
	"crypto/ed25519"
	"image"
	"strings"
	"time"

	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"github.com/angyedz/Waycord/internal/core/attachments"
	"github.com/angyedz/Waycord/internal/core/reactions"
	"github.com/angyedz/Waycord/internal/core/voicepreview"
	"github.com/angyedz/Waycord/internal/ota"
	uiMarkdown "github.com/angyedz/Waycord/internal/ui/markdown"
	"github.com/angyedz/Waycord/internal/ui/mock"
	"github.com/angyedz/Waycord/internal/ui/theme"
)

// App contains UI state only. Future protocol/media services stay outside it.
type App struct {
	palette                                theme.Palette
	theme                                  *material.Theme
	messageMarkdown                        *uiMarkdown.Cache
	channelButtons                         [][]widget.Clickable
	chatButton, updatesButton, sendButton  widget.Clickable
	searchButton, searchCloseButton        widget.Clickable
	pinnedButton                           widget.Clickable
	updateCheckButton, updateAutoButton    widget.Clickable
	updateOpenButton                       widget.Clickable
	updateChannelButtons                   [3]widget.Clickable
	settingsSections                       [5]widget.Clickable
	themeButtons                           [3]widget.Clickable
	compactButton                          widget.Clickable
	reduceMotionButton                     widget.Clickable
	settingsSection                        int
	compactMessages                        bool
	reduceMotion                           bool
	themeName                              string
	settingsStatus                         string
	saveUIPreferences                      func(string, bool, bool) error
	directButton, contactButton            widget.Clickable
	serverButtons                          []widget.Clickable
	categoryButtons                        map[string]*widget.Clickable
	collapsedCategories                    map[string]bool
	categoryTransitions                    map[string]*panelTransition
	sidebarButton, profileButton           widget.Clickable
	membersButton                          widget.Clickable
	voiceJoin, voiceLeave                  widget.Clickable
	voiceMute, voiceDeafen                 widget.Clickable
	voiceCamera, voiceScreen               widget.Clickable
	voice                                  voicepreview.State
	voicePanelSlide                        panelTransition
	selfProfileButton, selfProfileBackdrop widget.Clickable
	selfStatusButtons                      [4]widget.Clickable
	selfProfileEditButton                  widget.Clickable
	selfProfileSaveButton                  widget.Clickable
	selfProfileCancelButton                widget.Clickable
	selfProfileNameEditor                  widget.Editor
	selfProfileStatusEditor                widget.Editor
	replyCancelButton                      widget.Clickable
	selfProfileOpen                        bool
	selfProfileEditing                     bool
	selfProfileSlide                       panelTransition
	selfPresence                           int
	selfProfileName                        string
	selfProfileUsername                    string
	selfProfileStatus                      string
	voiceServer, voiceChannel              string
	micIcon, micOffIcon, headsetIcon       *widget.Icon
	videoIcon, videoOffIcon                *widget.Icon
	screenIcon, screenOffIcon, hangupIcon  *widget.Icon
	sidebar                                sidebarState
	menuIcon                               *widget.Icon
	sendIcon                               *widget.Icon
	settingsIcon                           *widget.Icon
	searchIcon, pinnedIcon, closeIcon      *widget.Icon
	replyIcon                              *widget.Icon
	moreIcon                               *widget.Icon
	attachIcon                             *widget.Icon
	attachButton, removeAttachmentButton   widget.Clickable
	attachmentDrafts                       map[string]*attachments.Attachment
	attachmentResults                      chan attachmentResult
	chooseAttachment                       func() (*attachments.Attachment, error)
	attachmentBusy                         bool
	attachmentError                        string
	attachmentPreviews                     attachmentPreviewCache
	membersIcon                            *widget.Icon
	interactions                           map[*widget.Clickable]*interactionState
	emojiIcon                              *widget.Icon
	emojiButton                            widget.Clickable
	emojiChoices                           [8]widget.Clickable
	emojiOpen                              bool
	emojiSlide                             panelTransition
	searchOpen                             bool
	pinsOnly                               bool
	searchEditor                           widget.Editor
	searchSlide                            panelTransition
	sendHover                              float32
	sendHoverTime                          time.Time
	activeChannel                          int
	activeServer                           int
	serverChannels                         []int
	composerDrafts                         map[string]string
	updatesView                            bool
	updateChannel                          string
	updateAutoCheck                        bool
	updateChecking                         bool
	updateStatus                           string
	updateResult                           chan updateCheckResult
	updateSignals                          chan struct{}
	updateChecker                          interface {
		Check(context.Context, string, string, string, string, ed25519.PublicKey) (ota.Update, error)
	}
	saveUpdatePreferences                      func(string, bool) error
	updateCtx                                  context.Context
	updateInvalidate                           func()
	availableUpdate                            *ota.Update
	direct                                     bool
	profileOpen                                bool
	profileSlide                               panelTransition
	membersOpen                                bool
	membersSlide                               panelTransition
	contentKey                                 string
	contentSlide                               panelTransition
	messageList                                layout.List
	updateList                                 layout.List
	serverChannelList                          layout.List
	memberList                                 layout.List
	composer                                   widget.Editor
	messages                                   []mock.Message
	reactions                                  reactions.Store
	reactionWidgets                            map[int]*reactionWidget
	reactionTarget                             int
	contextTarget                              int
	contextShown                               int
	replyTarget                                int
	contextPosition                            image.Point
	contextSlide                               panelTransition
	contextBackdrop, contextCopy, contextReact widget.Clickable
	contextReply                               widget.Clickable
	contextPin                                 widget.Clickable
}

func New() *App {
	a := &App{reactionWidgets: make(map[int]*reactionWidget), reactionTarget: -1, contextTarget: -1, contextShown: -1, replyTarget: -1, palette: theme.Dark, themeName: "dark", activeChannel: 1, sidebar: newSidebarState(), profileOpen: true, messageList: layout.List{Axis: layout.Vertical}, updateChannel: "stable", updateResult: make(chan updateCheckResult, 1), updateSignals: make(chan struct{}, 1), updateChecker: ota.GitHubSource{Owner: "angyedz", Repository: "Waycord"}, updateCtx: context.Background(), selfProfileName: "angyedz", selfProfileUsername: "angyedz", selfProfileStatus: "On a break"}
	a.selfPresence = 3
	a.theme = material.NewTheme()
	if shaper, err := theme.NewShaper(); err == nil {
		a.theme.Shaper = shaper
		a.theme.Face = font.Typeface(theme.TextFamilies)
	}
	a.menuIcon, _ = widget.NewIcon(icons.NavigationMenu)
	a.sendIcon, _ = widget.NewIcon(icons.ContentSend)
	a.settingsIcon, _ = widget.NewIcon(icons.ActionSettings)
	a.searchIcon, _ = widget.NewIcon(icons.ActionSearch)
	a.pinnedIcon, _ = widget.NewIcon(icons.ActionLabel)
	a.closeIcon, _ = widget.NewIcon(icons.NavigationClose)
	a.attachIcon, _ = widget.NewIcon(icons.ContentAddCircleOutline)
	a.attachmentDrafts = make(map[string]*attachments.Attachment)
	a.serverButtons = make([]widget.Clickable, len(mock.Servers))
	a.serverChannels = make([]int, len(mock.Servers))
	a.channelButtons = make([][]widget.Clickable, len(mock.Servers))
	for i, server := range mock.Servers {
		a.channelButtons[i] = make([]widget.Clickable, len(server.Channels))
	}
	a.categoryButtons = make(map[string]*widget.Clickable)
	a.collapsedCategories = make(map[string]bool)
	a.categoryTransitions = make(map[string]*panelTransition)
	a.composerDrafts = make(map[string]string)
	a.serverChannels[0] = a.activeChannel
	a.attachmentResults = make(chan attachmentResult, 1)
	a.replyIcon, _ = widget.NewIcon(icons.ContentReply)
	a.moreIcon, _ = widget.NewIcon(icons.NavigationMoreHoriz)
	a.membersIcon, _ = widget.NewIcon(icons.SocialPeople)
	a.micIcon, _ = widget.NewIcon(icons.AVMic)
	a.micOffIcon, _ = widget.NewIcon(icons.AVMicOff)
	a.headsetIcon, _ = widget.NewIcon(icons.HardwareHeadset)
	a.videoIcon, _ = widget.NewIcon(icons.AVVideocam)
	a.videoOffIcon, _ = widget.NewIcon(icons.AVVideocamOff)
	a.screenIcon, _ = widget.NewIcon(icons.CommunicationScreenShare)
	a.screenOffIcon, _ = widget.NewIcon(icons.CommunicationStopScreenShare)
	a.hangupIcon, _ = widget.NewIcon(icons.CommunicationCallEnd)
	a.emojiIcon, _ = widget.NewIcon(icons.SocialSentimentSatisfied)
	a.theme.Palette = material.Palette{Bg: a.palette.Canvas, Fg: a.palette.Text, ContrastBg: a.palette.Accent, ContrastFg: a.palette.Rail}
	a.messageMarkdown = uiMarkdown.NewCache(uiMarkdown.New(theme.TextFamilies, a.palette.Muted, a.palette.Accent), 64)
	a.composer.SingleLine = true
	a.composer.Submit = true
	a.composer.MaxLen = 2000
	a.searchEditor.SingleLine = true
	a.searchEditor.Submit = true
	a.searchEditor.MaxLen = 128
	a.selfProfileNameEditor.SingleLine = true
	a.selfProfileNameEditor.MaxLen = 32
	a.selfProfileNameEditor.SetText(a.selfProfileName)
	a.selfProfileStatusEditor.SingleLine = true
	a.selfProfileStatusEditor.MaxLen = 80
	a.selfProfileStatusEditor.SetText(a.selfProfileStatus)
	a.messages = append([]mock.Message(nil), mock.Messages...)
	return a
}

func (a *App) process(gtx layout.Context) {
	a.processAttachmentResults()
	a.processUpdateChecks()
	a.processContextMenu(gtx)
	a.processSearch(gtx)
	a.processReactions(gtx)
	if a.emojiButton.Clicked(gtx) {
		a.emojiOpen = !a.emojiOpen
	}
	for i := range a.emojiChoices {
		if a.emojiChoices[i].Clicked(gtx) {
			a.composer.Insert(emojiOptions[i])
			a.emojiOpen = false
			gtx.Execute(key.FocusCmd{Tag: &a.composer})
		}
	}
	if a.sidebarButton.Clicked(gtx) {
		a.sidebar.toggle()
	}
	if a.profileButton.Clicked(gtx) {
		a.profileOpen = !a.profileOpen
	}
	if a.membersButton.Clicked(gtx) {
		a.membersOpen = !a.membersOpen
	}
	if a.voiceJoin.Clicked(gtx) {
		server := mock.Servers[a.activeServer]
		if a.voice.Connected && (a.voiceServer != server.Name || a.voiceChannel != server.VoiceChannel) {
			a.voice.Leave()
		}
		a.voiceServer, a.voiceChannel = server.Name, server.VoiceChannel
		a.voice.Join()
	}
	if a.voiceLeave.Clicked(gtx) {
		a.voice.Leave()
	}
	if a.voiceMute.Clicked(gtx) {
		a.voice.ToggleMute()
	}
	if a.voiceDeafen.Clicked(gtx) {
		a.voice.ToggleDeafen()
	}
	if a.voiceCamera.Clicked(gtx) {
		a.voice.ToggleCamera()
	}
	if a.voiceScreen.Clicked(gtx) {
		a.voice.ToggleScreen()
	}
	a.processSelfProfile(gtx)
	if a.attachButton.Clicked(gtx) {
		a.beginAttachmentSelection()
	}
	if a.removeAttachmentButton.Clicked(gtx) {
		delete(a.attachmentDrafts, a.conversationKey())
		a.attachmentError = ""
	}
	if a.replyCancelButton.Clicked(gtx) {
		a.replyTarget = -1
	}
	a.sidebar.update(gtx)
	if a.directButton.Clicked(gtx) || a.contactButton.Clicked(gtx) {
		a.contextTarget = -1
		a.reactionTarget = -1
		a.switchConversation(true, a.activeServer, a.activeChannel)
		a.messageList.Position = layout.Position{}
		a.messageList.ScrollToEnd = false
	}
	for i := range a.serverButtons {
		if a.serverButtons[i].Clicked(gtx) {
			changedServer := a.activeServer != i
			a.switchConversation(false, i, a.serverChannels[i])
			if changedServer {
				a.serverChannelList.Position = layout.Position{}
				a.memberList.Position = layout.Position{}
			}
		}
	}
	for server, channels := range a.channelButtons {
		for channel := range channels {
			if server == a.activeServer && channels[channel].Clicked(gtx) {
				a.switchConversation(false, server, channel)
			}
		}
	}
	for key, button := range a.categoryButtons {
		if button.Clicked(gtx) {
			a.collapsedCategories[key] = !a.collapsedCategories[key]
		}
	}
	if a.chatButton.Clicked(gtx) {
		a.updatesView = false
	}
	if a.updatesButton.Clicked(gtx) {
		a.contextTarget = -1
		a.updatesView = true
		a.searchOpen = false
		a.searchEditor.SetText("")
		a.pinsOnly = false
	}
	for i := range a.settingsSections {
		if a.settingsSections[i].Clicked(gtx) {
			a.settingsSection = i
		}
	}
	if a.compactButton.Clicked(gtx) {
		a.setUIPreferences(a.themeName, !a.compactMessages, a.reduceMotion)
	}
	if a.reduceMotionButton.Clicked(gtx) {
		a.setUIPreferences(a.themeName, a.compactMessages, !a.reduceMotion)
	}
	for i, name := range [...]string{"dark", "midnight", "light"} {
		if a.themeButtons[i].Clicked(gtx) {
			a.setUIPreferences(name, a.compactMessages, a.reduceMotion)
		}
	}
	if a.updateCheckButton.Clicked(gtx) {
		a.checkForUpdates()
	}
	if a.updateAutoButton.Clicked(gtx) {
		a.setUpdatePreferences(a.updateChannel, !a.updateAutoCheck)
	}
	for i, channel := range [...]string{"stable", "beta", "nightly"} {
		if a.updateChannelButtons[i].Clicked(gtx) {
			a.setUpdatePreferences(channel, a.updateAutoCheck)
		}
	}
	if a.updateOpenButton.Clicked(gtx) && a.availableUpdate != nil {
		if err := openReleasePage(a.availableUpdate.ReleaseURL); err != nil {
			a.updateStatus = err.Error()
		}
	}
	if a.sendButton.Clicked(gtx) && !a.updatesView {
		a.sendPreview()
	}
	for {
		event, ok := a.composer.Update(gtx)
		if !ok {
			break
		}
		if _, ok := event.(widget.SubmitEvent); ok && !a.updatesView {
			a.sendPreview()
		}
	}
}
func (a *App) sendPreview() {
	body := strings.TrimSpace(a.composer.Text())
	key := a.conversationKey()
	attachment := a.attachmentDrafts[key]
	if body == "" && attachment == nil {
		return
	}
	now := time.Now()
	message := mock.Message{Server: a.activeServer, Channel: a.activeChannel, Direct: a.direct, Author: a.selfProfileName, Time: now.Format("15:04"), SentAt: now, Body: body, Self: true, Attachment: attachment}
	if a.replyTarget >= 0 && a.replyTarget < len(a.messages) {
		replied := a.messages[a.replyTarget]
		if replied.Direct == a.direct && (a.direct || (replied.Server == a.activeServer && replied.Channel == a.activeChannel)) {
			message.HasReply, message.ReplyTo = true, a.replyTarget
		}
	}
	a.messages = append(a.messages, message)
	delete(a.attachmentDrafts, key)
	a.composer.SetText("")
	a.replyTarget = -1
	a.messageList.ScrollToEnd = true
}

func (a *App) switchConversation(direct bool, server, channel int) {
	if server < 0 || server >= len(mock.Servers) || channel < 0 || channel >= len(mock.Servers[server].Channels) {
		return
	}
	a.composerDrafts[a.conversationKey()] = a.composer.Text()
	a.direct, a.activeServer, a.activeChannel, a.updatesView = direct, server, channel, false
	if !direct {
		a.serverChannels[server] = channel
	}
	a.composer.SetText(a.composerDrafts[a.conversationKey()])
	a.contextTarget, a.reactionTarget = -1, -1
	a.replyTarget = -1
	a.searchEditor.SetText("")
	a.pinsOnly = false
	a.messageList.Position = layout.Position{}
	a.messageList.ScrollToEnd = false
}

func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	a.process(gtx)
	width := a.sidebar.animate(gtx, a.reduceMotion)
	dims := rounded(gtx, a.palette.Canvas, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: 8, Top: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return rounded(gtx, a.palette.Sidebar, 18, func(gtx layout.Context) layout.Dimensions {
						return fixedWidth(52+width, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{}.Layout(gtx,
										layout.Rigid(fixedWidth(52, a.serverRail)),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											if width < 1 {
												return layout.Dimensions{}
											}
											return a.navigationDivider(gtx)
										}),
										layout.Rigid(fixedWidth(max(0, width-1), func(gtx layout.Context) layout.Dimensions {
											if width < 120 {
												return layout.Dimensions{Size: gtx.Constraints.Min}
											}
											return a.channelRail(gtx)
										})),
									)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.voicePanel(gtx, width >= 120) }),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.userPanel(gtx, width >= 120) }),
							)
						})(gtx)
					})
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if !a.sidebar.shown || width < 120 {
					return layout.Dimensions{}
				}
				return a.sidebar.handle(gtx, a.palette.Border)
			}),
			layout.Flexed(1, a.workspace),
		)
	})
	if a.contextTarget >= 0 {
		gtx.Constraints = layout.Exact(dims.Size)
		a.contextBackdrop.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: dims.Size}
		})
	}
	profileProgress := a.transition(gtx, &a.selfProfileSlide, boolProgress(a.selfProfileOpen), 220*time.Millisecond)
	if profileProgress > 0 {
		a.selfProfileOverlay(gtx, dims.Size, profileProgress)
	}
	return dims
}

func (a *App) serverRail(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		children := []layout.FlexChild{layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.railButton(gtx, &a.directButton, "@", a.direct) }), vspace(10)}
		for i, server := range mock.Servers {
			i, server := i, server
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.railButton(gtx, &a.serverButtons[i], server.Icon, !a.direct && a.activeServer == i)
			}), vspace(10))
		}
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx, children...)
	})
}

func (a *App) channelRail(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return fixedHeight(58, inset(14, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if a.direct {
							return a.text("Messages", 18, a.palette.Text)(gtx)
						}
						return a.text(mock.Servers[a.activeServer].Name, 18, a.palette.Text)(gtx)
					}),
				)
			}))(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: 8, Right: 8, Top: 12}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if a.direct {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(inset(9, a.text("DIRECT MESSAGES", 11, a.palette.Subtle))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.directContact(gtx, "mira", !a.updatesView)
						}),
					)
				}
				return a.serverChannelsLayout(gtx)
			})
		}),
	)
}

func (a *App) workspace(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(a.header),
		layout.Flexed(1, a.animatedWorkspaceContent),
	)
}

func (a *App) animatedWorkspaceContent(gtx layout.Context) layout.Dimensions {
	key := a.conversationKey()
	if a.direct {
		key = "direct:mira"
	}
	if a.updatesView {
		key = "settings:" + settingsNames[a.settingsSection]
	}
	if a.contentKey == "" {
		a.contentKey = key
		a.contentSlide.value, a.contentSlide.from, a.contentSlide.target = 1, 1, 1
		a.contentSlide.started = gtx.Now
	} else if a.contentKey != key {
		a.contentKey = key
		a.contentSlide.value, a.contentSlide.from, a.contentSlide.target = 0, 0, 0
		a.contentSlide.started = time.Time{}
	}
	progress := a.transition(gtx, &a.contentSlide, 1, 200*time.Millisecond)
	width := gtx.Dp(14)
	shift := int(float32(width) * (1 - progress))
	clipArea := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	offset := op.Offset(image.Pt(shift, 0)).Push(gtx.Ops)
	var child layout.Dimensions
	if a.updatesView {
		child = a.settings(gtx)
	} else {
		showProfile := a.direct && a.profileOpen && gtx.Constraints.Max.X >= gtx.Dp(730)
		profileProgress := a.transition(gtx, &a.profileSlide, boolProgress(showProfile), 220*time.Millisecond)
		if profileProgress > 0 && a.direct {
			child = layout.Flex{}.Layout(gtx, layout.Flexed(1, a.chat), layout.Rigid(a.slidingProfile(profileProgress)))
		} else if !a.direct {
			membersProgress := a.transition(gtx, &a.membersSlide, boolProgress(a.membersOpen && gtx.Constraints.Max.X >= gtx.Dp(730)), 220*time.Millisecond)
			if membersProgress > 0 {
				child = layout.Flex{}.Layout(gtx, layout.Flexed(1, a.chat), layout.Rigid(a.slidingMembers(membersProgress)))
			} else {
				child = a.chat(gtx)
			}
		} else {
			child = a.chat(gtx)
		}
	}
	offset.Pop()
	clipArea.Pop()
	return child
}
func (a *App) header(gtx layout.Context) layout.Dimensions {
	wide := gtx.Constraints.Max.X >= gtx.Dp(730)
	searchProgress := a.transition(gtx, &a.searchSlide, boolProgress(a.searchOpen && !a.updatesView), 180*time.Millisecond)
	return rounded(gtx, a.palette.Canvas, 0, fixedHeight(58, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: 20, Right: 20, Top: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			children := []layout.FlexChild{
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return a.iconControl(gtx, &a.sidebarButton, a.menuIcon, 36, 20, "Toggle sidebar")
					})
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if (a.searchOpen || searchProgress > 0) && !a.updatesView {
						return a.searchField(gtx, searchProgress)
					}
					title := "# " + mock.Servers[a.activeServer].Channels[a.activeChannel]
					if a.direct {
						title = "@ mira"
					}
					if a.updatesView {
						title = "Settings / " + settingsNames[a.settingsSection]
					}
					offset := op.Offset(image.Pt(0, gtx.Dp(4))).Push(gtx.Ops)
					dimensions := layout.W.Layout(gtx, a.text(title, 19, a.palette.Text))
					offset.Pop()
					return dimensions
				}),
			}
			if a.updatesView {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.pill(gtx, &a.chatButton, "Back", false)
				}))
			} else if a.searchOpen || searchProgress > 0 {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.iconControl(gtx, &a.searchCloseButton, a.closeIcon, 36, 16, "Close search")
				}))
			} else {
				if a.direct && wide {
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.pill(gtx, &a.profileButton, "Profile", a.profileOpen)
					}))
				} else if !a.direct && wide {
					children = append(children,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.iconToggle(gtx, &a.pinnedButton, a.pinnedIcon, 36, 18, "Toggle pinned messages", a.pinsOnly)
						}),
						hspace(4),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.iconControl(gtx, &a.membersButton, a.membersIcon, 36, 20, "Toggle members")
						}),
					)
				}
				children = append(children,
					hspace(4),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.iconControl(gtx, &a.searchButton, a.searchIcon, 36, 18, "Search this channel")
					}),
				)
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
		})
	}))
}
