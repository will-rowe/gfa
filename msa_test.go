package gfa

import (
	"testing"
)

var (
	testMSAfile = "./example.msa"
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

// test for creating a GFA instance from an MSA
func TestMSA2GFA(t *testing.T) {
	msa, err := ReadMSA(testMSAfile)
	if err != nil {
		t.Fatal(err)
	}
	gfa, err := MSA2GFA(msa)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(gfa.PrintHeader())
}

func TestGetNodesAndEdges(t *testing.T) {
	msa, _ := ReadMSA(testMSAfile)
	msaNodes, _ := getNodes(msa)
	for _, seqID := range msaNodes.seqIDs {
		t.Log(seqID)
	}
	err := msaNodes.drawEdges()
	if err != nil {
		t.Fatal(err)
	}
}
