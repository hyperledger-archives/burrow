package relic

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"
)

// The purpose of History is to capture version changes and change logs
// in a single location and use that data to generate releases and print
// changes to the command line to automate releases and provide a single
// source of truth for improved certainty around versions and releases
type History struct {
	ProjectName       string
	Releases          []Release
	ChangelogTemplate *template.Template
}

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

var _ ImmutableHistory = &History{}

type Release struct {
	Version Version
	Notes   string
}

var DefaultChangelogTemplate = template.Must(template.New("default_changelog_template").
	Parse(`# {{ .Project }} Changelog{{ range .Releases }}
## Version {{ .Version }}
{{ .Notes }}
{{ end }}`))

// Define a new project history to which releases can be added in code
// e.g. var history = relic.NewHistory().MustDeclareReleases(...)
func NewHistory(projectName string) *History {
	return &History{
		ProjectName:       projectName,
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
	for i := 0; i < len(releaseLikes); i++ {
		switch r := releaseLikes[i].(type) {
		case Release:
			rs = append(rs, r)
		case fmt.Stringer:
			releaseLikes[i] = r.String()
			i--
		case string:
			version, err := AsVersion(r)
			if err != nil {
				return nil, fmt.Errorf("could not interpret %v as a version: %v", r, err)
			}
			if i+1 >= len(releaseLikes) {
				return nil, fmt.Errorf("when specifying releases in pairs of version and note you must provide "+
					"both, but the last release (for version %s) has no note", version.String())
			}
			release := Release{
				Version: version,
			}
			// Get notes from next element
			switch n := releaseLikes[i+1].(type) {
			case string:
				release.Notes = n
			case fmt.Stringer:
				release.Notes = n.String()
			default:
				return nil, fmt.Errorf("release element %v should be notes but cannot be converted to string",
					releaseLikes[i+1])
			}
			if release.Notes == "" {
				return nil, fmt.Errorf("release note for version %s is empty", version.String())
			}
			rs = append(rs, release)
			// consume an additional element
			i++
		}
	}
	// Check we still have a valid sequence of releases
	rs = append(rs, h.Releases...)
	err := EnsureReleasesUniqueValidAndMonotonic(rs)
	if err != nil {
		return h, err
	}

	h.Releases = rs
	return h, err
}

// Like DeclareReleases but will panic if the Releases list becomes invalid
func (h *History) MustDeclareReleases(releaseLikes ...interface{}) *History {
	h, err := h.DeclareReleases(releaseLikes...)
	if err != nil {
		panic(fmt.Errorf("could not register releases: %v", err))
	}
	return h
}

func (h *History) CurrentVersion() Version {
	return h.Releases[0].Version
}

// Gets the release notes for the current version
func (h *History) CurrentNotes() string {
	return h.Releases[0].Notes
}

func (h *History) Project() string {
	return h.ProjectName
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
func EnsureReleasesUniqueValidAndMonotonic(rs []Release) error {
	if len(rs) == 0 {
		return errors.New("at least one release must be defined")
	}
	version := rs[0].Version
	for i := 1; i < len(rs); i++ {
		// The numbers of the lower version (expect descending sort)
		previousVersion := rs[i].Version
		// Check versions are consecutive
		if version.Major() == previousVersion.Major()+1 {
			// Major bump, so minor and patch versions must be reset
			if version.Minor() != 0 || version.Patch() != 0 {
				return fmt.Errorf("minor and patch versions must be reset to "+
					"0 after a major bump, but they are not in %s -> %s",
					rs[i].Version, rs[i-1].Version)
			}
		} else if version.Major() == previousVersion.Major() {
			// Same major number
			if version.Minor() == previousVersion.Minor()+1 {
				// Minor bump so patch version must be reset
				if version.Patch() != 0 {
					return fmt.Errorf("patch version must be reset to "+
						"0 after a minor bump, but they are not in %s -> %s",
						rs[i].Version, rs[i-1].Version)
				}
			} else if version.Minor() == previousVersion.Minor() {
				// Same minor number so must be patch bump to be valid
				if version.Patch() != previousVersion.Patch()+1 {
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
