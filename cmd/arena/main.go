package main

import (
	"os"
	"github.com/antonioforte/chaincode-carnival/agents/arena"
)

func main() {
	target := "chaincodes/benchmark/benchmark.go"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}
	arena.Run(target)
}
