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
	"strconv"
	"strings"
)

// The GFA type holds all the information from a GFA formatted file
type GFA struct {
	header       *header
	comments     [][]byte
	segments     []*segment
	links        []*link
	containments []containment // TODO: not yet implemented
	paths        []*path
	segRecord    map[string]struct{} // prevents duplicate segment IDs being added
}

// NewGFA returns a new GFA instance
func NewGFA() *GFA {
	return &GFA{
		header:    &header{recordType: "H"},
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
	return gfa.links, nil
}

// GetPaths returns a slice of all the paths held in the GFA instance
func (gfa *GFA) GetPaths() ([]*path, error) {
	if len(gfa.paths) == 0 {
		return nil, fmt.Errorf("no paths currently held in GFA instance")
	}
	return gfa.paths, nil
}

// PrintHeader prints the GFA formatted header line
func (gfa *GFA) PrintHeader() string {
	return fmt.Sprintf("%v\tVN:Z:%v", gfa.header.recordType, gfa.header.vn)
}

// PrintComments prints a string of GFA formatted comment line(s)
func (gfa *GFA) PrintComments() string {
	return fmt.Sprintf("%s", bytes.Join(gfa.comments, []byte("\n")))
}

/*
Validate performs several checks on the GFA instance TODO: add more checks

// checks that it contains a version (1/2)

// checks that is contains 1 or more segments
*/
func (gfa *GFA) Validate() error {
	if gfa.GetVersion() == 0 {
		return fmt.Errorf("Please set GFA to format version 1 or 2")
	}
	if gfa.GetVersion() > 2 {
		return fmt.Errorf("GFA version not recognised: %d", gfa.GetVersion())
	}
	if len(gfa.segments) == 0 {
		return fmt.Errorf("GFA instance contains no segments")
	}
	return nil
}

// MarshalHeader prepares the header/comment lines for a writer
func (gfa *GFA) MarshalHeader() []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%v\tVN:Z:%v\n", gfa.header.recordType, gfa.header.vn)
	if len(gfa.comments) != 0 {
		fmt.Fprintf(&buf, "%s", bytes.Join(gfa.comments, []byte("\n")))
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

// WriteGFAContent will dump the content of a GFA instance to file
func (gfa *GFA) WriteGFAContent(w *GFAwriter) error {
	if err := gfa.Validate(); err != nil {
		return fmt.Errorf("GFA validation failed, can't write GFA content: %v", err)
	}
	for _, seg := range gfa.segments {
		err := w.Write(seg)
		if err != nil {
			return fmt.Errorf("Can't write GFA content: %v", err)
		}
	}
	for _, link := range gfa.links {
		err := w.Write(link)
		if err != nil {
			return fmt.Errorf("Can't write GFA content: %v", err)
		}
	}
	for _, path := range gfa.paths {
		err := w.Write(path)
		if err != nil {
			return fmt.Errorf("Can't write GFA content: %v", err)
		}
	}
	return nil
}

// PrintSequence will return the sequence encoded by a specified pathName
func (gfa *GFA) PrintSequence(pathName []byte) ([]byte, error) {
	sequence := []byte{}
	if err := gfa.Validate(); err != nil {
		return nil, err
	}
	// get the specified path from the graph
	for _, path := range gfa.paths {
		if string(path.PathName) == string(pathName) {
			// read the graph segments into a map
			segMap := make(map[string][]byte)
			for _, seg := range gfa.segments {
				segMap[string(seg.Name)] = seg.Sequence
			}
			// build up the sequence using the path and segMap
			for _, name := range path.SegNames {
				// remove the plus from the seg name, lookup seg in map, append seq
				sequence = append(sequence, segMap[string(bytes.TrimSuffix(name, []byte("+")))]...)
			}
			break
		}
	}
	// if the specified pathName wasn't found, return error
	if len(sequence) == 0 {
		return nil, fmt.Errorf("specified pathName not found in GFA")
	}
	return sequence, nil
}

// A header contains a type field (required) and a GFA version number field (optional)
type header struct {
	recordType string
	vn         int
}

// An interface for the non-comment/header GFA lines
type gfaLine interface {
	AddOptionalFields(*optionalFields)
	PrintGFAline() string
	Add(*GFA) error
}

// A segment contains a type field, name and sequence (all required), plus optional fields (length, ...)
type segment struct {
	recordType string
	Name       []byte
	Sequence   []byte // this is technically not required by the spec but I have set it as required here
	Length     int    // this is technically an optional field but is added automatically when a sequence is supplied
	optional   *optionalFields
}

// NewSegment is a segment constructor
func NewSegment(n, seq []byte) (*segment, error) {
	if bytes.ContainsAny(n, "+-*= ") {
		return nil, fmt.Errorf("Segment name can't contain +/-/*/= or whitespace")
	}
	if len(seq) == 0 {
		return nil, fmt.Errorf("Segment must have a sequence")
	}
	return &segment{
		recordType: "S",
		Name:       n,
		Sequence:   seq,
		Length:     len(seq),
	}, nil
}

// AddOptionalFields adds a set of optional fields to a segment
func (seg *segment) AddOptionalFields(oFs *optionalFields) {
	seg.optional = oFs
}

// GetKmerCount returns the k-mer count of a segment
func (seg *segment) GetKmerCount() (int, error) {
	if seg.optional.kmerCount == "" {
		return 0, nil
	}
	return strconv.Atoi(seg.optional.kmerCount)
}

// PrintGFAline prints a GFA formatted segment line
func (seg *segment) PrintGFAline() string {
	if seg.optional != nil {
		return fmt.Sprintf("%v\t%v\t%v\tLN:i:%v\t%v", seg.recordType, string(seg.Name), string(seg.Sequence), seg.Length, seg.optional.printString)
	}
	return fmt.Sprintf("%v\t%v\t%v\tLN:i:%v", seg.recordType, string(seg.Name), string(seg.Sequence), seg.Length)
}

// Add checks that a segment is not already in a specified GFA isntance, then adds it
func (seg *segment) Add(gfa *GFA) error {
	if _, ok := gfa.segRecord[string(seg.Name)]; ok {
		return fmt.Errorf("Duplicate segment name already present in GFA instance: %v", string(seg.Name))
	}
	gfa.segments = append(gfa.segments, seg)
	gfa.segRecord[string(seg.Name)] = struct{}{}
	return nil
}

// A link connects oriented segments
type link struct {
	recordType string
	From       []byte
	fromOrient string
	To         []byte
	toOrient   string
	overlap    string
	optional   *optionalFields
}

// NewLink is a link constructor
func NewLink(from, fOrient, to, tOrient, overlap []byte) (*link, error) {
	if bytes.ContainsAny(from, "+-*= ") {
		return nil, fmt.Errorf("Segment name can't contain +/-/*/= or whitespace")
	}
	if bytes.ContainsAny(to, "+-*= ") {
		return nil, fmt.Errorf("Segment name can't contain +/-/*/= or whitespace")
	}
	link := new(link)
	link.From = from
	link.To = to
	link.recordType = "L"
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
	return link, nil
}

// AddOptionalFields adds a set of optional fields to a link
func (link *link) AddOptionalFields(oFs *optionalFields) {
	link.optional = oFs
}

// PrintGFAline prints a GFA formatted link line
func (link *link) PrintGFAline() string {
	if link.optional != nil {
		return fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v", link.recordType, string(link.From), link.fromOrient, string(link.To), link.toOrient, link.overlap, link.optional.printString)
	}
	return fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v", link.recordType, string(link.From), link.fromOrient, string(link.To), link.toOrient, link.overlap)
}

// Add appends a link to a specified GFA instance
func (link *link) Add(gfa *GFA) error {
	gfa.links = append(gfa.links, link)
	return nil
}

// containment
type containment struct {
	recordType string
}

// A path records a graph traversal
type path struct {
	recordType string
	PathName   []byte
	SegNames   [][]byte
	overlaps   [][]byte
	optional   *optionalFields
}

// NewPath is a path constructor
func NewPath(n []byte, segs, olaps [][]byte) (*path, error) {
	return &path{
		recordType: "P",
		PathName:   n,
		SegNames:   segs,
		overlaps:   olaps,
	}, nil
}

// PrintGFAline prints a GFA formatted segment line
func (path *path) PrintGFAline() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v", path.recordType, string(path.PathName), string(bytes.Join(path.SegNames, []byte(","))), string(bytes.Join(path.overlaps, []byte(","))))
}

// Add appends a path to a specified GFA instance
func (path *path) Add(gfa *GFA) error {
	gfa.paths = append(gfa.paths, path)
	return nil
}

// AddOptionalFields adds a set of optional fields to a path
func (path *path) AddOptionalFields(oFs *optionalFields) {
	path.optional = oFs
}

// The optional fields type is an effort to clean up the segment/containment/path types and have all the optional fields in one type
type optionalFields struct {
	readCount   string
	fragCount   string
	kmerCount   string
	checksum    []byte
	uri         string
	printString string
}

// NewOptionalFields is an optionalFields constructor
func NewOptionalFields(optional ...[]byte) (*optionalFields, error) {
	oFs := new(optionalFields)
	if len(optional) != 0 {
		for _, field := range optional {
			val := bytes.Split(field, []byte(":"))
			switch string(val[0]) {
			// segment optional fields
			case "LN":
				continue
			case "RC":
				oFs.readCount = string(val[2])
				oFs.printString = fmt.Sprintf("RC:i:%s\t%v", oFs.readCount, oFs.printString)
			case "FC":
				oFs.fragCount = string(val[2])
				oFs.printString = fmt.Sprintf("FC:i:%s\t%v", oFs.fragCount, oFs.printString)
			case "KC":
				oFs.kmerCount = string(val[2])
				oFs.printString = fmt.Sprintf("KC:i:%s\t%v", oFs.kmerCount, oFs.printString)
			case "SH":
				oFs.checksum = val[2]
				oFs.printString = fmt.Sprintf("SH:H:%s\t%v", oFs.checksum, oFs.printString)
			case "UR":
				oFs.uri = string(val[2])
				oFs.printString = fmt.Sprintf("UR:Z:%v\t%v", oFs.uri, oFs.printString)
			// TODO: add optional fields for links and containments
			default:
				continue
				//return nil, fmt.Errorf("Don't recognise optional field: %v", string(field))
			}
		}
	} else {
		return nil, fmt.Errorf("No optional fields supplied")
	}
	// trim the final tab from the printString
	oFs.printString = strings.Trim(oFs.printString, "\t")
	return oFs, nil
}
