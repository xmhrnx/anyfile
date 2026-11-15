package ui

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"

	"github.com/xmhrnx/anyfile/internal/detect"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"
)

// Palette: a deep slate canvas with an indigo→cyan accent.
const (
	colBg        = "#0a0e16"
	colSidebar   = "#0e131d"
	colSurface   = "#121926"
	colSurface2  = "#172031"
	colRaised    = "#1c2639"
	colBorder    = "#222d42"
	colBorderHi  = "#2f3d58"
	colText      = "#e7edf7"
	colTextSoft  = "#b9c3d4"
	colMuted     = "#7d8aa3"
	colFaint     = "#56627a"
	colAccent    = "#7c83ff"
	colAccent2   = "#22d3ee"
	colAccentDim = "#7c83ff26"
	colSuccess   = "#34d399"
	colWarning   = "#fbbf24"
	colDanger    = "#f87171"
)

const (
	// fontUI is the interface font: Segoe UI on Windows, Noto Sans
	// elsewhere (loadFonts). The name is AnyFile's own so it never
	// matches a system font: one that isn't registered falls back to
	// the system's fonts before the bundled ones.
	fontUI   = "AnyFile UI"
	fontMono = "Go Mono"
)

// catColor gives every category its own hue.
func catColor(c detect.Category) string {
	switch c {
	case detect.CatArchive, detect.CatCompressed:
		return "#f5b041"
	case detect.CatDocument:
		return "#60a5fa"
	case detect.CatSpreadsheet:
		return "#34d399"
	case detect.CatPresentation:
		return "#fb923c"
	case detect.CatImage:
		return "#f472b6"
	case detect.CatAudio:
		return "#a78bfa"
	case detect.CatVideo:
		return "#f87171"
	case detect.CatExecutable:
		return "#22d3ee"
	case detect.CatFont:
		return "#2dd4bf"
	case detect.CatDatabase:
		return "#facc15"
	case detect.CatText:
		return "#cbd5e1"
	case detect.CatCode:
		return "#38bdf8"
	case detect.CatDiskImage:
		return "#c084fc"
	case detect.CatEmail:
		return "#4ade80"
	case detect.CatCrypto, detect.CatEncrypted:
		return "#fda4af"
	case "Folder":
		return "#94a3b8"
	}
	return "#8b98ad"
}

// notoFonts are Noto fonts, and Hind Siliguri for Bengali, cut down to
// the characters of AnyFile's translations (fonts/subset.py), so every
// language draws the same on every system: left to the system's
// fonts, Korean fell to a Windows font without real Hangul glyphs.
//
//go:embed fonts/*.ttf fonts/*.txt
var notoFonts embed.FS

// appLicenses is licensesJSON, which lists Go modules, with the
// bundled fonts' licenses added.
func appLicenses() []byte {
	var list []map[string]string
	if err := json.Unmarshal(licensesJSON, &list); err != nil {
		return licensesJSON
	}
	for _, f := range []struct{ name, file string }{
		{"Noto Sans fonts", "fonts/OFL-Noto.txt"},
		{"Hind Siliguri font", "fonts/OFL-HindSiliguri.txt"},
	} {
		if b, err := notoFonts.ReadFile(f.file); err == nil {
			list = append(list, map[string]string{"module": f.name, "type": "OFL-1.1", "text": string(b)})
		}
	}
	b, err := json.Marshal(list)
	if err != nil {
		return licensesJSON
	}
	return b
}

// notoOrder is the order the Noto fonts are tried in. Japanese and
// Chinese share most characters but draw them differently, so the
// interface language's comes first (loadFonts).
var notoOrder = []string{"NotoSans", "NotoSansSC", "NotoSansJP", "NotoSansKR", "NotoSansArabic",
	"NotoSansHebrew", "NotoSansDevanagari", "HindSiliguri", "NotoSansTamil", "NotoSansThai",
	"NotoSansArmenian", "NotoSansGeorgian"}

// loadFonts prefers the platform UI font for a native feel and keeps
// the Go fonts for monospace text and as a fallback, then the Noto
// fonts for the scripts those lack. lang is the interface language.
func loadFonts(lang string) {
	faces := []font.FontFace{}
	dir := filepath.Join(os.Getenv("WINDIR"), "Fonts")
	for _, f := range []struct {
		file   string
		weight font.Weight
		style  font.Style
	}{
		{"segoeui.ttf", font.Normal, font.Regular},
		{"seguisb.ttf", font.SemiBold, font.Regular},
		{"segoeuib.ttf", font.Bold, font.Regular},
		{"segoeuil.ttf", font.Light, font.Regular},
		{"segoeuii.ttf", font.Normal, font.Italic},
	} {
		b, err := os.ReadFile(filepath.Join(dir, f.file))
		if err != nil {
			continue
		}
		parsed, err := opentype.ParseCollection(b)
		if err != nil {
			continue
		}
		for _, p := range parsed {
			p.Font.Typeface = fontUI
			p.Font.Weight = f.weight
			p.Font.Style = f.style
			faces = append(faces, p)
		}
	}
	haveUI := len(faces) > 0
	faces = append(faces, gofont.Collection()...)
	order := notoOrder
	if strings.HasPrefix(lang, "ja") {
		order = append([]string{"NotoSansJP"}, notoOrder...)
	}
	seen := map[string]bool{}
	for _, name := range order {
		if seen[name] {
			continue
		}
		seen[name] = true
		for _, w := range []struct {
			style  string
			weight font.Weight
		}{{"Regular", font.Normal}, {"Bold", font.Bold}} {
			b, err := notoFonts.ReadFile("fonts/" + name + "-" + w.style + ".ttf")
			if err != nil {
				continue
			}
			parsed, err := opentype.ParseCollection(b)
			if err != nil {
				continue
			}
			for _, p := range parsed {
				// The prefix keeps the system's own Noto fonts (macOS
				// ships Noto Sans Armenian and Georgian) from being
				// picked instead: font names match ignoring spaces.
				p.Font.Typeface = font.Typeface("AnyFile " + name)
				p.Font.Weight = w.weight
				faces = append(faces, p)
				if name == "NotoSans" && !haveUI {
					p.Font.Typeface = fontUI
					faces = append(faces, p)
				}
			}
		}
	}
	giode.RegisterFonts(faces)
}

// Text style helpers.
func txt(size float32, color string) giode.Styles {
	return giode.Styles{FontSize: properties.FontSize(size), Color: properties.Color(color), FontFamily: fontUI}
}

func txtw(size float32, color string, w giode.Styles) giode.Styles {
	s := txt(size, color)
	return s.With(w)
}

// applyTheme switches giode to dark mode with AnyFile's palette, so
// components styled with theme tokens (the tree, toasts, copy buttons)
// match the hand-styled ones.
func applyTheme() {
	giode.SetPalette(giode.ModeDark, giode.Palette{
		Primary:    colAccent,
		Secondary:  colTextSoft,
		Background: colBg,
		Surface:    colRaised,
		Text:       colText,
		Border:     colBorderHi,
		OnPrimary:  "#ffffff",
		Muted:      colMuted,
		Accent:     colAccent2,
		Success:    colSuccess,
		Warning:    colWarning,
		Danger:     colDanger,
		Info:       colAccent2,
	})
	giode.SetMode(giode.ModeDark)
}

// color converts a plain string for a style field.
func color(s string) properties.Color { return properties.Color(s) }
