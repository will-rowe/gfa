package gfa

import (
	"io"
	"os"
	"testing"
)

var (
	testFile = "./example.gfa"
	pathID = []byte("argannot~~~(Bla)SHV-191~~~KP868754:1-861")
	pathSeq = []byte("ATGCGTTATATTCGCCTGTGTATTATCTCCCTGTTAGCCACCCTGCCGCTGGCGGTACACGCCAGCCCGCAGCCGCTTGAGCAAATTAAACTAAGCGAAAGCCAGCTGTCGGGCCGCGTAGGCATGATAGAAATGGATCTGGTCAGCGGCCGCACGCTGACCGCCTGGCGCGCCGATGAACGCTTTCCCATGATGAGCACCTTTAAAGTAGTGCTCTGCGGCGCAGTGCTGGCGCGGGTGGATGCCGGTGACGAACAGCTGGAGCGAAAGATCCACTATCGCCAGCAGGATCTGGTGGACTACTCGCCGGTCAGCGAAAAACATCTTGCCGACGGCATGACGGTCGGCGAACTCTGTGCCGCCGCCATTACCATGAGCGATAACAGCGCCGCCAATCTGCTGCTGGCCACCGTCGGCGGCCCCGCAGGATTGACTGCCTTTTTGCGCCAGATCGACGACAACGTCACCCGCCTTGACCGCTGGGAAACGGAACTGAATGAGGCGCTTCCCGGCGACGCCCGCGACACCACTACCCCGGCCAGCATGGCCGCGACCCTGCGCAAGCTGCTGACCAGCCAGCGTCTGAGCGCCCGTTCGCAACGGCAGCTGCTGCAGTGGATGGTGGACGATCGGGTCGCCGGACCGTTGATCCGCTCCGTGCTGCCGGCGGGCTGGTTTATCGCCGATAAGACCGGAGCTGGCGAGCGGGGTGCGCGCGGGATTGTCGCCCTGCTTGGCCCGAATAACAAAGCAGAGCGCATTGTGGTGATTTATCTGCGGGATACCCCGGCGAGCATGGCCGAGCGAAATCAGCAAATCGCCGGGATCGGCGCGGCGCTGATCGAGCACTGGCAACGCTAA")
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
		// add the line to a GFA instance
		if err := line.Add(myGFA); err != nil {
			t.Fatal(err)
		}
	}
	// dump the content from a GFA instance
	/*
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
	for _, link := range links {
		t.Log(link.PrintGFAline())
	}
	paths, err := myGFA.GetPaths()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		t.Log(path.PrintGFAline())
		t.Log(string(path.PathName))
	}
	*/
}

// write a GFA instance to file
func TestNewWriter(t *testing.T) {
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
		// add the line to a GFA instance
		if err := line.Add(myGFA); err != nil {
			t.Fatal(err)
		}
	}
	// create a gfaWriter (overwrite original GFA)
	outfile, err := os.Create(testFile)
	defer outfile.Close()
	writer, err := NewWriter(outfile, myGFA)
	if err != nil {
		t.Fatal(err)
	}
	// write the GFA content
	if err := myGFA.WriteGFAContent(writer); err != nil {
		t.Fatal(err)
	}
}

// test the PrintSequence method
func TestPrintSequence(t *testing.T) {
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
		// add the line to a GFA instance
		if err := line.Add(myGFA); err != nil {
			t.Fatal(err)
		}
	}
	// print a sequence from the GFA
	seq, err := myGFA.PrintSequence(pathID)
	if err != nil {
		t.Fatal(err)
	}
	if string(seq) == string(pathSeq) {
		t.Log(string(pathID))
		t.Log(string(seq))
	} else {
		t.Fatal("could not extract sequence from graph")
	}
}
