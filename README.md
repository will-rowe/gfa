<div align="center">
  <h1>gfa</h1>
  <h3>a Go library for working with Graphical Fragment Assembly format</h3>
  <hr>
  <a href="https://travis-ci.org/will-rowe/gfa"><img src="https://travis-ci.org/will-rowe/gfa.svg?branch=master" alt="travis"></a>
  <a href="https://godoc.org/github.com/will-rowe/gfa"><img src="https://godoc.org/github.com/will-rowe/gfa?status.svg" alt="GoDoc"></a>
  <a href="https://goreportcard.com/report/github.com/will-rowe/gfa"><img src="https://goreportcard.com/badge/github.com/will-rowe/gfa" alt="goreportcard"></a>
  <a href="https://codecov.io/gh/will-rowe/gfa"><img src="https://codecov.io/gh/will-rowe/gfa/branch/master/graph/badge.svg" alt="codecov"></a>
  <a href=""><img src="https://img.shields.io/badge/status-unstable-red.svg" alt="unstable"></a>
</div>

***

## Overview

This is a Go library for working with the `Graphical Fragment Assembly` (GFA) format.

`This is a work in progress -- please check back soon...`

## Installation

``` go
go get github.com/will-rowe/gfa
```

## Example usage

This is just a basic example for now:
* creates a GFA reader to read a GFA file from STDIN/disk
* creates a GFA instance to store GFA data
* checks the version and prints header /  comments
* reads/prints the GFA lines
* store lines in a GFA instance

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
			log.Fatal("error line to GFA instance: %v", err)
		}
	}
}
```
