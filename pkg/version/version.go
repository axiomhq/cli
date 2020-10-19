// Package version provides variables which, when used correctly, provide
// build information about the application referencing this package. At compile
// time the appropriate linker flags must be passed:
//
// 	go build -ldflags "-X github.com/axiomhq/cli/pkg/version.release=1.0.0"
//
// Adapt the flags for all other exported variables. Eventually use vendored
// version.
package version

import (
	"fmt"
	"html/template"
	"runtime"
	"strings"
	"time"
)

var (
	release   = "-"
	revision  = "-"
	buildUser = "-"
	buildDate = "-"
	goVersion = runtime.Version()
)

// Release is the semantic version of the current build.
func Release() string {
	return release
}

// Revision is the last git commit hash of the source repository at the moment
// the binary was built.
func Revision() string {
	return revision
}

// BuildUser is the username of the user who performed the build.
func BuildUser() string {
	return buildUser
}

// BuildDate is a timestamp of the moment when the binary was built.
func BuildDate() (time.Time, error) {
	return time.ParseInLocation(time.RFC3339, buildDate, time.UTC)
}

// BuildDateString is BuildDate formatted as RFC3339.
func BuildDateString() string {
	ts, err := BuildDate()
	if err != nil {
		return buildDate
	}
	return ts.UTC().Format(time.RFC3339)
}

// GoVersion is the go version the build utilizes.
func GoVersion() string {
	return goVersion
}

// String returns the version and build information.
func String() string {
	return fmt.Sprintf("release=%s, revision=%s, user=%s, build=%s, go=%s",
		release, revision, buildUser, buildDate, goVersion)
}

// versionInfoTmpl contains the template used by Info.
var versionInfoTmpl = `
{{.program}}, release {{.release}} (revision: {{.revision}})
  build user:       {{.buildUser}}
  build date:       {{.buildDate}}
  go version:       {{.goVersion}}
`

// Print returns version and build information in a user friendly, formatted
// string.
func Print(program string) string {
	m := map[string]string{
		"program":   program,
		"release":   Release(),
		"revision":  Revision(),
		"buildUser": BuildUser(),
		"buildDate": BuildDateString(),
		"goVersion": GoVersion(),
	}
	t := template.Must(template.New("version").Parse(versionInfoTmpl))

	var b strings.Builder
	if err := t.ExecuteTemplate(&b, "version", m); err != nil {
		return ""
	}
	return strings.TrimSpace(b.String())
}
