<div align="center">
  <h1>gfa</h1>
  <h3>a Go library for working with Graphical Fragment Assembly format</h3>
  <hr>
  <a href="https://travis-ci.org/will-rowe/gfa"><img src="https://travis-ci.org/will-rowe/gfa.svg?branch=master" alt="travis"></a>
  <a href="https://godoc.org/github.com/will-rowe/gfa"><img src="https://godoc.org/github.com/will-rowe/gfa?status.svg" alt="GoDoc"></a>
  <a href="https://goreportcard.com/report/github.com/will-rowe/gfa"><img src="https://goreportcard.com/badge/github.com/will-rowe/gfa" alt="goreportcard"></a>
  <a href="https://codecov.io/gh/will-rowe/gfa"><img src="https://codecov.io/gh/will-rowe/gfa/branch/master/graph/badge.svg" alt="codecov"></a>
</div>

***

## Overview

This is a Go library for working with the `Graphical Fragment Assembly` (GFA) format.

> The purpose of the GFA format is to capture sequence graphs as the product of an assembly, a representation of variation in genomes, splice graphs in genes, or even overlap between reads from long-read sequencing technology.

Read the GFA spec [here](https://github.com/GFA-spec/GFA-spec/blob/master/GFA1.md).

Current limitations:

* restricted to GFA version 1
* does not handle the containment field
* validation is limited

## Installation

``` go
go get github.com/will-rowe/gfa
```

## Example usage

### convert an MSA file to a GFA file

``` go
package main

import (
	"log"
	"os"

	"github.com/will-rowe/gfa"
)

var (
	inputFile = "./example.msa"
)

func main() {
	// open the MSA
	msa, _ := gfa.ReadMSA(inputFile)

	// convert the MSA to a GFA instance
	myGFA, err := gfa.MSA2GFA(msa)
	if err != nil {
		log.Fatal(err)
	}

	// create a gfaWriter
	outfile, err := os.Create("./example.gfa")
	defer outfile.Close()
	writer, err := gfa.NewWriter(outfile, myGFA)
	if err != nil {
		log.Fatal(err)
	}

	// write the GFA content
	if err := myGFA.WriteGFAContent(writer); err != nil {
		log.Fatal(err)
	}
}
```

### process a GFA file line by line

``` go
package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/will-rowe/gfa"
)

var (
	inputFile = flag.String("inputFile", "", "input GFA file (empty for STDIN)")
)

func main() {
	flag.Parse()
	var r io.Reader

	// open file stream and close it when finished
	if *inputFile == "" {
		r = os.Stdin
	} else {
		fh, err := os.Open(*inputFile)
		if err != nil {
			log.Fatalf("could not open file %q:", err)
		}
		defer fh.Close()
		r = fh
	}

	// create a GFA reader
	reader, err := gfa.NewReader(r)
	if err != nil {
		log.Fatal("can't read gfa file: %v", err)
	}

	// collect the GFA instance
	myGFA := reader.CollectGFA()

	// check version and print the header / comment lines
	if myGFA.GetVersion() != 1 {
		log.Fatal("gfa file is not in version 1 format")
	}
	log.Println(myGFA.PrintHeader())
	if comments := myGFA.PrintComments(); comments != "" {
		log.Println(comments)
	}

	// read the GFA file
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("error reading line in gfa file: %v", err)
		}

		// each line produced by Read() satisfies the gfaLine interface
		formattedLine := line.PrintGFAline()
		log.Printf("gfa line: %v", formattedLine)

		// you can also add the line to the GFA instance
		if err := line.Add(myGFA); err != nil {
			log.Fatal("error adding line to GFA instance: %v", err)
		}
	}
}
```
