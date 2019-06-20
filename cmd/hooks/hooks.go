package hooks

import (
	"bufio"
	"io"
)

// ReadRefs reads a list of newline delimeted refs from a reader
func ReadRefs(r io.Reader) ([]string, error) {
	s := bufio.NewScanner(r)

	var refs []string
	for s.Scan() {
		refs = append(refs, s.Text())
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return refs, nil
}
