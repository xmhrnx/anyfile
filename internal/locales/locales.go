// Package locales holds AnyFile's translations: one <locale>.json per
// language, in giode's message format. The keys are the English text,
// so anything a language lacks shows in English; en.json only holds
// English plural forms.
package locales

import "embed"

// FS holds the message files.
//
//go:embed *.json
var FS embed.FS

// Language is a translation the settings page offers.
type Language struct {
	Locale string
	Name   string // in the language itself
}

// Languages are the interface languages, sorted by their own names.
var Languages = []Language{
	{"id", "Bahasa Indonesia"},
	{"ms", "Bahasa Melayu"},
	{"ca", "Català"},
	{"cs", "Čeština"},
	{"da", "Dansk"},
	{"de", "Deutsch"},
	{"et", "Eesti"},
	{"en", "English"},
	{"es", "Español"},
	{"fil", "Filipino"},
	{"fr", "Français"},
	{"hr", "Hrvatski"},
	{"it", "Italiano"},
	{"sw", "Kiswahili"},
	{"lv", "Latviešu"},
	{"lt", "Lietuvių"},
	{"hu", "Magyar"},
	{"nl", "Nederlands"},
	{"nb", "Norsk bokmål"},
	{"pl", "Polski"},
	{"pt", "Português"},
	{"ro", "Română"},
	{"sk", "Slovenčina"},
	{"sl", "Slovenščina"},
	{"fi", "Suomi"},
	{"sv", "Svenska"},
	{"vi", "Tiếng Việt"},
	{"tr", "Türkçe"},
	{"el", "Ελληνικά"},
	{"bg", "Български"},
	{"ru", "Русский"},
	{"uk", "Українська"},
	{"ar", "العربية"},
	{"fa", "فارسی"},
	{"ur", "اردو"},
	{"hy", "Հայերեն"},
	{"bn", "বাংলা"},
	{"ka", "ქართული"},
	{"hi", "हिन्दी"},
	{"he", "עברית"},
	{"ta", "தமிழ்"},
	{"th", "ไทย"},
	{"ja", "日本語"},
	{"ko", "한국어"},
	{"zh", "中文"},
}
