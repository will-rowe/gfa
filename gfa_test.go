package gfa

import (
	"testing"
)

// initialise GFA, add header and comments
func TestHeader(t *testing.T) {
	t.Log("creating a GFA instance")
	myGFA := NewGFA()
	t.Log("adding version number, printing header and version:")
	err := myGFA.AddVersion(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(myGFA.PrintHeader())
	t.Log(myGFA.GetVersion())
	t.Log("adding bad version number:")
	err = myGFA.AddVersion(3)
	if err != nil {
		t.Log(err)
	}
	t.Log("adding and printing comments:")
	myGFA.AddComment([]byte("a gfa comment"))
	myGFA.AddComment([]byte("another one"))
	if comments := myGFA.PrintComments(); comments != "" {
		t.Log(comments)
	} else {
		t.Fatal()
	}
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
	t.Log("checking empty GFA gives error for GetSegments")
	myGFA := NewGFA()
	// test for GFA with no segments
	if _, err := myGFA.GetSegments(); err != nil {
		t.Log(err)
	}
	t.Log("adding the segments to a GFA instance")
	if err = seg.Add(myGFA); err != nil {
		t.Fatal(err)
	}
	if err = seg2.Add(myGFA); err != nil {
		t.Fatal(err)
	}
	t.Log("printing the segments from the GFA:")
	segments, err := myGFA.GetSegments()
	if err != nil {
		t.Fatal(err)
	}
	for _, seg := range segments {
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

// validate a GFA instance
func TestGFAvalidity(t *testing.T) {
	myGFA := NewGFA()
	if err := myGFA.Validate(); err != nil {
		t.Log(err)
	}
	err := myGFA.AddVersion(1)
	if err != nil {
		t.Fatal(err)
	}
	if err := myGFA.Validate(); err != nil {
		t.Log(err)
	}
	seg, err := NewSegment([]byte("1"), []byte("actg"))
	if err != nil {
		t.Fatal(err)
	}
	if err = seg.Add(myGFA); err != nil {
		t.Fatal(err)
	}
	if err := myGFA.Validate(); err != nil {
		t.Fatal(err)
	}
}
