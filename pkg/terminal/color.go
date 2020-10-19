package terminal

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/mgutz/ansi"
)

// A ColorFunc colors a string.
type ColorFunc func(string) string

// Additional colors codes.
var (
	CodeGray256 = "\x1b[38;5;242m"
)

// Available color functions
var (
	Magenta ColorFunc = ansi.ColorFunc("magenta")
	Cyan    ColorFunc = ansi.ColorFunc("cyan")
	Red     ColorFunc = ansi.ColorFunc("red")
	Yellow  ColorFunc = ansi.ColorFunc("yellow")
	Blue    ColorFunc = ansi.ColorFunc("blue")
	Green   ColorFunc = ansi.ColorFunc("green")
	Gray    ColorFunc = ansi.ColorFunc("black+h")
	Gray256 ColorFunc = func(s string) string {
		buf := bytes.NewBufferString(CodeGray256)
		buf.WriteString(s)
		buf.WriteString(ansi.Reset)
		return buf.String()
	}
	Bold ColorFunc = ansi.ColorFunc("default+b")
)

// ColorScheme - when enabled - returns colored string sequences. If not, they
// are returned unchanged.
type ColorScheme struct {
	enabled         bool
	color256enabled bool
}

// NewColorScheme creates a new color scheme. If not enabled, it will return the
// string sequences unchanged.
func NewColorScheme(enabled, color256enabled bool) *ColorScheme {
	return &ColorScheme{
		enabled:         enabled,
		color256enabled: color256enabled,
	}
}

// ColorCode returns the ANSI color code for a named color style.
func (c *ColorScheme) ColorCode(style string) string {
	if !c.enabled {
		return ""
	}
	return ansi.ColorCode(style)
}

// Magenta styles the text in magenta color, if the color scheme is enabled.
func (c *ColorScheme) Magenta(s string) string {
	if !c.enabled || s == "" {
		return s
	}
	return Magenta(s)
}

// Cyan styles the text in cyan color, if the color scheme is enabled.
func (c *ColorScheme) Cyan(s string) string {
	if !c.enabled || s == "" {
		return s
	}
	return Cyan(s)
}

// Red styles the text in red color, if the color scheme is enabled.
func (c *ColorScheme) Red(s string) string {
	if !c.enabled || s == "" {
		return s
	}
	return Red(s)
}

// Yellow styles the text in yellow color, if the color scheme is enabled.
func (c *ColorScheme) Yellow(s string) string {
	if !c.enabled || s == "" {
		return s
	}
	return Yellow(s)
}

// Blue styles the text in blue color, if the color scheme is enabled.
func (c *ColorScheme) Blue(s string) string {
	if !c.enabled || s == "" {
		return s
	}
	return Blue(s)
}

// Green styles the text in green color, if the color scheme is enabled.
func (c *ColorScheme) Green(s string) string {
	if !c.enabled || s == "" {
		return s
	}
	return Green(s)
}

// Gray styles the text in gray color, if the color scheme is enabled.
func (c *ColorScheme) Gray(s string) string {
	if !c.enabled || s == "" {
		return s
	} else if c.color256enabled {
		return Gray256(s)
	}
	return Gray(s)
}

// Bold styles the text bold, if the color scheme is enabled.
func (c *ColorScheme) Bold(s string) string {
	if !c.enabled || s == "" {
		return s
	}
	return Bold(s)
}

// SuccessIcon returns the success icon.
func (c *ColorScheme) SuccessIcon() string {
	return c.Green("✓")
}

// WarningIcon returns the warning icon.
func (c *ColorScheme) WarningIcon() string {
	return c.Yellow("!")
}

// ErrorIcon returns the error icon.
func (c *ColorScheme) ErrorIcon() string {
	return c.Red("✖")
}

// TemplateFuncs maps the color schemes functions to a map of template
// functions.
func (c *ColorScheme) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"bold":        c.Bold,
		"magenta":     c.Magenta,
		"cyan":        c.Cyan,
		"red":         c.Red,
		"yellow":      c.Yellow,
		"blue":        c.Blue,
		"green":       c.Green,
		"gray":        c.Gray,
		"successIcon": c.SuccessIcon,
		"warningIcon": c.WarningIcon,
		"errorIcon":   c.ErrorIcon,
	}
}

func envColorDisabled() bool {
	return os.Getenv("NO_COLOR") != "" || os.Getenv("CLICOLOR") == "0"
}

func envColorForced() bool {
	return os.Getenv("CLICOLOR_FORCE") != "" && os.Getenv("CLICOLOR_FORCE") != "0"
}

// https://github.com/cli/cli/blob/trunk/pkg/iostreams/color.go#L34
func color256Supported() bool {
	term := os.Getenv("TERM")
	colorterm := os.Getenv("COLORTERM")

	return strings.Contains(term, "256") ||
		strings.Contains(term, "24bit") ||
		strings.Contains(term, "truecolor") ||
		strings.Contains(colorterm, "256") ||
		strings.Contains(colorterm, "24bit") ||
		strings.Contains(colorterm, "truecolor")
}
