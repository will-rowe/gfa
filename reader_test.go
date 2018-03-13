package gfa

import (
	"io"
	"os"
	"testing"
)

var (
	testFile = "./example.gfa"
)

func TestNewReader(t *testing.T) {
	fh, err := os.Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	reader, err := NewReader(fh)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("collecting metadata from file")
	meta := reader.CollectMeta()
	t.Log("printing header and comments")
	t.Log(meta.Header())
	if comments := meta.Comments(); comments != "" {
		t.Log(comments)
	}
	t.Log("creating an empty GFA instance and adding metadata")
	myGFA, err := NewGFA()
	if err != nil {
		t.Fatal(err)
	}
	myGFA.Metadata = meta
	t.Log("reading GFA file and populating GFA")
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := line.Add(myGFA); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("dumping GFA content")
	t.Log(myGFA.Metadata.Header())
	if comments := myGFA.Metadata.Comments(); comments != "" {
		t.Log(comments)
	}
	segments := myGFA.GetSegments()
	links := myGFA.GetLinks()
	for _, seg := range segments {
		t.Log(seg.PrintGFAline())
	}
	for _, link := range links {
		t.Log(link.PrintGFAline())
	}

}
