// Package gfa is a Go library for working with the Graphical Fragment Assembly (GFA) format.
/*

For more information:
    GFA Format Specification
    https://github.com/GFA-spec/GFA-spec

This package currently only conforms to GFA1 spec.
*/
package gfa

import (
	"bytes"
	"fmt"
)

// The GFA type holds all the information from a GFA formatted file
type GFA struct {
	header       *header
	comments     [][]byte
	segments     []*segment
	links        []*link
	containments []containment       // TODO: not yet implemented
	paths        []path              // TODO: not yet implemented
	segRecord    map[string]struct{} // prevent duplicate segment IDs being added
}

// NewGFA returns a new GFA instance
func NewGFA() *GFA {
	return &GFA{
		header:    &header{entryType: "H"},
		segRecord: make(map[string]struct{}),
	}
}

// AddVersion adds the GFA format version to the GFA instance
func (gfa *GFA) AddVersion(v int) error {
	switch v {
	case 0:
		return fmt.Errorf("GFA instance already has a version number attached")
	case 1:
		gfa.header.vn = v
	case 2:
		return fmt.Errorf("GFA version 2 is currently unsupported...")
	default:
		return fmt.Errorf("GFA format must be either version 1 or version 2")
	}
	return nil
}

// AddComment appends a comment to the comments held by the GFA instance
func (gfa *GFA) AddComment(c []byte) {
	comment := append([]byte("#\t"), c...)
	gfa.comments = append(gfa.comments, comment)
}

/*
GetVersion returns the GFA version

// a return value of 0 indicates no version set
*/
func (gfa *GFA) GetVersion() int {
	return gfa.header.vn
}

// GetSegments returns a slice of all the segments held in the GFA instance
func (gfa *GFA) GetSegments() ([]*segment, error) {
	if len(gfa.segments) == 0 {
		return nil, fmt.Errorf("no segments currently held in GFA instance")
	}
	return gfa.segments, nil
}

// GetLinks returns a slice of all the links held in the GFA instance
func (gfa *GFA) GetLinks() ([]*link, error) {
	if len(gfa.links) == 0 {
		return nil, fmt.Errorf("no links currently held in GFA instance")
	}
	return gfa.links, nil
}

// PrintHeader prints the GFA formatted header line
func (gfa *GFA) PrintHeader() string {
	return fmt.Sprintf("%v\tVN:Z:%v", gfa.header.entryType, gfa.header.vn)
}

// PrintComments prints a string of GFA formatted comment line(s)
func (gfa *GFA) PrintComments() string {
	return fmt.Sprintf("%s", bytes.Join(gfa.comments, []byte("\n")))
}

// A header contains a type field (required) and a GFA version number field (optional)
type header struct {
	entryType string
	vn        int
}

// An interface for the non-comment/header GFA lines
type gfaLine interface {
	PrintGFAline() string
	Add(*GFA) error
}

// A segment contains a type field, name and sequence (all required), plus optional fields (length, ...)
type segment struct {
	entryType string
	name      []byte
	sequence  []byte // this is technically not required by the spec but I have set it as required here
	length    int
	readCount int
	fragCount int
	kmerCount int
	checksum  []byte
	uri       string
}

/*
NewSegment is a segment constructor

// segment name (n) and sequence (seq) are requred fields

// optional fields can be supplied
*/
func NewSegment(n, seq []byte, optional ...[]byte) (*segment, error) {
	if bytes.ContainsAny(n, "+-*= ") {
		return nil, fmt.Errorf("Segment name can't contain +/-/*/= or whitespace")
	}
	if len(seq) == 0 {
		return nil, fmt.Errorf("Segment must have a sequence")
	}
	seg := new(segment)
	seg.entryType = "S"
	seg.name = n
	seg.sequence = seq
	seg.length = len(seq)
	if len(optional) != 0 {
		for _, field := range optional {
			val := bytes.Split(field, []byte(":"))
			switch string(val[0]) {
			case "RC":
				seg.readCount = int(val[2][0])
			case "FC":
				seg.fragCount = int(val[2][0])
			case "KC":
				seg.kmerCount = int(val[2][0])
			case "SH":
				seg.checksum = val[2]
			case "UR":
				seg.uri = string(val[2])
			case "LN":
				continue
			default:
				return nil, fmt.Errorf("Don't recognise optional field: %v", string(field))
			}
		}
	}
	return seg, nil
}

// PrintGFAline prints a GFA formatted segment line
func (seg *segment) PrintGFAline() string {
	line := fmt.Sprintf("%v\t%v\t%v\tLN:i:%v", seg.entryType, string(seg.name), string(seg.sequence), seg.length)
	if seg.readCount != 0 {
		line = fmt.Sprintf("%v\tRC:i:%v", line, seg.readCount)
	}
	if seg.fragCount != 0 {
		line = fmt.Sprintf("%v\tFC:i:%v", line, seg.fragCount)
	}
	if seg.kmerCount != 0 {
		line = fmt.Sprintf("%v\tKC:i:%v", line, seg.kmerCount)
	}
	if seg.checksum != nil {
		line = fmt.Sprintf("%v\tSH:i:%s", line, seg.checksum)
	}
	if seg.uri != "" {
		line = fmt.Sprintf("%v\tUR:i:%v", line, seg.uri)
	}
	return line
}

// Add appends a segment to a specified GFA instance
func (seg *segment) Add(gfa *GFA) error {
	if _, ok := gfa.segRecord[string(seg.name)]; ok {
		return fmt.Errorf("Duplicate segment name already present in GFA instance: %v", string(seg.name))
	}
	gfa.segments = append(gfa.segments, seg)
	gfa.segRecord[string(seg.name)] = struct{}{}
	return nil
}

/*
Links are the primary mechanism to connect segments. Links connect oriented segments.
A link from A to B means that the end of A overlaps with the start of B.
If either is marked with -, we replace the sequence of the segment with its reverse complement, whereas a + indicates the segment sequence is used as-is.
The length of the overlap is determined by the CIGAR string of the link.
When the overlap is 0M the B segment follows directly after A.
When the CIGAR string is *, the nature of the overlap is not specified.
*/
type link struct {
	entryType  string
	from       []byte
	fromOrient string
	to         []byte
	toOrient   string
	overlap    string
}

/*
NewLink is a link constructor

// from, fOrient, to, tOrient and overlap are requred fields

// optional fields can be supplied
*/
func NewLink(from, fOrient, to, tOrient, overlap []byte, optional ...[]byte) (*link, error) {
	if bytes.ContainsAny(from, "+-*= ") {
		return nil, fmt.Errorf("Segment name can't contain +/-/*/= or whitespace")
	}
	if bytes.ContainsAny(to, "+-*= ") {
		return nil, fmt.Errorf("Segment name can't contain +/-/*/= or whitespace")
	}
	link := new(link)
	link.from = from
	link.to = to
	link.entryType = "L"
	fori, tori := string(fOrient), string(tOrient)
	if (fori == "+") || (fori == "-") {
		link.fromOrient = fori
	} else {
		return nil, fmt.Errorf("From orientation field must be either + or -")
	}
	if (tori == "+") || (tori == "-") {
		link.toOrient = tori
	} else {
		return nil, fmt.Errorf("To orientation field must be either + or -")
	}
	link.overlap = string(overlap)
	// TODO: add optional fields...

	return link, nil
}

// PrintGFAline prints a GFA formatted link line
func (self *link) PrintGFAline() string {
	line := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v", self.entryType, string(self.from), self.fromOrient, string(self.to), self.toOrient, self.overlap)
	return line
}

// Add appends a link to a specified GFA instance
func (link *link) Add(gfa *GFA) error {
	gfa.links = append(gfa.links, link)
	return nil
}

// containment
type containment struct {
	entryType string
}

// path
type path struct {
	entryType string
}
