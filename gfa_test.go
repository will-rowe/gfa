package gfa

import (
	"testing"
)

// initialise GFA
func TestNewGFA(t *testing.T) {
	t.Log("creating a GFA instance")
	_, err := NewGFA()
	if err != nil {
		t.Fatal(err)
	}
}

// initalise metadata
func TestMetadata(t *testing.T) {
	t.Log("initalise metadata")
	meta := NewHeader()
	t.Log("adding version number, printing header and version:")
	err := meta.AddVersionNumber(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(meta.Header())
	t.Log(meta.Version())
	t.Log("adding bad version number:")
	err = meta.AddVersionNumber(3)
	if err != nil {
		t.Log(err)
	}
	t.Log("adding and printing comments:")
	meta.AddComment([]byte("a gfa comment"))
	meta.AddComment([]byte("another one"))
	if comments := meta.Comments(); comments != "" {
		t.Log(comments)
	} else {
		t.Fatal()
	}
	t.Log("creating a GFA instance, linking metadata and printing header")
	myGFA, err := NewGFA()
	if err != nil {
		t.Fatal(err)
	}
	myGFA.Metadata = meta
	t.Log(myGFA.Metadata.Header())
}

// initialise GFA and add some segments
func TestNewSegment(t *testing.T) {
	t.Log("creating some segments")
	seg, err := NewSegment([]byte("1"), []byte("actg"))
	if err != nil {
		t.Fatal(err)
	}
	seg2, err := NewSegment([]byte("2"), []byte("aaaaatgacgt"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("printing the segments:")
	t.Log(seg.PrintGFAline())
	t.Log("adding the segments to a GFA instance")
	myGFA, err := NewGFA()
	if err != nil {
		t.Fatal(err)
	}
	if err = seg.Add(myGFA); err != nil {
		t.Fatal(err)
	}
	if err = seg2.Add(myGFA); err != nil {
		t.Fatal(err)
	}
	t.Log("printing the segments from the GFA:")
	for _, seg := range myGFA.GetSegments() {
		t.Log(seg.PrintGFAline())
	}
	t.Log("testing a badly named segment:")
	_, err = NewSegment([]byte("+ - 2"), []byte("actg"))
	if err != nil {
		t.Log(err)
	}
	t.Log("testing a duplicate named segment:")
	seg4, err := NewSegment([]byte("1"), []byte("actg"))
	if err != nil {
		t.Log(err)
	}
	if err = seg4.Add(myGFA); err != nil {
		t.Log(err)
	}
}
