// Package color provides a simple set of typical 62 basic colors for use in UI
// color specification.
//
// Provides the complete Material Design 2014 palette by the same name scheme it
// uses plus RGBA and NRGBA methods to convert them with optional alpha value to
// derive standard 32-bit color values for UI elements.
package color

import (
	"image/color"
)

const (
	R = iota
	G
	B
)

// C is a string constant assumed to represent an RGB value
type C st

// NRGBA returns the corresponding color with an optional alpha value,
// otherwise it is set to 0xff (no alpha).
func (c C) NRGBA(A ...byte) color.NRGBA {
	if len(A) > 1 {
		return color.NRGBA{R: c[R], G: c[G], B: c[B], A: c[A[0]]}
	}
	return color.NRGBA{R: c[R], G: c[G], B: c[B], A: 0xff}
}

// RGBA returns the corresponding color with an optional alpha value,
// otherwise it is set to 0xff (no alpha).
func (c C) RGBA(A ...byte) color.RGBA {
	if len(A) > 1 {
		return color.RGBA{R: c[R], G: c[G], B: c[B], A: c[A[0]]}
	}
	return color.RGBA{R: c[R], G: c[G], B: c[B], A: c[0xff]}
}

const (
	Red50          C = "\xFF\xEB\xEE"
	Red100         C = "\xFF\xCD\xD2"
	Red200         C = "\xEF\x9A\x9A"
	Red300         C = "\xE5\x73\x73"
	Red400         C = "\xEF\x53\x50"
	Red500         C = "\xF4\x43\x36"
	Red600         C = "\xE5\x39\x35"
	Red700         C = "\xD3\x2F\x2F"
	Red800         C = "\xC6\x28\x28"
	Red900         C = "\xB7\x1C\x1C"
	RedA100        C = "\xFF\x8A\x80"
	RedA200        C = "\xFF\x52\x52"
	RedA400        C = "\xFF\x17\x44"
	RedA700        C = "\xD5\x00\x00"
	Pink50         C = "\xFC\xE4\xEC"
	Pink100        C = "\xF8\xBB\xD0"
	Pink200        C = "\xF4\x8F\xB1"
	Pink300        C = "\xF0\x62\x92"
	Pink400        C = "\xEC\x40\x7A"
	Pink500        C = "\xE9\x1E\x63"
	Pink600        C = "\xD8\x1B\x60"
	Pink700        C = "\xC2\x18\x5B"
	Pink800        C = "\xAD\x14\x57"
	Pink900        C = "\x88\x0E\x4F"
	PinkA100       C = "\xFF\x80\xAB"
	PinkA200       C = "\xFF\x40\x81"
	PinkA400       C = "\xF5\x00\x57"
	PinkA700       C = "\xC5\x11\x62"
	Purple50       C = "\xF3\xE5\xF5"
	Purple100      C = "\xE1\xBE\xE7"
	Purple200      C = "\xCE\x93\xD8"
	Purple300      C = "\xBA\x68\xC8"
	Purple400      C = "\xAB\x47\xBC"
	Purple500      C = "\x9C\x27\xB0"
	Purple600      C = "\x8E\x24\xAA"
	Purple700      C = "\x7B\x1F\xA2"
	Purple800      C = "\x6A\x1B\x9A"
	Purple900      C = "\x4A\x14\x8C"
	PurpleA100     C = "\xEA\x80\xFC"
	PurpleA200     C = "\xE0\x40\xFB"
	PurpleA400     C = "\xD5\x00\xF9"
	PurpleA700     C = "\xAA\x00\xFF"
	DeepPurple50   C = "\xED\xE7\xF6"
	DeepPurple100  C = "\xD1\xC4\xE9"
	DeepPurple200  C = "\xB3\x9D\xDB"
	DeepPurple300  C = "\x95\x75\xCD"
	DeepPurple400  C = "\x7E\x57\xC2"
	DeepPurple500  C = "\x67\x3A\xB7"
	DeepPurple600  C = "\x5E\x35\xB1"
	DeepPurple700  C = "\x51\x2D\xA8"
	DeepPurple800  C = "\x45\x27\xA0"
	DeepPurple900  C = "\x31\x1B\x92"
	DeepPurpleA100 C = "\xB3\x88\xFF"
	DeepPurpleA200 C = "\x7C\x4D\xFF"
	DeepPurpleA400 C = "\x65\x1F\xFF"
	DeepPurpleA700 C = "\x62\x00\xEA"
	Indigo50       C = "\xE8\xEA\xF6"
	Indigo100      C = "\xC5\xCA\xE9"
	Indigo200      C = "\x9F\xA8\xDA"
	Indigo300      C = "\x79\x86\xCB"
	Indigo400      C = "\x5C\x6B\xC0"
	Indigo500      C = "\x3F\x51\xB5"
	Indigo600      C = "\x39\x49\xAB"
	Indigo700      C = "\x30\x3F\x9F"
	Indigo800      C = "\x28\x35\x93"
	Indigo900      C = "\x1A\x23\x7E"
	IndigoA100     C = "\x8C\x9E\xFF"
	IndigoA200     C = "\x53\x6D\xFE"
	IndigoA400     C = "\x3D\x5A\xFE"
	IndigoA700     C = "\x30\x4F\xFE"
	Blue50         C = "\xE3\xF2\xFD"
	Blue100        C = "\xBB\xDE\xFB"
	Blue200        C = "\x90\xCA\xF9"
	Blue300        C = "\x64\xB5\xF6"
	Blue400        C = "\x42\xA5\xF5"
	Blue500        C = "\x21\x96\xF3"
	Blue600        C = "\x1E\x88\xE5"
	Blue700        C = "\x19\x76\xD2"
	Blue800        C = "\x15\x65\xC0"
	Blue900        C = "\x0D\x47\xA1"
	BlueA100       C = "\x82\xB1\xFF"
	BlueA200       C = "\x44\x8A\xFF"
	BlueA400       C = "\x29\x79\xFF"
	BlueA700       C = "\x29\x62\xFF"
	LightBlue50    C = "\xE1\xF5\xFE"
	LightBlue100   C = "\xB3\xE5\xFC"
	LightBlue200   C = "\x81\xD4\xFA"
	LightBlue300   C = "\x4F\xC3\xF7"
	LightBlue400   C = "\x29\xB6\xF6"
	LightBlue500   C = "\x03\xA9\xF4"
	LightBlue600   C = "\x03\x9B\xE5"
	LightBlue700   C = "\x02\x88\xD1"
	LightBlue800   C = "\x02\x77\xBD"
	LightBlue900   C = "\x01\x57\x9B"
	LightBlueA100  C = "\x80\xD8\xFF"
	LightBlueA200  C = "\x40\xC4\xFF"
	LightBlueA400  C = "\x00\xB0\xFF"
	LightBlueA700  C = "\x00\x91\xEA"
	Cyan50         C = "\xE0\xF7\xFA"
	Cyan100        C = "\xB2\xEB\xF2"
	Cyan200        C = "\x80\xDE\xEA"
	Cyan300        C = "\x4D\xD0\xE1"
	Cyan400        C = "\x26\xC6\xDA"
	Cyan500        C = "\x00\xBC\xD4"
	Cyan600        C = "\x00\xAC\xC1"
	Cyan700        C = "\x00\x97\xA7"
	Cyan800        C = "\x00\x83\x8F"
	Cyan900        C = "\x00\x60\x64"
	CyanA100       C = "\x84\xFF\xFF"
	CyanA200       C = "\x18\xFF\xFF"
	CyanA400       C = "\x00\xE5\xFF"
	CyanA700       C = "\x00\xB8\xD4"
	Teal50         C = "\xE0\xF2\xF1"
	Teal100        C = "\xB2\xDF\xDB"
	Teal200        C = "\x80\xCB\xC4"
	Teal300        C = "\x4D\xB6\xAC"
	Teal400        C = "\x26\xA6\x9A"
	Teal500        C = "\x00\x96\x88"
	Teal600        C = "\x00\x89\x7B"
	Teal700        C = "\x00\x79\x6B"
	Teal800        C = "\x00\x69\x5C"
	Teal900        C = "\x00\x4D\x40"
	TealA100       C = "\xA7\xFF\xEB"
	TealA200       C = "\x64\xFF\xDA"
	TealA400       C = "\x1D\xE9\xB6"
	TealA700       C = "\x00\xBF\xA5"
	Green50        C = "\xE8\xF5\xE9"
	Green100       C = "\xC8\xE6\xC9"
	Green200       C = "\xA5\xD6\xA7"
	Green300       C = "\x81\xC7\x84"
	Green400       C = "\x66\xBB\x6A"
	Green500       C = "\x4C\xAF\x50"
	Green600       C = "\x43\xA0\x47"
	Green700       C = "\x38\x8E\x3C"
	Green800       C = "\x2E\x7D\x32"
	Green900       C = "\x1B\x5E\x20"
	GreenA100      C = "\xB9\xF6\xCA"
	GreenA200      C = "\x69\xF0\xAE"
	GreenA400      C = "\x00\xE6\x76"
	GreenA700      C = "\x00\xC8\x53"
	LightGreen50   C = "\xF1\xF8\xE9"
	LightGreen100  C = "\xDC\xED\xC8"
	LightGreen200  C = "\xC5\xE1\xA5"
	LightGreen300  C = "\xAE\xD5\x81"
	LightGreen400  C = "\x9C\xCC\x65"
	LightGreen500  C = "\x8B\xC3\x4A"
	LightGreen600  C = "\x7C\xB3\x42"
	LightGreen700  C = "\x68\x9F\x38"
	LightGreen800  C = "\x55\x8B\x2F"
	LightGreen900  C = "\x33\x69\x1E"
	LightGreenA100 C = "\xCC\xFF\x90"
	LightGreenA200 C = "\xB2\xFF\x59"
	LightGreenA400 C = "\x76\xFF\x03"
	LightGreenA700 C = "\x64\xDD\x17"
	Lime50         C = "\xF9\xFB\xE7"
	Lime100        C = "\xF0\xF4\xC3"
	Lime200        C = "\xE6\xEE\x9C"
	Lime300        C = "\xDC\xE7\x75"
	Lime400        C = "\xD4\xE1\x57"
	Lime500        C = "\xCD\xDC\x39"
	Lime600        C = "\xC0\xCA\x33"
	Lime700        C = "\xAF\xB4\x2B"
	Lime800        C = "\x9E\x9D\x24"
	Lime900        C = "\x82\x77\x17"
	LimeA100       C = "\xF4\xFF\x81"
	LimeA200       C = "\xEE\xFF\x41"
	LimeA400       C = "\xC6\xFF\x00"
	LimeA700       C = "\xAE\xEA\x00"
	Yellow50       C = "\xFF\xFD\xE7"
	Yellow100      C = "\xFF\xF9\xC4"
	Yellow200      C = "\xFF\xF5\x9D"
	Yellow300      C = "\xFF\xF1\x76"
	Yellow400      C = "\xFF\xEE\x58"
	Yellow500      C = "\xFF\xEB\x3B"
	Yellow600      C = "\xFD\xD8\x35"
	Yellow700      C = "\xFB\xC0\x2D"
	Yellow800      C = "\xF9\xA8\x25"
	Yellow900      C = "\xF5\x7F\x17"
	YellowA100     C = "\xFF\xFF\x8D"
	YellowA200     C = "\xFF\xFF\x00"
	YellowA400     C = "\xFF\xEA\x00"
	YellowA700     C = "\xFF\xD6\x00"
	Amber50        C = "\xFF\xF8\xE1"
	Amber100       C = "\xFF\xEC\xB3"
	Amber200       C = "\xFF\xE0\x82"
	Amber300       C = "\xFF\xD5\x4F"
	Amber400       C = "\xFF\xCA\x28"
	Amber500       C = "\xFF\xC1\x07"
	Amber600       C = "\xFF\xB3\x00"
	Amber700       C = "\xFF\xA0\x00"
	Amber800       C = "\xFF\x8F\x00"
	Amber900       C = "\xFF\x6F\x00"
	AmberA100      C = "\xFF\xE5\x7F"
	AmberA200      C = "\xFF\xD7\x40"
	AmberA400      C = "\xFF\xC4\x00"
	AmberA700      C = "\xFF\xAB\x00"
	Orange50       C = "\xFF\xF3\xE0"
	Orange100      C = "\xFF\xE0\xB2"
	Orange200      C = "\xFF\xCC\x80"
	Orange300      C = "\xFF\xB7\x4D"
	Orange400      C = "\xFF\xA7\x26"
	Orange500      C = "\xFF\x98\x00"
	Orange600      C = "\xFB\x8C\x00"
	Orange700      C = "\xF5\x7C\x00"
	Orange800      C = "\xEF\x6C\x00"
	Orange900      C = "\xE6\x51\x00"
	OrangeA100     C = "\xFF\xD1\x80"
	OrangeA200     C = "\xFF\xAB\x40"
	OrangeA400     C = "\xFF\x91\x00"
	OrangeA700     C = "\xFF\x6D\x00"
	DeepOrange50   C = "\xFB\xE9\xE7"
	DeepOrange100  C = "\xFF\xCC\xBC"
	DeepOrange200  C = "\xFF\xAB\x91"
	DeepOrange300  C = "\xFF\x8A\x65"
	DeepOrange400  C = "\xFF\x70\x43"
	DeepOrange500  C = "\xFF\x57\x22"
	DeepOrange600  C = "\xF4\x51\x1E"
	DeepOrange700  C = "\xE6\x4A\x19"
	DeepOrange800  C = "\xD8\x43\x15"
	DeepOrange900  C = "\xBF\x36\x0C"
	DeepOrangeA100 C = "\xFF\x9E\x80"
	DeepOrangeA200 C = "\xFF\x6E\x40"
	DeepOrangeA400 C = "\xFF\x3D\x00"
	DeepOrangeA700 C = "\xDD\x2C\x00"
	Brown50        C = "\xEF\xEB\xE9"
	Brown100       C = "\xD7\xCC\xC8"
	Brown200       C = "\xBC\xAA\xA4"
	Brown300       C = "\xA1\x88\x7F"
	Brown400       C = "\x8D\x6E\x63"
	Brown500       C = "\x79\x55\x48"
	Brown600       C = "\x6D\x4C\x41"
	Brown700       C = "\x5D\x40\x37"
	Brown800       C = "\x4E\x34\x2E"
	Brown900       C = "\x3E\x27\x23"
	Gray50         C = "\xFA\xFA\xFA"
	Gray100        C = "\xF5\xF5\xF5"
	Gray200        C = "\xEE\xEE\xEE"
	Gray300        C = "\xE0\xE0\xE0"
	Gray400        C = "\xBD\xBD\xBD"
	Gray500        C = "\x9E\x9E\x9E"
	Gray600        C = "\x75\x75\x75"
	Gray700        C = "\x61\x61\x61"
	Gray800        C = "\x42\x42\x42"
	Gray900        C = "\x21\x21\x21"
	BlueGray50     C = "\xEC\xEF\xF1"
	BlueGray100    C = "\xCF\xD8\xDC"
	BlueGray200    C = "\xB0\xBE\xC5"
	BlueGray300    C = "\x90\xA4\xAE"
	BlueGray400    C = "\x78\x90\x9C"
	BlueGray500    C = "\x60\x7D\x8B"
	BlueGray600    C = "\x54\x6E\x7A"
	BlueGray700    C = "\x45\x5A\x64"
	BlueGray800    C = "\x37\x47\x4F"
	BlueGray900    C = "\x26\x32\x38"
	Black          C = "\x00\x00\x00"
	White          C = "\xFF\xFF\xFF"
)

type Element no

// UI element names, these are grouped according to purpose and usage in a UI
//
// Primary and Secondary are the accent colours, Primary should be the main used
// for header bars and "important" buttons. Secondary is where the intent is
// similar but for contrast. These colors should relate to branding.
//
// Doc* colors are for bodies of text, the majority of textual elements should
// use this.
//
// Panel* colors are for special areas where controls are located, like button
// bars and menus.
//
// The last 5 are essentially threat level style colors, usually colors of the
// rainbow in the order shown, red, orange, yellow, green, blue.
const (
	Primary Element = iota
	Secondary
	PrimaryText
	SecondaryText

	DocText
	DocBg
	DocTextDim
	DocBgDim
	DocBgHighlight

	PanelText
	PanelBg
	PanelTextDim
	PanelBgDim

	Fatal
	Warning
	Check
	Success
	Info
	last
)

const (
	Day   = 0
	Night = 1
)

type Theme [2][last]C

// GetDefaultTheme returns the standard realy GUI theme, based on the theme
// colors of Nostr and Bitcoin (purple and orange).
func GetDefaultTheme() (t Theme) {
	return Theme{
		Day: {
			Primary:        Purple500,
			Secondary:      Orange500,
			PrimaryText:    White,
			SecondaryText:  Black,
			DocText:        Gray800,
			DocBg:          Gray100,
			DocTextDim:     Gray600,
			DocBgDim:       Gray300,
			DocBgHighlight: White,
			PanelText:      Black,
			PanelBg:        Gray300,
			PanelTextDim:   Gray700,
			PanelBgDim:     Gray100,
			Fatal:          Red500,
			Warning:        Orange500,
			Check:          Yellow500,
			Success:        Green500,
			Info:           Blue500,
		},
		Night: {
			Primary:        DeepPurple500,
			Secondary:      DeepOrange500,
			PrimaryText:    Black,
			SecondaryText:  White,
			DocText:        Gray200,
			DocBg:          Black,
			DocTextDim:     Gray400,
			DocBgDim:       Gray800,
			DocBgHighlight: Black,
			PanelText:      White,
			PanelBg:        Gray700,
			PanelTextDim:   Gray300,
			PanelBgDim:     Gray900,
			Fatal:          Red700,
			Warning:        Orange700,
			Check:          Yellow700,
			Success:        Green700,
			Info:           Blue700,
		},
	}
}

// ModeDef is a struct that gives meaningful names to a theme mode (day or
// night)
type ModeDef struct {
	Primary        C
	Secondary      C
	PrimaryText    C
	SecondaryText  C
	DocText        C
	DocBg          C
	DocTextDim     C
	DocBgDim       C
	DocBgHighlight C
	PanelText      C
	PanelBg        C
	PanelTextDim   C
	PanelBgDim     C
	Fatal          C
	Warning        C
	Check          C
	Success        C
	Info           C
}

// ThemeDef is a day and night theme that should be designed to create a
// suitable dark/light theme mode.
type ThemeDef struct {
	Day, Night ModeDef
}

// DefineTheme accepts a ThemeDef and returns a Theme.
func DefineTheme(t ThemeDef) Theme {
	return Theme{
		Day: {
			Primary:        t.Day.Primary,
			Secondary:      t.Day.Secondary,
			PrimaryText:    t.Day.PrimaryText,
			SecondaryText:  t.Day.SecondaryText,
			DocText:        t.Day.DocText,
			DocBg:          t.Day.DocBg,
			DocTextDim:     t.Day.DocTextDim,
			DocBgDim:       t.Day.DocBgDim,
			DocBgHighlight: t.Day.DocBgHighlight,
			PanelText:      t.Day.PanelText,
			PanelBg:        t.Day.PanelBg,
			PanelTextDim:   t.Day.PanelTextDim,
			PanelBgDim:     t.Day.PanelBgDim,
			Fatal:          t.Day.Fatal,
			Warning:        t.Day.Warning,
			Check:          t.Day.Check,
			Success:        t.Day.Success,
			Info:           t.Day.Info,
		},
		Night: {
			Primary:        t.Night.Primary,
			Secondary:      t.Night.Secondary,
			PrimaryText:    t.Night.PrimaryText,
			SecondaryText:  t.Night.SecondaryText,
			DocText:        t.Night.DocText,
			DocBg:          t.Night.DocBg,
			DocTextDim:     t.Night.DocTextDim,
			DocBgDim:       t.Night.DocBgDim,
			DocBgHighlight: t.Night.DocBgHighlight,
			PanelText:      t.Night.PanelText,
			PanelBg:        t.Night.PanelBg,
			PanelTextDim:   t.Night.PanelTextDim,
			PanelBgDim:     t.Night.PanelBgDim,
			Fatal:          t.Night.Fatal,
			Warning:        t.Night.Warning,
			Check:          t.Night.Check,
			Success:        t.Night.Success,
			Info:           t.Night.Info,
		},
	}
}
