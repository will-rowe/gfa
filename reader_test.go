package gfa

import (
	"io"
	"os"
	"testing"
)

var (
	testFile = "./example.gfa"
)

// open a GFA file and collect header/comments
func TestNewReader(t *testing.T) {
	// open a file
	fh, err := os.Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	// create a new GFA reader
	reader, err := NewReader(fh)
	if err != nil {
		t.Fatal(err)
	}
	// collect the GFA instance
	myGFA := reader.CollectGFA()
	// print some stuff
	t.Log(myGFA.PrintHeader())
	if comments := myGFA.PrintComments(); comments != "" {
		t.Log(comments)
	}
}

// open a GFA file and collect header/comments
func TestRead(t *testing.T) {
	// open a file
	fh, err := os.Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	// create a new GFA reader
	reader, err := NewReader(fh)
	if err != nil {
		t.Fatal(err)
	}
	// collect the GFA instance
	myGFA := reader.CollectGFA()
	// read the GFA file
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		// print the line
		t.Log(line)
		t.Log(line.PrintGFAline())
		// add the line to a GFA instance
		if err := line.Add(myGFA); err != nil {
			t.Fatal(err)
		}
	}
	// dump the content from a GFA instance
	t.Log("dumping content from GFA instance")
	segments, err := myGFA.GetSegments()
	if err != nil {
		t.Fatal(err)
	}
	for _, seg := range segments {
		t.Log(seg.PrintGFAline())
	}

	links, err := myGFA.GetLinks()
	if err != nil {
		t.Fatal(err)
	}
	for _, seg := range segments {
		t.Log(seg.PrintGFAline())
	}
	for _, link := range links {
		t.Log(link.PrintGFAline())
	}

}
