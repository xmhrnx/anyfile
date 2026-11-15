package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/xmhrnx/anyfile/internal/platform"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/button"
	"github.com/xmhrnx/giode/update"
)

const (
	// releasesURL is the page with the latest downloads.
	releasesURL = "https://github.com/xmhrnx/anyfile/releases/latest"
	latestAPI   = "https://api.github.com/repos/xmhrnx/anyfile/releases/latest"
)

type updateState int

const (
	updIdle updateState = iota
	updChecking
	updLatest
	updAvailable
	updFailed
)

// updates checks GitHub for a newer release. Copies from the Mac App
// Store or the Microsoft Store are updated by the store, so they don't
// check.
type updates struct {
	u      *UI
	state  updateState
	latest string // the newest release's tag
	err    string

	check, get *button.Button
}

func newUpdates(u *UI) *updates {
	up := &updates{u: u}
	up.check = newBtn("reset", "Check now", btnSecondary, func() { up.run(true) })
	up.get = newBtn("download", "Download", btnPrimary, func() { u.openURL(releasesURL) })
	return up
}

// enabled reports whether this copy updates itself from GitHub rather
// than through a store.
func (up *updates) enabled() bool {
	return !platform.Sandboxed() && !platform.Packaged()
}

// run asks GitHub for the latest release. A check the user didn't ask
// for stays quiet unless it finds one.
func (up *updates) run(manual bool) {
	if up.state == updChecking {
		return
	}
	up.state, up.err = updChecking, ""
	go func() {
		tag, err := latestRelease()
		up.u.app.Post(func() {
			switch {
			case err != nil:
				up.state, up.err = updFailed, err.Error()
			case update.Newer(tag, up.u.version):
				up.state, up.latest = updAvailable, tag
				if !manual {
					up.u.toast("ok", tr("AnyFile {version} is available. Download it from Settings.", "version", tag))
				}
			default:
				up.state, up.latest = updLatest, tag
			}
		})
	}()
}

// latestRelease returns the tag of the newest GitHub release.
func latestRelease() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "AnyFile")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub answered %s", resp.Status)
	}
	var rel struct {
		Tag string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	if rel.Tag == "" {
		return "", errors.New("GitHub listed no release")
	}
	return rel.Tag, nil
}

// view is the Updates card's status line and buttons.
func (up *updates) view() giode.Element {
	var line string
	c := colMuted
	switch up.state {
	case updChecking:
		line = tr("Checking for updates…")
	case updLatest:
		line, c = tr("You have the latest version."), colSuccess
	case updAvailable:
		line, c = tr("AnyFile {version} is available.", "version", up.latest), colAccent2
	case updFailed:
		line, c = tr("Could not check for updates: {err}", "err", up.err), colDanger
	}
	left := []giode.Element{giode.Text(tr("You have AnyFile {version}.", "version", up.u.version), txtw(13.5, colText, giode.Styles{FontWeight: giode.FontWeightMedium}))}
	if line != "" {
		left = append(left, giode.Text(line, txtw(12, c, giode.Styles{LineHeight: 17})))
	}
	btns := []giode.Element{up.check}
	if up.state == updAvailable {
		btns = append(btns, up.get)
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 18, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(0, 4)},
		giode.Stack(giode.Styles{FlexGrow: 1, Width: -1, Gap: 3}, left...),
		giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, btns...),
	)
}

// openURL opens a web page in the browser, off the UI goroutine.
func (u *UI) openURL(url string) {
	go func() {
		if err := u.app.OpenURL(context.Background(), url); err != nil {
			u.app.Post(func() { u.toast("error", tr("Could not open {url}: {err}", "url", url, "err", err)) })
		}
	}()
}
