package ui

import (
	"image"
	"path/filepath"
	"strings"

	"gioui.org/layout"

	"github.com/xmhrnx/anyfile/internal/session"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/elements"
)

// tab is one opened file: its own session (with its temp folder) and
// its own tree, so expansion, selection and filters stay per file.
type tab struct {
	id   string // the file's path, its ID in the tab strip
	sess *session.Session
	tree *treeView
}

// dragRoom is the empty title bar kept free of tabs, to move the
// window by.
const dragRoom = 96 // dp

const (
	tabFontSize    = 12.5
	tabMin, tabMax = 96, 200 // dp
	tabHeight      = 43      // the title bar, less its border
)

// newTabStrip makes the title bar's tabs. giode's DocTabs shrinks them
// when they don't fit, then scrolls, reorders them by dragging and
// shows each file's path as a tooltip. Its right-click menu is off: its
// items (Close Others, ...) are in English only.
func (u *UI) newTabStrip() {
	u.strip = giode.DocTabs(giode.Styles{
		Background: colSidebar, Color: colText, BorderColor: colBorder, FontSize: tabFontSize, Height: tabHeight,
	}).TabWidth(tabMin, tabMax).NoTabMenu().
		OnSelect(func(id string) {
			if t := u.tabByID(id); t != nil {
				u.activate(t)
			}
		}).
		OnClose(func(id string) {
			if t := u.tabByID(id); t != nil {
				u.closeTab(t)
			}
		}).
		OnReorder(func(ids []string) {
			order := make([]*tab, 0, len(ids))
			for _, id := range ids {
				if t := u.tabByID(id); t != nil {
					order = append(order, t)
				}
			}
			u.tabs = order
		})
}

func (u *UI) tabByID(id string) *tab {
	for _, t := range u.tabs {
		if t.id == id {
			return t
		}
	}
	return nil
}

// openPaths opens each path in a tab of its own, or switches to the tab
// already showing it.
func (u *UI) openPaths(paths []string) {
	for _, p := range paths {
		u.openPath(p)
	}
}

func (u *UI) openPath(p string) {
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	for _, t := range u.tabs {
		if r := t.sess.Root(); r != nil && strings.EqualFold(r.Local, p) {
			u.activate(t)
			return
		}
	}
	t := &tab{sess: session.New(u.app.Post, u.toast)}
	t.sess.KeepAwakeWith(u.app.KeepAlive)
	t.tree = newTreeView(u, t.sess)
	t.sess.OnExpand(t.tree.expand)
	t.sess.Open(p)
	root := t.sess.Root()
	if root == nil {
		return // Open reported why
	}
	t.id = root.Local
	u.tabs = append(u.tabs, t)
	u.strip.Add(giode.DocTab{ID: t.id, Title: root.Name, Tooltip: root.Local})
	u.activate(t)
}

func (u *UI) activate(t *tab) {
	u.showSettings = false
	u.strip.Select(t.id)
	if u.active != t {
		u.active = t
		u.inspector.scroll.ScrollTo(0)
	}
}

func (u *UI) closeTab(t *tab) {
	i := -1
	for k, x := range u.tabs {
		if x == t {
			i = k
		}
	}
	if i < 0 {
		return
	}
	t.sess.Close()
	u.tabs = append(u.tabs[:i], u.tabs[i+1:]...)
	if u.active == t {
		u.active = nil
	}
	// Removing the selected tab selects a neighbour through OnSelect.
	u.strip.Remove(t.id)
}

// tabStrip lays the tabs out in the title bar: as wide as they are, up
// to the room left after the Settings button, keeping dragRoom free to
// move the window by.
func (u *UI) tabStrip(gtx layout.Context) layout.Dimensions {
	tabs := u.strip.Tabs()
	if len(tabs) == 0 {
		return layout.Dimensions{}
	}
	room := gtx.Constraints.Max.X - gtx.Dp(dragRoom) - elements.Measure(gtx, u.settingsBtn).Size.X
	// Each tab's natural width, as DocTabs works it out: the title,
	// 12 dp before it, then a 4 dp gap, the 20 dp close button and
	// 6 dp, within TabWidth. (Measuring the strip itself would reset
	// its scrolling.)
	mg := gtx
	mg.Constraints = layout.Constraints{Max: image.Pt(1<<20, gtx.Constraints.Max.Y)}
	natural := 0
	for _, t := range tabs {
		w := elements.Measure(mg, giode.Text(t.Title, giode.Styles{FontSize: tabFontSize, MaxLines: 1})).Size.X
		natural += min(max(w+gtx.Dp(12+4+20+6), gtx.Dp(tabMin)), gtx.Dp(tabMax))
	}
	w := max(min(natural, room), 0)
	if !gtx.Source.Enabled() {
		// A measuring pass (input is off), as when the title bar sizes
		// its parts: laying the strip out here would use up its request
		// to scroll the selected tab into view.
		return layout.Dimensions{Size: image.Pt(w, gtx.Dp(tabHeight))}
	}
	gtx.Constraints.Min.X, gtx.Constraints.Max.X = w, w
	return u.strip.Layout(gtx)
}
