package conflict

// This file is a direct port of Ruby source previously hosted at
// ruby/lib/gitlab/git/conflict/parser.rb (git show fb5717dd5567082)
// ruby/lib/gitlab/git/conflict/file.rb (git show 6f787458251b5f)

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
)

// Errors that can occur during parsing of a merge conflict file
var (
	ErrUnmergeableFile     = errors.New("merging is not supported for file")
	ErrUnexpectedDelimiter = errors.New("unexpected conflict delimiter")
	ErrMissingEndDelimiter = errors.New("missing last delimiter")
)

type section uint

const (
	sectionNone = section(iota)
	sectionOld
	sectionNew
	sectionNoNewline
)

const fileLimit = 200 * (1 << 10) // 200k

type line struct {
	objIndex uint // where are we in the object?
	oldIndex uint // where are we in the old file?
	newIndex uint // where are we in the new file?

	payload string // actual line contents (minus the newline)
	section section
}

// File contains an ordered list of lines with metadata about potential
// conflicts.
type File struct {
	path  string
	lines []line
}

func (f File) sectionID(l line) string {
	pathSHA1 := sha1.Sum([]byte(f.path))
	return fmt.Sprintf("%x_%d_%d", pathSHA1, l.oldIndex, l.newIndex)
}

// Resolution indicates how to resolve a conflict
type Resolution struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`

	// key is a sectionID, value is "head" or "origin"
	Sections map[string]string `json:"sections"`
}

const (
	head   = "head"
	origin = "origin"
)

// Resolve will iterate through each conflict line and replace it with the
// specified resolution
func (f File) Resolve(resolution Resolution) ([]byte, error) {
	var sectionID string
	b := bytes.NewBuffer(nil)

	for _, l := range f.lines {
		if l.section == sectionNone {
			sectionID = ""
			if _, err := b.WriteString(l.payload + "\n"); err != nil {
				return nil, err
			}
			continue
		}

		if sectionID == "" {
			sectionID = f.sectionID(l)
		}

		r, ok := resolution.Sections[sectionID]
		if !ok {
			return nil, fmt.Errorf("missing resolution for section ID: %s", sectionID)
		}

		switch r {
		case head:
			if l.section != sectionNew {
				continue
			}
		case origin:
			if l.section != sectionOld {
				continue
			}
		default:
			return nil, fmt.Errorf("unknown resolution: %q", r)
		}

		if _, err := b.WriteString(l.payload); err != nil {
			return nil, err
		}

		if l.section == sectionNoNewline {
			continue
		}
		if _, err := b.WriteString("\n"); err != nil {
			return nil, err
		}
	}

	return b.Bytes(), nil
}

// Parse will read each line and maintain which conflict section it belongs to
func Parse(src io.Reader, ourPath, theirPath, parentPath string) (File, error) {
	var (
		// conflict markers
		start     = "<<<<<<< " + ourPath
		middle    = "======="
		end       = ">>>>>>> " + theirPath
		noNewline = `\`

		f                                 = File{path: parentPath}
		objIndex, oldIndex, newIndex uint = 0, 1, 1
		currentSection               section
		bytesRead                    int

		s = bufio.NewScanner(src)
	)

	s.Buffer(make([]byte, 4096), fileLimit) // allow for line scanning up to the file limit

	s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		bytesRead += len(data)
		if bytesRead >= fileLimit {
			return 0, nil, ErrUnmergeableFile
		}
		return bufio.ScanLines(data, atEOF)
	})

	for s.Scan() {
		switch l := s.Text(); l {
		case start:
			if currentSection != sectionNone {
				return File{}, ErrUnexpectedDelimiter
			}
			currentSection = sectionNew
		case middle:
			if currentSection != sectionNew {
				return File{}, ErrUnexpectedDelimiter
			}
			currentSection = sectionOld
		case end:
			if currentSection != sectionOld {
				return File{}, ErrUnexpectedDelimiter
			}
			currentSection = sectionNone
		case noNewline:
			currentSection = sectionNoNewline
			fallthrough
		default:
			f.lines = append(f.lines, line{
				objIndex: objIndex,
				oldIndex: oldIndex,
				newIndex: newIndex,
				payload:  l,
				section:  currentSection,
			})

			objIndex++
			if currentSection != sectionNew {
				oldIndex++
			}
			if currentSection != sectionOld {
				newIndex++
			}
		}
	}

	if err := s.Err(); err != nil {
		return File{}, err
	}

	if currentSection == sectionOld || currentSection == sectionNew {
		return File{}, ErrMissingEndDelimiter
	}

	if bytesRead == 0 {
		return File{}, ErrUnmergeableFile // typically a binary file
	}

	return f, nil
}