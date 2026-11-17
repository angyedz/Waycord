package ui

import (
	"fmt"
	"image"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

type memberListEntry struct {
	role     string
	count    int
	member   *mock.Member
	activity bool
}

type memberRoleGroup struct {
	role    string
	members []*mock.Member
}

// slidingMembers only lays out the member list while it is visible or moving.
func (a *App) slidingMembers(progress float32) layout.Widget {
	return fixedWidth(unit.Dp(210*progress), func(gtx layout.Context) layout.Dimensions {
		visible, full := gtx.Constraints.Max.X, gtx.Dp(210)
		area := clip.Rect{Max: image.Pt(visible, gtx.Constraints.Max.Y)}.Push(gtx.Ops)
		offset := op.Offset(image.Pt(visible-full, 0)).Push(gtx.Ops)
		child := gtx
		child.Constraints.Min.X, child.Constraints.Max.X = full, full
		child.Constraints.Min.Y = child.Constraints.Max.Y
		a.membersPanel(child)
		offset.Pop()
		area.Pop()
		return layout.Dimensions{Size: image.Pt(visible, gtx.Constraints.Max.Y)}
	})
}

func (a *App) membersPanel(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	return layout.Inset{Top: 4, Right: 14, Bottom: 14}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Sidebar, 16, padding(layout.Inset{Top: 16, Left: 14, Right: 14, Bottom: 12}, func(gtx layout.Context) layout.Dimensions {
			members := mock.Servers[a.activeServer].Members
			entries := memberEntries(members)
			if len(entries) == 0 {
				return a.text("No members to show", 13, a.palette.Subtle)(gtx)
			}
			a.memberList.Axis = layout.Vertical
			return a.memberList.Layout(gtx, len(entries), func(gtx layout.Context, index int) layout.Dimensions {
				entry := entries[index]
				if entry.member == nil {
					heading := strings.ToUpper(entry.role)
					if entry.activity {
						heading = "ACTIVITY"
					}
					return layout.Inset{Bottom: 10}.Layout(gtx, a.text(fmt.Sprintf("%s — %d", heading, entry.count), 11, a.palette.Subtle))
				}
				if entry.activity {
					return layout.Inset{Bottom: 8}.Layout(gtx, a.memberActivityCard(*entry.member))
				}
				return layout.Inset{Bottom: 10}.Layout(gtx, a.memberRow(*entry.member))
			})
		}))
	})
}

func memberEntries(members []mock.Member) []memberListEntry {
	groups := make([]memberRoleGroup, 0, len(members))
	activities := make([]*mock.Member, 0, len(members))
	positions := make(map[string]int, len(members))
	for index := range members {
		member := &members[index]
		if hasVisibleActivity(*member) {
			activities = append(activities, member)
		}
		role := strings.TrimSpace(member.Role)
		if role == "" {
			role = "Members"
		}
		group, ok := positions[role]
		if !ok {
			group = len(groups)
			positions[role] = group
			groups = append(groups, memberRoleGroup{role: role})
		}
		groups[group].members = append(groups[group].members, member)
	}
	// Place each role heading before its members without changing role order.
	entries := make([]memberListEntry, 0, len(members)+len(groups)+len(activities)+1)
	if len(activities) > 0 {
		entries = append(entries, memberListEntry{role: "Activity", count: len(activities), activity: true})
		for _, member := range activities {
			entries = append(entries, memberListEntry{member: member, activity: true})
		}
	}
	for _, group := range groups {
		entries = append(entries, memberListEntry{role: group.role, count: len(group.members)})
		for _, member := range group.members {
			entries = append(entries, memberListEntry{member: member})
		}
	}
	return entries
}

func hasVisibleActivity(member mock.Member) bool {
	status := strings.ToLower(strings.TrimSpace(member.Status))
	return strings.TrimSpace(member.Activity) != "" && status != "" && status != "offline"
}

func (a *App) memberActivityCard(member mock.Member) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Surface, 11, padding(layout.Inset{Left: 8, Right: 8, Top: 7, Bottom: 7}, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(fixedWidth(30, fixedHeight(30, func(gtx layout.Context) layout.Dimensions {
					return rounded(gtx, a.palette.AccentDim, 15, func(gtx layout.Context) layout.Dimensions {
						return layout.Center.Layout(gtx, a.text(firstLetter(member.Name), 12, a.palette.Text))
					})
				}))),
				hspace(8),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(a.text(member.Name, 12, a.palette.Text)),
						layout.Rigid(a.text(member.Activity, 11, a.palette.Muted)),
					)
				}),
			)
		}))
	}
}

func (a *App) memberRow(member mock.Member) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(fixedWidth(30, fixedHeight(30, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.AccentDim, 15, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, a.text(firstLetter(member.Name), 13, a.palette.Text))
				})
			}))),
			hspace(9),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				nameLine := []layout.FlexChild{layout.Rigid(a.text(member.Name, 13, a.palette.Text))}
				if member.Bot {
					nameLine = append(nameLine, hspace(5), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return rounded(gtx, a.palette.AccentDim, 4, padding(layout.Inset{Left: 4, Right: 4, Top: 1, Bottom: 1}, a.text("BOT", 9, a.palette.Text)))
					}))
				}
				status := member.Status
				if status == "" {
					status = "Offline"
				}
				if member.Activity != "" {
					status += " · " + member.Activity
				}
				statusColor := a.palette.Subtle
				if strings.EqualFold(member.Status, "online") {
					statusColor = a.palette.Positive
				} else if strings.EqualFold(member.Status, "idle") {
					statusColor = a.palette.Warning
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx, nameLine...)
					}),
					layout.Rigid(a.text(status, 10, statusColor)),
				)
			}),
		)
	}
}
