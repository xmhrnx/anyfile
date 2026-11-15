package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"unicode"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
	gotext "github.com/go-text/typesetting/font"

	"github.com/xmhrnx/anyfile/internal/locales"
)

// TestFontsCoverLocales checks that the Go fonts and the bundled Noto
// fonts draw every character of every translation and language name,
// without help from the system's fonts. Run fonts/subset.py when it
// fails.
func TestFontsCoverLocales(t *testing.T) {
	var faces []*gotext.Face
	add := func(ff []font.FontFace) {
		for _, f := range ff {
			if g, ok := f.Face.(interface{ Face() *gotext.Face }); ok {
				faces = append(faces, g.Face())
			}
		}
	}
	add(gofont.Collection())
	ents, err := notoFonts.ReadDir("fonts")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range ents {
		if !strings.HasSuffix(e.Name(), ".ttf") {
			continue
		}
		b, _ := notoFonts.ReadFile("fonts/" + e.Name())
		ff, err := opentype.ParseCollection(b)
		if err != nil {
			t.Fatalf("%s: %v", e.Name(), err)
		}
		add(ff)
	}
	covered := func(r rune) bool {
		for _, f := range faces {
			if _, ok := f.NominalGlyph(r); ok {
				return true
			}
		}
		return false
	}

	texts := map[string][]string{}
	files, _ := locales.FS.ReadDir(".")
	for _, f := range files {
		b, _ := locales.FS.ReadFile(f.Name())
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("%s: %v", f.Name(), err)
		}
		lang := strings.TrimSuffix(f.Name(), ".json")
		var walk func(v any)
		walk = func(v any) {
			switch x := v.(type) {
			case string:
				texts[lang] = append(texts[lang], x)
			case map[string]any:
				for _, y := range x {
					walk(y)
				}
			}
		}
		walk(m)
	}
	for _, l := range locales.Languages {
		texts[l.Locale] = append(texts[l.Locale], l.Name)
	}
	for lang, ss := range texts {
		missing := map[rune]bool{}
		for _, s := range ss {
			// Section headings are shown in capitals.
			for _, r := range s + strings.ToUpper(s) {
				// Joiners and marks shape with their neighbours.
				if r <= ' ' || unicode.In(r, unicode.Cf) || covered(r) {
					continue
				}
				missing[r] = true
			}
		}
		if len(missing) > 0 {
			var list []string
			for r := range missing {
				list = append(list, fmt.Sprintf("%c (U+%04X)", r, r))
			}
			sort.Strings(list)
			t.Errorf("%s: no bundled font has %s", lang, strings.Join(list, ", "))
		}
	}
}
