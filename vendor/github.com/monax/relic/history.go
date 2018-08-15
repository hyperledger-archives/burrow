package relic

import (
	"bytes"
	"fmt"
	"text/template"
)

// Provides the read-only methods of History to ensure releases are not accidentally mutated when it is not intended
type ImmutableHistory interface {
	// Get latest version
	CurrentVersion() Version
	// Get latest release note
	CurrentNotes() string
	// Get the project name
	Project() string
	// Get complete changelog
	Changelog() (string, error)
	// Get complete changelog or panic if error
	MustChangelog() string
	// Find the release specified by Version struct or version string
	Release(versionLike interface{}) (Release, error)
}

// The purpose of History is to capture version changes and change logs
// in a single location and use that data to generate releases and print
// changes to the command line to automate releases and provide a single
// source of truth for improved certainty around versions and releases
type History struct {
	ProjectName       string
	ProjectURL        string
	Releases          []Release
	ChangelogTemplate *template.Template
}

var _ ImmutableHistory = &History{}

type Release struct {
	Version Version
	Notes   string
}

type ReleasePair struct {
	Release
	Previous Release
}

var DefaultChangelogTemplate = template.Must(template.New("default_changelog_template").
	Parse(`# [{{ .Project }}]({{ .ProjectURL }}) Changelog{{ range .Releases }}
## [{{ .Version }}]{{ if .Version.Dated }} - {{ .Version.FormatDate }}{{ end }}
{{ .Notes }}
{{ end }}
{{ range .ReleasePairs }}[{{ .Version }}]: {{ $.URL }}/compare/{{ .Previous.Version.Ref }}...{{ .Version.Ref }}
{{ end }}[{{ .FirstRelease.Version }}]: {{ $.URL }}/commits/{{ .FirstRelease.Version.Ref }}`))

// Define a new project history to which releases can be added in code
// e.g. var history = relic.NewHistory().MustDeclareReleases(...)
func NewHistory(projectName, projectURL string) *History {
	return &History{
		ProjectName:       projectName,
		ProjectURL:        projectURL,
		ChangelogTemplate: DefaultChangelogTemplate,
	}
}

// Change the default changelog template from DefaultChangelogTemplate
func (h *History) WithChangelogTemplate(tmpl *template.Template) *History {
	h.ChangelogTemplate = tmpl
	return h
}

// Adds releases to the History with the newest releases provided first (so latest is at top).
// Releases can be specified by pairs of version (string or struct), notes (string) or by sequence of Release (struct)
// or mixtures thereof.
func (h *History) DeclareReleases(releaseLikes ...interface{}) (*History, error) {
	var rs []Release
	var err error
	for len(releaseLikes) > 0 {
		r, ok := releaseLikes[0].(Release)
		if ok {
			releaseLikes = releaseLikes[1:]
		} else {
			r, releaseLikes, err = readRelease(releaseLikes)
			if err != nil {
				return nil, err
			}
		}
		rs = append(rs, r)
	}
	// Check we still have a valid sequence of releases
	rs = append(rs, h.Releases...)
	err = ValidateReleases(rs)
	if err != nil {
		return h, err
	}

	h.Releases = rs
	return h, err
}

func readRelease(releaseLikes []interface{}) (rel Release, tail []interface{}, err error) {
	const fields = 2
	if len(releaseLikes) < fields {
		return rel, releaseLikes, fmt.Errorf("readRelease expects exactly 3 elements of version, date, notes")
	}
	rel.Version, err = AsVersion(releaseLikes[0])
	if err != nil {
		return rel, releaseLikes, err
	}
	rel.Notes, err = AsString(releaseLikes[1])
	if err != nil {
		return rel, releaseLikes, err
	}
	return rel, releaseLikes[fields:], nil
}

// Like DeclareReleases but will panic if the Releases list becomes invalid
func (h *History) MustDeclareReleases(releaseLikes ...interface{}) *History {
	h, err := h.DeclareReleases(releaseLikes...)
	if err != nil {
		panic(fmt.Errorf("could not register releases: %v", err))
	}
	return h
}

func (h *History) CurrentRelease() Release {
	for _, r := range h.Releases {
		if r.Version != ZeroVersion {
			return r
		}
	}
	return Release{}
}

func (h *History) FirstRelease() Release {
	l := len(h.Releases)
	if l == 0 {
		return Release{}
	}
	return h.Releases[l-1]
}

func (h *History) CurrentVersion() Version {
	return h.CurrentRelease().Version
}

// Gets the release notes for the current version
func (h *History) CurrentNotes() string {
	return h.CurrentRelease().Notes
}

func (h *History) Project() string {
	return h.ProjectName
}

func (h *History) URL() string {
	return h.ProjectURL
}

func (h *History) ReleasePairs() []ReleasePair {
	pairs := make([]ReleasePair, len(h.Releases)-1)
	for i := 0; i < len(pairs); i++ {
		pairs[i] = ReleasePair{
			Release:  h.Releases[i],
			Previous: h.Releases[i+1],
		}
	}
	return pairs
}

// Gets the Release for version given in versionString
func (h *History) Release(versionLike interface{}) (Release, error) {
	version, err := AsVersion(versionLike)
	if err != nil {
		return Release{}, err
	}
	for _, r := range h.Releases {
		if r.Version == version {
			return r, nil
		}
	}
	return Release{}, fmt.Errorf("could not find release with version %v", version)
}

// Get the changelog for the complete history
func (h *History) Changelog() (string, error) {
	buf := new(bytes.Buffer)
	err := h.ChangelogTemplate.Execute(buf, h)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Generates changelog, panicking if there is an error
func (h *History) MustChangelog() string {
	changelog, err := h.Changelog()
	if err != nil {
		panic(err)
	}
	return changelog
}

// Checks that a sequence of releases are monotonically decreasing with each
// version being a simple major, minor, or patch bump of its successor in the
// slice
func ValidateReleases(rs []Release) error {
	if len(rs) == 0 {
		return fmt.Errorf("at least one release must be defined")
	}
	// Allow the first version to be zero indicating unreleased as an place to document features
	if rs[0].Version == ZeroVersion {
		rs = rs[1:]
	}
	return EnsureReleasesUniqueValidAndMonotonic(rs)
}

func EnsureReleasesUniqueValidAndMonotonic(rs []Release) error {
	if len(rs) == 0 {
		return nil
	}
	version := rs[0].Version
	if version == ZeroVersion {
		return fmt.Errorf("only the top release may have an empty version to indicate unreleased, all" +
			"additional (earlier) releases must have a non-zero version")
	}
	for i := 1; i < len(rs); i++ {
		// The numbers of the lower version (expect descending sort)
		previousVersion := rs[i].Version
		if previousVersion == ZeroVersion {
			return fmt.Errorf("%v has version 0.0.0 but versions must start from 0.0.1", previousVersion)
		}
		// Check versions are consecutive
		if version.Major == previousVersion.Major+1 {
			// Major bump, so minor and patch versions must be reset
			if version.Minor != 0 || version.Patch != 0 {
				return fmt.Errorf("minor and patch versions must be reset to "+
					"0 after a major bump, but they are not in %s -> %s",
					rs[i].Version, rs[i-1].Version)
			}
		} else if version.Major == previousVersion.Major {
			// Same major number
			if version.Minor == previousVersion.Minor+1 {
				// Minor bump so patch version must be reset
				if version.Patch != 0 {
					return fmt.Errorf("patch version must be reset to "+
						"0 after a minor bump, but they are not in %s -> %s",
						rs[i].Version, rs[i-1].Version)
				}
			} else if version.Minor == previousVersion.Minor {
				// Same minor number so must be patch bump to be valid
				if version.Patch != previousVersion.Patch+1 {
					return fmt.Errorf("consecutive patch versions must be equal "+
						"or incremented by 1, but they are not in %s -> %s",
						rs[i].Version, rs[i-1].Version)
				}
			} else {
				return fmt.Errorf("consecutive minor versions must be equal or "+
					"incremented by 1, but they are not in %s -> %s",
					rs[i].Version, rs[i-1].Version)
			}
		} else {
			return fmt.Errorf("consecutive major versions must be equal or "+
				"incremented by 1, but they are not in  %s -> %s",
				rs[i].Version, rs[i-1].Version)
		}

		version = previousVersion
	}
	return nil
}
