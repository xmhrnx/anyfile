// Package ui is AnyFile's interface, built with giode: a frameless
// window with one tab per opened file, each showing a content tree on
// the left and an inspector on the right. All state lives in the tabs'
// sessions, which belong to the UI goroutine; the UI renders them and
// turns clicks into session calls.
package ui

import (
	"context"
	"image"
	"log"
	"strings"
	"time"

	"gioui.org/layout"

	"github.com/xmhrnx/anyfile/internal/explore"
	"github.com/xmhrnx/anyfile/internal/oodle"
	"github.com/xmhrnx/anyfile/internal/platform"
	"github.com/xmhrnx/anyfile/internal/session"
	"github.com/xmhrnx/anyfile/internal/settings"
	"github.com/xmhrnx/anyfile/internal/tools"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/button"
	"github.com/xmhrnx/giode/components/titlebar"
)

// UI is the application window.
type UI struct {
	app    *giode.App
	bar    *titlebar.TitleBar
	toasts *giode.ToasterElement

	tabs      []*tab
	active    *tab
	strip     *giode.DocTabsElement
	inspector *inspector

	openBtn      *button.Button
	settingsBtn  *button.Button
	emptyOpenBtn *button.Button

	settings     *settingsPage
	updates      *updates
	engines      map[string]*engineInstall
	moreEngines  *giode.PressableElement
	showSettings bool
	dropZone     *giode.PressableElement
	version      string
}

// Run opens the window and blocks until it closes. A path argument is
// opened right away; version is shown in the status bar.
func Run(args []string, version string) error {
	applyTheme()
	settingsErr := settings.Load()
	loadFonts(locale(settings.Get()))
	u := &UI{version: version}
	cfg := giode.Config{
		Title:      "AnyFile",
		Size:       image.Pt(1280, 820),
		MinSize:    image.Pt(900, 560),
		Resizable:  true,
		Frameless:  true,
		Center:     true,
		Icon:       iconPNG,
		Background: colBg,
		LogLevel:   giode.LogLevelError,
		Logger:     log.Default(),
		Intl:       intlOptions(),
		Locale:     locale(settings.Get()),
		OnShutdown: func() {
			for _, t := range u.tabs {
				t.sess.Close()
			}
		},
	}
	if platform.FileDropSupported() {
		cfg.OnFileDrop = u.openPaths
	}
	// One window: files opened while AnyFile runs, from the command line
	// or the file manager's "Inspect with AnyFile", become tabs in it.
	cfg.SingleInstanceLock = &giode.SingleInstanceLock{
		UniqueID:     "com.xmhrnx.anyfile",
		OnLaunchArgs: func(args []string) { u.openPaths(fileArgs(args)) },
	}
	// Also in giode.toml, for the Windows installer and the macOS app;
	// declared here so builds without the giode CLI can register it.
	cfg.ContextActions = []giode.ContextAction{{ID: contextInspect, Label: tr("Inspect with AnyFile"), Files: true}}
	cfg.OnContextAction = func(id string, paths []string) {
		if id == contextInspect {
			u.openPaths(paths)
		}
	}
	u.app = giode.New(cfg)
	u.toasts = giode.Toaster(giode.Styles{Background: colRaised, BorderColor: colBorderHi, Color: colText, BorderRadius: 11}).
		Placement(giode.ToastBottomRight).Duration(5 * time.Second).Max(4).
		Inset(giode.Inset{Top: 44, Bottom: 28}) // clear of the title and status bars
	u.engines = newEngineInstalls(u)
	u.inspector = newInspector(u)
	u.updates = newUpdates(u)
	u.settings = newSettingsPage(u)
	u.newTabStrip()
	if settingsErr != nil {
		u.toast("error", tr("Could not read settings, using defaults: {err}", "err", settingsErr))
	}
	// Tool paths may have changed: look for the tools again.
	settings.Watch(func(settings.Settings) { tools.Reload() })
	u.settingsBtn = newBtn("settings", "Settings", btnGhost, func() { u.showSettings = !u.showSettings })
	u.openBtn = newBtn("open", "Open file", btnSecondary, u.pickFile)
	u.emptyOpenBtn = newBtn("open", "Choose a file", btnPrimary, u.pickFile)
	u.dropZone = giode.Pressable(u.dropZoneView).
		OnClick(func(giode.ClickEvent) { u.pickFile() })
	appText.add(func() { u.dropZone.Label(tr("Choose a file")) })
	u.moreEngines = giode.Pressable(func(s giode.PressState) giode.Element {
		c := colAccent
		if s.Hovered || s.Focused {
			c = colAccent2
		}
		return giode.Text(tr("More engines?"), txt(11.5, c))
	}).OnClick(func(giode.ClickEvent) {
		u.showSettings = true
		u.settings.scroll.ScrollTo(0)
	})
	appText.add(func() { u.moreEngines.Label(tr("More engines?")) })
	u.bar = giode.TitleBar(u.app, "AnyFile", giode.Styles{
		Background: colSidebar, Color: colTextSoft, Height: 44, FontSize: 13,
		BorderWidth: 1, BorderColor: colBorder,
	}).Icon(logo(22)).Leading(u.openBtn, rawEl(u.tabStrip)).Trailing(u.settingsBtn)
	// giode runs a context action itself (OnContextAction).
	if paths := fileArgs(args); len(paths) > 0 {
		u.app.Post(func() { u.openPaths(paths) })
	}
	if u.updates.enabled() && settings.Get().CheckUpdates {
		u.app.Post(func() { u.updates.run(false) })
	}
	return u.app.Run(u.view)
}

// contextInspect is the ID of the "Inspect with AnyFile" entry in the
// file manager's right-click menu (giode.toml's context_actions).
const contextInspect = "inspect"

// fileArgs returns the command-line arguments that name files to open:
// none for a context action, which giode delivers to OnContextAction,
// and no flags, including the /Embedding COM may start the app with when
// Explorer hands it a selection.
func fileArgs(args []string) []string {
	var out []string
	for _, a := range args {
		if strings.HasPrefix(a, "--giode-action=") {
			return nil
		}
		if !strings.HasPrefix(a, "-") && !strings.EqualFold(a, "/Embedding") {
			out = append(out, a)
		}
	}
	return out
}

// toast reports a session outcome; safe from any goroutine.
func (u *UI) toast(kind, msg string) {
	k := giode.ToastSuccess
	if kind == "error" {
		k = giode.ToastError
	}
	u.toasts.Show(k, msg)
}

// pickFile shows the system's file picker, which blocks, off the UI
// goroutine. Every chosen file opens in its own tab. giode's pickers,
// unlike a helper process, give a sandboxed (Mac App Store) build
// access to what the user chose.
func (u *UI) pickFile() {
	go func() {
		if paths := pickPaths(u.app, true); len(paths) > 0 {
			u.app.Post(func() { u.openPaths(paths) })
		}
	}()
}

// pickFolder asks for a folder, then runs fn with it on the UI goroutine.
func (u *UI) pickFolder(title string, fn func(dir string)) {
	go func() {
		if dir, err := u.app.PickFolder(context.Background(), giode.FolderOptions{Title: title}); err == nil && dir != "" {
			u.app.Post(func() { fn(dir) })
		}
	}()
}

// pickPaths shows the file picker and returns the paths chosen, none
// when the user cancelled; call it off the UI goroutine.
func pickPaths(app *giode.App, multiple bool) []string {
	files, err := app.PickFiles(context.Background(), giode.PickOptions{Multiple: multiple})
	if err != nil {
		return nil
	}
	var paths []string
	for _, f := range files {
		f.Close()
		if f.Path != "" {
			paths = append(paths, f.Path)
		}
	}
	return paths
}

func (u *UI) view() giode.Element {
	appText.update()
	giode.Shortcuts(
		giode.Shortcut("Ctrl+Tab", func() { u.strip.SelectNext(1) }),
		giode.Shortcut("Ctrl+Shift+Tab", func() { u.strip.SelectNext(-1) }),
		giode.Shortcut("Mod+W", func() {
			if u.active != nil && !u.showSettings {
				u.closeTab(u.active)
			}
		}),
	)
	var body giode.Element
	var root *session.Node
	if t := u.active; u.showSettings {
		body = u.settings.view()
	} else if t == nil {
		body = u.emptyState()
	} else {
		root = t.sess.Root()
		body = giode.Row(giode.Styles{Width: -1, Height: -1, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{Width: 400, Height: -1, Background: colSidebar}, t.tree.view()),
			giode.Box(giode.Styles{Width: 1, Height: -1, Background: colBorder}),
			giode.Box(giode.Styles{FlexGrow: 1, Width: -1, Height: -1}, u.inspector.view(t.sess.Selected())),
		)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: colBg},
		u.bar,
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, body),
		u.statusBar(root),
		u.toasts,
	)
}

// --- empty state ----------------------------------------------------------------

func (u *UI) emptyState() giode.Element {
	features := []struct{ ic, title, text string }{
		{"scan", "Identifies anything", "Signatures, structure probes and TrID, plus its own heuristics when those fail."},
		{"layers", "Explores inside", "Archives, Office files, PDFs, installers, mail, disc images, media streams, blobs."},
		{"external", "Opens in your apps", "Every item opens in the viewer your system uses for its real type."},
	}
	if platform.Sandboxed() {
		// The Mac App Store build has no TrID, and no FFmpeg for media
		// streams.
		features[0].text = "Signatures and structure probes, plus its own heuristics when those fail."
		features[1].text = "Archives, Office files, PDFs, installers, mail, disc images, blobs."
	}
	var cards []giode.Element
	for _, f := range features {
		f.title, f.text = tr(f.title), tr(f.text)
		cards = append(cards, giode.Box(giode.Styles{
			FlexDirection: giode.FlexDirectionColumn, Gap: 8, Width: 230, Height: 200,
			Padding: giode.UniformInset(16), Background: colSurface, BorderRadius: 12,
			BorderWidth: 1, BorderColor: colBorder,
		},
			giode.Box(giode.Styles{Padding: giode.UniformInset(8), Background: colAccentDim, BorderRadius: 9}, icon(f.ic, colAccent, 18)),
			giode.Text(f.title, txtw(14, colText, giode.Styles{FontWeight: giode.FontWeightBold})),
			giode.Text(f.text, txtw(12.5, colMuted, giode.Styles{LineHeight: 18})),
		))
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Align: giode.AlignCenter, Justify: giode.JustifyCenter, Gap: 28},
		u.dropZone,
		giode.Row(giode.Styles{Gap: 16, Justify: giode.JustifyCenter, Wrap: giode.WrapNoWrap}, cards...),
	)
}

func (u *UI) dropZoneView(s giode.PressState) giode.Element {
	border, bg := colBorderHi, colSurface
	if s.Hovered || s.Focused {
		border, bg = colAccent, "#151c2c"
	}
	hint := tr("or choose one, and AnyFile will tell you what it really is and what's inside")
	title := tr("Drop any file here")
	if !platform.FileDropSupported() {
		title, hint = tr("Open any file"), tr("AnyFile will tell you what it really is and what's inside")
	}
	return giode.Box(giode.Styles{
		FlexDirection: giode.FlexDirectionColumn, Align: giode.AlignCenter, Justify: giode.JustifyCenter,
		Gap: 14, Width: 720, Height: 260, Background: bg, BorderRadius: 20,
		BorderWidth: 2, BorderColor: border,
	},
		giode.Box(giode.Styles{Padding: giode.UniformInset(16), BorderRadius: 18,
			BackgroundGradient: giode.Linear(135, "#7c83ff33", "#22d3ee22")}, icon("upload", colAccent2, 34)),
		giode.Text(title, txtw(22, colText, giode.Styles{FontWeight: giode.FontWeightBold})),
		giode.Text(hint, txt(13.5, colMuted)),
		u.emptyOpenBtn,
	)
}

// --- status bar -------------------------------------------------------------------

func (u *UI) statusBar(root *session.Node) giode.Element {
	status := func(name string, ok bool) giode.Element {
		c, ic := colSuccess, "check"
		if !ok {
			c, ic = colFaint, "x"
		}
		return giode.Row(giode.Styles{Gap: 4, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			icon(ic, c, 11), giode.Text(name, txt(11.5, c)))
	}
	engines := []giode.Element{giode.Text(tr("Engines"), txt(11.5, colFaint))}
	// The Mac App Store build has only the built-in engines.
	if !platform.Sandboxed() {
		for _, t := range tools.All() {
			engines = append(engines, status(t.Name, t.Found()))
		}
		// Oodle comes with games, so whether it's available depends on
		// where the opened file lives.
		near := ""
		if root != nil {
			near = root.Local
		}
		engines = append(engines, status("Oodle", oodle.Find(near) != ""))
	}
	engines = append(engines, status(tr("Built-in (signatures, carver, {count} engines)", "count", explore.EngineCount()), true))
	if platform.Sandboxed() {
		engines = append(engines, u.moreEngines)
	}
	right := giode.Text("AnyFile "+u.version+" · "+tr("Drop a file anywhere in the window"), txt(11.5, colFaint))
	if !platform.FileDropSupported() {
		right = giode.Text("AnyFile "+u.version+" · "+tr("Use Open file to inspect a file"), txt(11.5, colFaint))
	}
	if root != nil {
		right = giode.Text(root.Local, txtw(11.5, colFaint, giode.Styles{MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}))
	}
	return giode.Row(giode.Styles{
		Width: -1, Height: 28, Align: giode.AlignCenter, Justify: giode.JustifySpaceBetween,
		Padding: giode.SymmetricInset(14, 0), Background: colSidebar, Wrap: giode.WrapNoWrap,
	},
		giode.Row(giode.Styles{Gap: 14, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, engines...),
		giode.Box(giode.Styles{MaxWidth: 520}, right),
	)
}

// rawEl is an element that lays itself out.
type rawEl func(gtx layout.Context) layout.Dimensions

func (r rawEl) Layout(gtx layout.Context) layout.Dimensions { return r(gtx) }
