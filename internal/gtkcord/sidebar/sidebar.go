// Package sidebar contains the sidebar showing guilds and channels.
package sidebar

import (
	"context"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotkit/gtkutil/cssutil"
	"github.com/diamondburned/gtkcord4/internal/gtkcord/sidebar/channels"
	"github.com/diamondburned/gtkcord4/internal/gtkcord/sidebar/guilds"
)

// Controller is the parent controller that Sidebar controls.
type Controller interface {
	channels.Controller
	CloseGuild(permanent bool)
}

// Sidebar is the bar on the left side of the application once it's logged in.
type Sidebar struct {
	*gtk.Box // horizontal

	Left   *gtk.Box
	Guilds *guilds.View
	Right  *gtk.Stack

	// Keep track of the last child to remove.
	current struct {
		w  gtk.Widgetter
		id discord.GuildID
	}

	ctx  context.Context
	ctrl Controller
}

var sidebarCSS = cssutil.Applier("sidebar-sidebar", `
	windowcontrols.end:not(.empty) {
		margin-right: 4px;
	}
	windowcontrols.start:not(.empty) {
		margin: 4px;
		margin-right: 0;
	}
	.sidebar-guildside {
		background-color: mix(@borders, @theme_bg_color, 0.25);
	}
`)

// NewSidebar creates a new Sidebar.
func NewSidebar(ctx context.Context, ctrl Controller) *Sidebar {
	s := Sidebar{
		ctx:  ctx,
		ctrl: ctrl,
	}

	s.Guilds = guilds.NewView(ctx, (*guildsSidebar)(&s))
	s.Guilds.Invalidate()

	// leftBox holds just the DM button and the guild view, as opposed to s.Left
	// which holds the scrolled window and the window controls.
	leftBox := gtk.NewBox(gtk.OrientationVertical, 0)
	leftBox.Append(s.Guilds)

	leftScroll := gtk.NewScrolledWindow()
	leftScroll.SetVExpand(true)
	leftScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyExternal)
	leftScroll.SetChild(leftBox)

	leftCtrl := gtk.NewWindowControls(gtk.PackStart)
	leftCtrl.SetHAlign(gtk.AlignCenter)

	s.Left = gtk.NewBox(gtk.OrientationVertical, 0)
	s.Left.AddCSSClass("sidebar-guildside")
	s.Left.Append(leftCtrl)
	s.Left.Append(leftScroll)

	s.current.w = gtk.NewWindowHandle()

	s.Right = gtk.NewStack()
	s.Right.SetSizeRequest(channels.ChannelsWidth, -1)
	s.Right.SetHExpand(true)
	s.Right.AddChild(s.current.w)
	s.Right.SetVisibleChild(s.current.w)
	s.Right.SetTransitionType(gtk.StackTransitionTypeCrossfade)

	s.Box = gtk.NewBox(gtk.OrientationHorizontal, 0)
	s.Box.Append(s.Left)
	s.Box.Append(s.Right)
	s.Box.Append(gtk.NewSeparator(gtk.OrientationVertical))
	sidebarCSS(s)

	return &s
}

// guildsSidebar implements guilds.Controller.
type guildsSidebar Sidebar

func (s *guildsSidebar) OpenGuild(guildID discord.GuildID) {
	s.ctrl.CloseGuild(true)

	ch := channels.NewView(s.ctx, s.ctrl, guildID)
	ch.InvalidateHeader()
	ch.InvalidateChannels()

	s.Right.AddChild(ch)
	s.Right.SetVisibleChild(ch)

	s.removeCurrent()
	s.current.w = ch
	s.current.id = guildID
}

// CloseGuild implements guilds.Controller.
func (s *guildsSidebar) CloseGuild(permanent bool) {
	s.ctrl.CloseGuild(permanent)
	s.removeCurrent()
}

func (s *guildsSidebar) removeCurrent() {
	s.Right.Remove(s.current.w)

	s.current.w = nil
	s.current.id = 0
}