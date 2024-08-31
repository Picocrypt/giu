package giu

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Picocrypt/imgui-go"
)

var (
	shouldRebuildFontAtlas bool
	stringMap              sync.Map // key is rune, value indicates whether it's a new rune.
	defaultFonts           []FontInfo
	extraFonts             []FontInfo
	extraFontMap           map[string]*imgui.Font
)

const (
	preRegisterString = "\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
)

// FontInfo represents a the font.
//
type FontInfo struct {
	fontName string
	fontPath string
	fontByte []byte
	size     float32
}

func (f *FontInfo) String() string {
	return fmt.Sprintf("%s:%.2f", f.fontName, f.size)
}

func init() {
	extraFontMap = make(map[string]*imgui.Font)

	// Pre register numbers
	tStr(preRegisterString)
}

// SetDefaultFontFromBytes changes default font by bytes of the font file.
func SetDefaultFontFromBytes(fontBytes []byte, size float32) {
	defaultFonts = append([]FontInfo{
		{
			fontByte: fontBytes,
			size:     size,
		},
	}, defaultFonts...)
}

// Register string to font atlas builder.
// Note only register strings that will be displayed on the UI.
func tStr(str string) string {
	for _, s := range str {
		if _, ok := stringMap.Load(s); !ok {
			stringMap.Store(s, false)
			shouldRebuildFontAtlas = true
		}
	}

	return str
}

// Register string pointer to font atlas builder.
// Note only register strings that will be displayed on the UI.
func tStrPtr(str *string) *string {
	tStr(*str)
	return str
}

func tStrSlice(str []string) []string {
	for _, s := range str {
		tStr(s)
	}

	return str
}

// Rebuild font atlas when necessary.
func rebuildFontAtlas() {
	if !shouldRebuildFontAtlas {
		return
	}

	fonts := Context.IO().Fonts()
	fonts.Clear()

	var sb strings.Builder

	stringMap.Range(func(k, v interface{}) bool {
		stringMap.Store(k, true)
		if ks, ok := k.(rune); ok {
			sb.WriteRune(ks)
		}

		return true
	})

	ranges := imgui.NewGlyphRanges()
	builder := imgui.NewFontGlyphRangesBuilder()

	// Because we pre-regestered numbers, so default string map's length should greater then 11.
	if sb.Len() > len(preRegisterString) {
		builder.AddText(sb.String())
	} else {
		builder.AddRanges(fonts.GlyphRangesDefault())
	}

	builder.BuildRanges(ranges)

	if len(defaultFonts) > 0 {
		fontConfig := imgui.NewFontConfig()
		fontConfig.SetOversampleH(2)
		fontConfig.SetOversampleV(2)
		fontConfig.SetRasterizerMultiply(1.5)

		for i, fontInfo := range defaultFonts {
			if i > 0 {
				fontConfig.SetMergeMode(true)
			}

			if len(fontInfo.fontByte) == 0 {
				fonts.AddFontFromFileTTFV(fontInfo.fontPath, fontInfo.size, fontConfig, ranges.Data())
			} else {
				fonts.AddFontFromMemoryTTFV(fontInfo.fontByte, fontInfo.size, fontConfig, ranges.Data())
			}
		}

		// Fall back if no font is added
		if fonts.GetFontCount() == 0 {
			fonts.AddFontDefault()
		}
	} else {
		fonts.AddFontDefault()
	}

	// Add extra fonts
	for _, fontInfo := range extraFonts {
		// Store imgui.Font for PushFont
		var f imgui.Font
		if len(fontInfo.fontByte) == 0 {
			f = fonts.AddFontFromFileTTFV(fontInfo.fontPath, fontInfo.size, imgui.DefaultFontConfig, ranges.Data())
		} else {
			f = fonts.AddFontFromMemoryTTFV(fontInfo.fontByte, fontInfo.size, imgui.DefaultFontConfig, ranges.Data())
		}
		extraFontMap[fontInfo.String()] = &f
	}

	fontTextureImg := fonts.TextureDataRGBA32()
	Context.renderer.SetFontTexture(fontTextureImg)

	shouldRebuildFontAtlas = false
}
