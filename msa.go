package gfa

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/io/seqio/alignio"
	"github.com/biogo/biogo/io/seqio/fasta"
	"github.com/biogo/biogo/seq"
	"github.com/biogo/biogo/seq/linear"
	"github.com/biogo/biogo/seq/multi"
)

// ReadMSA will read in an MSA file and store it as a Multi (MSA)
func ReadMSA(fileName string) (*multi.Multi, error) {
	// open a file
	fh, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Can't open file: %v", fileName)
	}
	defer fh.Close()
	// read in the MSA
	r := fasta.NewReader(fh, linear.NewSeq("", nil, alphabet.DNA))
	m, _ := multi.NewMulti("", nil, seq.DefaultConsensus)
	msa, err := alignio.NewReader(r, m).Read()
	if err != nil {
		return nil, fmt.Errorf("Can't read MSA from file: %v", fileName)
	}
	return msa, nil
}

// MSA2GFA converts an MSA to a GFA instance
func MSA2GFA(msa *multi.Multi) (*GFA, error) {
	// create an empty GFA instance and then add version (1)
	myGFA := NewGFA()
	err := myGFA.AddVersion(1)
	if err != nil {
		return nil, err
	}
	// clean the MSA to remove consensus seq
	msaCleaner(msa)
	// create nodes for each unique base in every column of the alignment
	msaNodes, _ := getNodes(msa)
	// draw edges between nodes
	if err := msaNodes.drawEdges(); err != nil {
		return nil, err
	}
	// squash neighbouring nodes and update edges
	if err := msaNodes.squashNodes(); err != nil {
		return nil, err
	}
	// populate the GFA instance
	orderedList := msaNodes.getOrderedList()
	for _, nodeID := range orderedList {
		node := msaNodes.nodeHolder[nodeID]
		// create the segment(s)
		seg, err := NewSegment([]byte(strconv.Itoa(nodeID)), []byte(node.base))
		if err != nil {
			return nil, err
		}
		seg.Add(myGFA)
		// create link(s)
		for outEdge := range node.outEdges {
			link, err := NewLink([]byte(strconv.Itoa(nodeID)), []byte("+"), []byte(strconv.Itoa(outEdge)), []byte("+"), []byte("0M"))
			if err != nil {
				return nil, err
			}
			link.Add(myGFA)
		}
	}
	// get paths for each MSA entry
	type pathHolder struct {
		seqID    []byte
		segments [][]byte
		overlaps [][]byte
	}
	for _, seqID := range msaNodes.seqIDs {
		tmpPath := &pathHolder{seqID: []byte(seqID)}
		for _, nodeID := range orderedList {
			node := msaNodes.nodeHolder[nodeID]
			for _, seq := range node.parentSeqIDs {
				if seq == seqID {
					segment := strconv.Itoa(nodeID) + "+"
					overlap := strconv.Itoa(len(node.base)) + "M"
					tmpPath.segments = append(tmpPath.segments, []byte(segment))
					tmpPath.overlaps = append(tmpPath.overlaps, []byte(overlap))
					break
				}
			}
		}
		// add the path
		path, err := NewPath(tmpPath.seqID, tmpPath.segments, tmpPath.overlaps)
		if err != nil {
			return nil, err
		}
		path.Add(myGFA)
	}
	return myGFA, nil
}

// msaCleaner removes the consensus entry (if present) from a MSA
func msaCleaner(msa *multi.Multi) {
	for i := 0; i < msa.Rows(); i++ {
		id := msa.Row(i).Name()
		if id == "consensus" {
			msa.Delete(i)
		}
	}
}

// the node type is used to store unique bases and associated IDs from each MSA column
type node struct {
	parentSeqIDs []string
	base         string
	inEdges      map[int]struct{}
	outEdges     map[int]struct{}
}

// the msaNodes store all the nodes from a MSA
type msaNodes struct {
	nodeHolder map[int]*node
	seqIDs     []string
}

// this function returns a sorted list of node IDs, sorted by increasing ID value
// it's used to iterate over the nodeHolder nodes in order
func (msaNodes *msaNodes) getOrderedList() []int {
	list := []int{}
	for node := range msaNodes.nodeHolder {
		list = append(list, node)
	}
	sort.Ints(list)
	return list
}

// getNodes is an msaNodes constructor. It moves through each column of an MSA, making a node for each unique base per column
func getNodes(msa *multi.Multi) (*msaNodes, error) {
	nodeHolder := make(map[int]*node)
	nodeID := 1
	// get all the entry ids
	ids := []string{}
	for i := 0; i < msa.Rows(); i++ {
		ids = append(ids, string(msa.Row(i).Name()))
	}
	// process each column of the MSA
	for i := 0; i < msa.Len(); i++ {
		columnBases := msa.Column(i, true)
		// keep a record of seen bases so that identical bases can be consolidated
		record := make(map[string][]string)
		// process each row of the column
		for rowIterator, base := range columnBases {
			stringifiedBase := string(base)
			if _, ok := record[stringifiedBase]; !ok {
				record[stringifiedBase] = []string{ids[rowIterator]}
			} else {
				record[stringifiedBase] = append(record[stringifiedBase], ids[rowIterator])
			}
		}
		// convert the record of bases and ids to a slice of nodes for each column
		columnNodes := []*node{}
		for base, ids := range record {
			// treat all gaps as individual nodes
			if base == "-" {
				for _, id := range ids {
					columnNodes = append(columnNodes, &node{parentSeqIDs: []string{id}, base: "", outEdges: make(map[int]struct{}), inEdges: make(map[int]struct{})})
				}
			} else {
				columnNodes = append(columnNodes, &node{parentSeqIDs: ids, base: base, outEdges: make(map[int]struct{}), inEdges: make(map[int]struct{})})
			}
		}
		// add the nodes for the column to the msaNodes
		for _, node := range columnNodes {
			nodeHolder[nodeID] = node
			nodeID++
		}
	}
	// construct the msaNodes
	msaNodes := &msaNodes{nodeHolder: nodeHolder, seqIDs: ids}
	return msaNodes, nil
}

// drawEdges connects neighbouring nodes derived from the same MSA entry
func (msa *msaNodes) drawEdges() error {
	// getFirstNode returns the nodeID for the first node in the MSA derived from a specified sequence
	getFirstNode := func(seqID string) int {
		for nodeID := 1; nodeID <= len(msa.nodeHolder); nodeID++ {
			for _, id := range msa.nodeHolder[nodeID].parentSeqIDs {
				if seqID == id {
					return nodeID
				}
			}
		}
		return 0
	}
	// findNextNode returns the ID of the next node that is derived from a query MSA sequence
	findNextNode := func(seqID string, startNode int) int {
		for nextNode := startNode + 1; nextNode <= len(msa.nodeHolder); nextNode++ {
			for _, parentSeqID := range msa.nodeHolder[nextNode].parentSeqIDs {
				if seqID == parentSeqID {
					return nextNode
				}
			}
		}
		return 0
	}
	// iterate over each MSA sequence, connecting edges for each one
	for _, seqID := range msa.seqIDs {
		startNode := getFirstNode(seqID)
		if startNode == 0 {
			return fmt.Errorf("Node parse error: Could not identify start node for %v", seqID)
		}
		for {
			nextNode := findNextNode(seqID, startNode)
			if nextNode == 0 {
				break
			}
			// draw edges
			msa.nodeHolder[startNode].outEdges[nextNode] = struct{}{}
			msa.nodeHolder[nextNode].inEdges[startNode] = struct{}{}
			startNode = nextNode
		}
	}
	return nil
}

// squashNodes collapses neighbouring nodes into a single node if no branches exist
func (msa *msaNodes) squashNodes() error {
	squashableNodes := make(map[int]int)
	squashNodesSortList := []int{}
	// iterate through all the nodes in order and identify squashable nodes
	for nodeIterator := 1; nodeIterator <= len(msa.nodeHolder); nodeIterator++ {
		// if there is only one node connected via an out edge, see if the out node has multiple in edges
		if len(msa.nodeHolder[nodeIterator].outEdges) == 1 {
			var outNode int
			for key := range msa.nodeHolder[nodeIterator].outEdges {
				outNode = key
			}
			// if the out node has only one in edge, we can squash these nodes together
			if len(msa.nodeHolder[outNode].inEdges) == 1 {
				squashableNodes[outNode] = nodeIterator
				squashNodesSortList = append(squashNodesSortList, outNode)
			}
		}
	}
	// reverse sort the identified squashale nodes and squash them
	sort.Sort(sort.Reverse(sort.IntSlice(squashNodesSortList)))
	for _, outNode := range squashNodesSortList {
		inNode := squashableNodes[outNode]
		msa.nodeHolder[inNode].base += msa.nodeHolder[outNode].base
		delete(msa.nodeHolder, outNode)
	}
	// remove all gaps, create new IDs and clear edges
	newNodeHolder := make(map[int]*node)
	nodeNamer := 1
	orderedList := msa.getOrderedList()
	for _, i := range orderedList {
		if msa.nodeHolder[i].base != "" {
			msa.nodeHolder[i].outEdges = make(map[int]struct{})
			msa.nodeHolder[i].inEdges = make(map[int]struct{})
			newNodeHolder[nodeNamer] = msa.nodeHolder[i]
			nodeNamer++
		}
	}
	msa.nodeHolder = newNodeHolder
	// draw edges between the new nodes
	if err := msa.drawEdges(); err != nil {
		return err
	}
	return nil
}
