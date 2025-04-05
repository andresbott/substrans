package main

import (
	"github.com/andresbott/substrans/app/cmd"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	cmd.Execute()
}

var _ = spew.Dump
