package gfa

import (
	"os"
	"testing"
)

var (
	testMSAfile = "./example.msa"
	testGFAfile = "./example.gfa"
	newGFAfile = "./converted-msa.gfa"
)

// test for loading MSA from file
func TestMSAload(t *testing.T) {
	msa, err := ReadMSA(testMSAfile)
	if err != nil {
		t.Fatal(err)
	}
	// alignment length
	t.Log(msa.Len())
}

// test for running the clean function to remove consensus seq from MSA
func TestMSAcleaner(t *testing.T) {
	msa, err := ReadMSA(testMSAfile)
	if err != nil {
		t.Fatal(err)
	}
	beforeCleaning := msa.Rows()
	msaCleaner(msa)
	afterCleaning := msa.Rows()
	if afterCleaning != (beforeCleaning - 1) {
		t.Fatal("msaCleaner did not remove the consensus entry")
	}
}

// test to create nodes for each unique base in every column of the alignment
func TestGetNodesAndEdges(t *testing.T) {
	msa, _ := ReadMSA(testMSAfile)
	msaNodes, _ := getNodes(msa)
	// draw edges between nodes
	err := msaNodes.drawEdges()
	if err != nil {
		t.Fatal(err)
	}
}


// test for converting MSA to a GFA file - combines all the functions tested above
func TestMSA2GFA(t *testing.T) {
	// convert the MSA
	msa, err := ReadMSA(testMSAfile)
	if err != nil {
		t.Fatal(err)
	}
	myGFA, err := MSA2GFA(msa)
	if err != nil {
		t.Fatal(err)
	}
	// create a gfaWriter
	outfile, err := os.Create(newGFAfile)
	defer outfile.Close()
	writer, err := NewWriter(outfile, myGFA)
	if err != nil {
		t.Fatal(err)
	}
	// write the GFA content
	if err := myGFA.WriteGFAContent(writer); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(newGFAfile); err != nil {
		t.Fatal(err)
	}
}
