"""Cuts the fonts down to the characters AnyFile's interface uses.

    python subset.py <folder with the fonts from github.com/google/fonts>

The folder holds the Noto variable fonts (NotoSansKR.ttf, ...) and
HindSiliguri-Regular.ttf and -Bold.ttf: go-text shapes Noto Sans
Bengali's conjuncts wrongly, Hind Siliguri's correctly.

Needs fontTools. Writes a Regular and a Bold of each font here; run it
again after a translation adds characters (TestFontsCoverLocales says
when).
"""
import json, os, re, sys

from fontTools import subset
from fontTools.ttLib import TTFont
from fontTools.varLib import instancer

here = os.path.dirname(os.path.abspath(__file__))
root = os.path.join(here, "..", "..", "..")
src = sys.argv[1]

# Each font and the languages it draws; NotoSans covers the rest
# (Latin, Greek, Cyrillic, Vietnamese).
FONTS = {
    "NotoSansKR": ["ko"], "NotoSansJP": ["ja"], "NotoSansSC": ["zh"],
    "NotoSansArabic": ["ar", "fa", "ur"], "NotoSansHebrew": ["he"],
    "NotoSansDevanagari": ["hi"], "HindSiliguri": ["bn"], "NotoSansTamil": ["ta"],
    "NotoSansThai": ["th"], "NotoSansArmenian": ["hy"], "NotoSansGeorgian": ["ka"],
}
FONTS["NotoSans"] = None


def strings(v):
    if isinstance(v, str):
        yield v
    elif isinstance(v, dict):
        for k, x in v.items():
            yield k
            yield from strings(x)


base = set(chr(c) for c in range(0x20, 0x7F))
base |= set("…–—‘’“”«»•·×°±µ€£¥©®™←→↑↓✓")
loc = os.path.join(root, "internal", "locales")
by_lang = {}
for name in sorted(os.listdir(loc)):
    if name.endswith(".json"):
        with open(os.path.join(loc, name), encoding="utf-8") as f:
            by_lang[name[:-5]] = set("".join(strings(json.load(f))))
# The language menu shows every language's name in its own script.
with open(os.path.join(loc, "locales.go"), encoding="utf-8") as f:
    for lang, s in re.findall(r'\{"([a-z-]+)", "([^"]+)"\}', f.read()):
        by_lang.setdefault(lang, set()).update(s)


def text_for(langs):
    chars = set(base)
    for lang, cs in by_lang.items():
        if langs is None or lang in langs:
            chars |= cs
    return "".join(sorted(chars))


for name in FONTS:
    for style, wght in (("Regular", 400), ("Bold", 700)):
        static = os.path.join(src, f"{name}-{style}.ttf")
        if os.path.exists(static):
            font = TTFont(static)
        else:
            font = TTFont(os.path.join(src, name + ".ttf"))
            axes = {"wght": wght}
            if "wdth" in [a.axisTag for a in font["fvar"].axes]:
                axes["wdth"] = 100
            font = instancer.instantiateVariableFont(font, axes)
        opts = subset.Options()
        opts.layout_features = ["*"]
        opts.hinting = False
        opts.name_IDs = [0, 1, 2, 3, 4, 5, 6]
        sub = subset.Subsetter(opts)
        sub.populate(text=text_for(FONTS[name]))
        sub.subset(font)
        if not any(len(t.cmap) for t in font["cmap"].tables):
            continue
        out = os.path.join(here, f"{name}-{style}.ttf")
        font.save(out)
        print(f"{name}-{style}: {os.path.getsize(out) // 1024} KB")
