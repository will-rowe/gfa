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
	gfa    *GFA
}

// NewReader returns a new Reader, reading from the given io.Reader
func NewReader(r io.Reader) (*Reader, error) {
	gfaReader := &Reader{
		reader: bufio.NewReader(r),
		gfa:    NewGFA(),
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
				err := gfaReader.gfa.AddVersion(1)
				if err != nil {
					return nil, err
				}
			} else if bytes.Contains(line, []byte("VN:Z:2")) {
				err := gfaReader.gfa.AddVersion(2)
				if err != nil {
					return nil, err
				}
			}
			if line[0] == '#' {
				gfaReader.gfa.AddComment(line[1 : len(line)-1])
			}
		} else {
			break
		}
	}
	return gfaReader, nil
}

// CollectGFA returns the GFA instance held by the reader
func (r *Reader) CollectGFA() *GFA {
	return r.gfa
}

// Read returns the next (non H/#) GFA line from the reader
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
	// split the line on tab
	fields := bytes.Split(bytesLine, []byte("\t"))
	if len(fields) < 3 {
		return nil, fmt.Errorf("Not enough fields in GFA line: %v", string(bytesLine))
	}
	// determine what type of line it is and then create a gfaLine using the required fields
	switch bytesLine[0] {
	// segment line (S)
	case 83:
		line, err = NewSegment(fields[1], fields[2])
		if err != nil {
			return nil, fmt.Errorf("Could not read segment line: %v", err)
		}
		fields = fields[3:]
	// link line (L)
	case 76:
		line, err = NewLink(fields[1], fields[2], fields[3], fields[4], fields[5])
		if err != nil {
			return nil, fmt.Errorf("Could not read link line: %v", err)
		}
		fields = fields[6:]
	// containment line (C)
	case 67:
		line, err = NewSegment([]byte("dummy4containment"), []byte("actg"))
		if err != nil {
			return nil, fmt.Errorf("Could not read containment line: %v", err)
		}
		fields = fields[3:]
	// path line (P)
	case 80:
		line, err = NewPath(fields[1], bytes.Split(fields[2], []byte(",")), bytes.Split(fields[3], []byte(",")))
		if err != nil {
			return nil, fmt.Errorf("Could not read path line: %v", err)
		}
		fields = fields[4:]
	default:
		return nil, fmt.Errorf("Encountered unknown line type: %v", string(bytesLine[0]))
	}
	// fields slice will now only contain optional fields (if present)
	if len(fields) != 0 {
		oFs, err := NewOptionalFields(fields...)
		if err != nil {
			return nil, fmt.Errorf("Could not parse optional fields: %v", err)
		}
		line.AddOptionalFields(oFs)
	}

	return line, nil
}
