package terminal

import (
	"os"
	"text/template"

	"github.com/muesli/termenv"
)

// A ColorFunc colors a string.
type ColorFunc func(string) string

// ColorPair specifies a color with different accents for light and dark
// background.
type ColorPair struct {
	light string
	dark  string
}

// Color returns the color code for the background.
func (cp *ColorPair) Color(dark bool) string {
	if dark {
		return cp.dark
	}
	return cp.light
}

// Available colors
var (
	Black   = ColorPair{"0", "8"}
	Red     = ColorPair{"1", "9"}
	Green   = ColorPair{"2", "10"}
	Yellow  = ColorPair{"3", "11"}
	Blue    = ColorPair{"4", "12"}
	Magenta = ColorPair{"5", "13"}
	Cyan    = ColorPair{"6", "14"}
	White   = ColorPair{"7", "15"}
	Gray    = ColorPair{"242", "248"}
)

// ColorScheme - when enabled - returns colored string sequences. If not, they
// are returned unchanged.
type ColorScheme struct {
	colorProfile termenv.Profile

	enabled bool
	dark    bool
}

// NewColorScheme creates a new color scheme. If not enabled, it will return the
// string sequences unchanged.
func NewColorScheme(enabled bool) *ColorScheme {
	return &ColorScheme{
		colorProfile: termenv.ColorProfile(),

		enabled: enabled,
		dark:    termenv.HasDarkBackground(),
	}
}

// Black styles the text in black color, if the color scheme is enabled.
func (cs *ColorScheme) Black(s string) string {
	return cs.colorize(s, Black)
}

// Red styles the text in red color, if the color scheme is enabled.
func (cs *ColorScheme) Red(s string) string {
	return cs.colorize(s, Red)
}

// Green styles the text in green color, if the color scheme is enabled.
func (cs *ColorScheme) Green(s string) string {
	return cs.colorize(s, Green)
}

// Yellow styles the text in yellow color, if the color scheme is enabled.
func (cs *ColorScheme) Yellow(s string) string {
	return cs.colorize(s, Yellow)
}

// Blue styles the text in blue color, if the color scheme is enabled.
func (cs *ColorScheme) Blue(s string) string {
	return cs.colorize(s, Blue)
}

// Magenta styles the text in magenta color, if the color scheme is enabled.
func (cs *ColorScheme) Magenta(s string) string {
	return cs.colorize(s, Magenta)
}

// Cyan styles the text in cyan color, if the color scheme is enabled.
func (cs *ColorScheme) Cyan(s string) string {
	return cs.colorize(s, Cyan)
}

// White styles the text in white color, if the color scheme is enabled.
func (cs *ColorScheme) White(s string) string {
	return cs.colorize(s, White)
}

// Gray styles the text in gray color, if the color scheme is enabled.
func (cs *ColorScheme) Gray(s string) string {
	return cs.colorize(s, Gray)
}

// Bold styles the text bold, if the color scheme is enabled.
func (cs *ColorScheme) Bold(s string) string {
	if !cs.enabled || s == "" {
		return s
	}
	return termenv.String(s).Bold().String()
}

// Title returns styles the string as a title, if the color scheme is enabled.
func (cs *ColorScheme) Title(s string) string {
	return termenv.String(s).Foreground(cs.color(White)).Background(cs.color(Magenta)).String()
}

// SuccessIcon returns the success icon.
func (cs *ColorScheme) SuccessIcon() string {
	return cs.Green("✓")
}

// WarningIcon returns the warning icon.
func (cs *ColorScheme) WarningIcon() string {
	return cs.Yellow("!")
}

// ErrorIcon returns the error icon.
func (cs *ColorScheme) ErrorIcon() string {
	return cs.Red("✖")
}

// IsDark returns true if the terminals background is dark.
func (cs *ColorScheme) IsDark() bool {
	return cs.dark
}

// Code returns the escape sequence for the given foreground color.
func (cs *ColorScheme) Code(s string) string {
	return cs.colorProfile.Color(s).Sequence(false)
}

// TemplateFuncs maps the color schemes functions to a map of template
// functions.
func (cs *ColorScheme) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"black":       cs.Black,
		"red":         cs.Red,
		"green":       cs.Green,
		"yellow":      cs.Yellow,
		"blue":        cs.Blue,
		"magenta":     cs.Magenta,
		"cyan":        cs.Cyan,
		"white":       cs.White,
		"gray":        cs.Gray,
		"bold":        cs.Bold,
		"title":       cs.Title,
		"successIcon": cs.SuccessIcon,
		"warningIcon": cs.WarningIcon,
		"errorIcon":   cs.ErrorIcon,
	}
}

func (cs *ColorScheme) colorize(s string, cp ColorPair) string {
	if !cs.enabled || s == "" {
		return s
	}
	c := cs.color(cp)
	return termenv.String(s).Foreground(c).String()
}

func (cs *ColorScheme) color(cp ColorPair) termenv.Color {
	return cs.colorProfile.Color(cp.Color(cs.dark))
}

func envColorDisabled() bool {
	return os.Getenv("NO_COLOR") != "" || os.Getenv("CLICOLOR") == "0"
}

func envColorForced() bool {
	return os.Getenv("CLICOLOR_FORCE") != "" && os.Getenv("CLICOLOR_FORCE") != "0"
}
