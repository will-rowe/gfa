package gfa

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// Reader implements GFA format reading.
type Reader struct {
	reader *bufio.Reader
	meta   *header
}

// NewReader returns a new Reader, reading from the given io.Reader.
func NewReader(r io.Reader) (*Reader, error) {
	h := NewHeader()
	gfaReader := &Reader{
		reader: bufio.NewReader(r),
		meta:   h,
	}
	// check there is something in the file
	_, err := gfaReader.reader.Peek(1)
	if err != nil {
		return nil, err
	}
	// get the header lines, ignore comments and stop looking once a non header/comment line encountered
	for {
		peek, err := gfaReader.reader.Peek(1)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// only look at lines beginning with H (72) and # (35)
		if (peek[0] == 72) || (peek[0] == 35) {
			line, err := gfaReader.reader.ReadBytes('\n')
			if err != nil {
				return nil, io.ErrUnexpectedEOF
			}
			if bytes.Contains(line, []byte("VN:Z:1")) {
				err := gfaReader.meta.AddVersionNumber(1)
				if err != nil {
					return nil, err
				}
			} else if bytes.Contains(line, []byte("VN:Z:2")) {
				err := gfaReader.meta.AddVersionNumber(2)
				if err != nil {
					return nil, err
				}
			}
			if line[0] == '#' {
				gfaReader.meta.AddComment(line[1 : len(line)-1])
			}
		} else {
			break
		}
	}
	return gfaReader, nil
}

// Returns the GFA metadata held by the reader
func (r *Reader) CollectMeta() *header {
	return r.meta
}

// Read returns the next (non H/#) line in the GFA file
func (r *Reader) Read() (gfaLine, error) {
	bytesLine, err := r.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	// trim the line
	bytesLine = bytesLine[:len(bytesLine)-1]
	if bytesLine[len(bytesLine)-1] == '\r' {
		bytesLine = bytesLine[:len(bytesLine)-1]
	}
	var line gfaLine
	// grab segment lines (S)
	if bytesLine[0] == 83 {
		fields, err := parseLine(bytesLine)
		if err != nil {
			return nil, fmt.Errorf("Could parse line: %v", err)
		}
		if len(fields) > 3 {
			line, err = NewSegment(fields[1], fields[2], fields[3:]...)
			if err != nil {
				return nil, fmt.Errorf("Could not read segment line: %v", err)
			}
		} else {
			line, err = NewSegment(fields[1], fields[2])
			if err != nil {
				return nil, fmt.Errorf("Could not read segment line: %v", err)
			}
		}
	}
	// grab link lines (L)
	if bytesLine[0] == 76 {
		fields, err := parseLine(bytesLine)
		if err != nil {
			return nil, fmt.Errorf("Could parse line: %v", err)
		}
		if len(fields) > 6 {
			line, err = NewLink(fields[1], fields[2], fields[3], fields[4], fields[5], fields[6:]...)
			if err != nil {
				return nil, fmt.Errorf("Could not read link line: %v", err)
			}
		} else {
			line, err = NewLink(fields[1], fields[2], fields[3], fields[4], fields[5])
			if err != nil {
				return nil, fmt.Errorf("Could not read link line: %v", err)
			}
		}
	}

	// grab containment lines (C)
	if bytesLine[0] == 67 {
		line, err = NewSegment([]byte("dummy"), []byte("actg"))
		if err != nil {
			return nil, fmt.Errorf("Could not read link line: %v", err)
		}
	}

	// grab containment lines (P)
	if bytesLine[0] == 80 {
		line, err = NewSegment([]byte("dummy"), []byte("actg"))
		if err != nil {
			return nil, fmt.Errorf("Could not read link line: %v", err)
		}
	}

	// return error if the line type could not be identified
	if line == nil {
		return nil, fmt.Errorf("Encountered unknown line type: %v", string(bytesLine[0]))
	}
	return line, nil
}

// Parses the gfa (non header/comment) lines
func parseLine(line []byte) ([][]byte, error) {
	fields := bytes.Split(line, []byte("\t"))
	// need at least 3 fields
	if len(fields) < 3 {
		return nil, fmt.Errorf("Not enough fields in GFA line: %v", string(line))
	}
	// basic check of the sequence field
	if bytes.ContainsAny(fields[2], ": ") {
		return nil, fmt.Errorf("Doesn't look like sequence (%v): %v", string(fields[2]), string(line))
	}

	return fields, nil
}
