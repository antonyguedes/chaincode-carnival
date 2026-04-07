package benchmark

import (
	"fmt"
	"os"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type BenchmarkContract struct {
	contractapi.Contract
}

func (c *BenchmarkContract) VulnerableACL(ctx contractapi.TransactionContextInterface, id string, value []byte) error {
	// Triggers ACL-001 (has PutState, no GetCreator)
	// We check len to avoid triggering SIZE-001 simultaneously
	if len(value) > 100 {
		return fmt.Errorf("too large")
	}
	return ctx.GetStub().PutState(id, value)
}

func (c *BenchmarkContract) VulnerableMVCC001(ctx contractapi.TransactionContextInterface, startKey string) error {
	// Triggers MVCC-001 (Unbounded range query)
	iter, _ := ctx.GetStub().GetStateByRange(startKey, "")
	if iter != nil {
		iter.Close()
	}
	return nil
}

func (c *BenchmarkContract) VulnerableMVCC002(ctx contractapi.TransactionContextInterface, id string) error {
	// Triggers MVCC-002 (Read-then-write race)
	ctx.GetStub().GetCreator() // avoid ACL-001

	val, _ := ctx.GetStub().GetState(id)
	
	if len(val) > 100 { // avoid SIZE-001
		return nil
	}
	return ctx.GetStub().PutState(id, val)
}

func (c *BenchmarkContract) VulnerableDET001(ctx contractapi.TransactionContextInterface) error {
	// Triggers DET-001 (Non-determinism)
	_ = time.Now()
	_ = os.Getenv("FOO")
	return nil
}

func (c *BenchmarkContract) VulnerableKEY001(ctx contractapi.TransactionContextInterface, pref, suff string, val []byte) error {
	// Triggers KEY-001 (State key collision)
	ctx.GetStub().GetCreator() // avoid ACL-001
	
	if len(val) > 100 { // avoid SIZE-001
		return nil
	}
	key := pref + "_" + suff
	return ctx.GetStub().PutState(key, val)
}

func (c *BenchmarkContract) VulnerableSIZE001(ctx contractapi.TransactionContextInterface, id string, val []byte) error {
	// Triggers SIZE-001 (PutState without len check)
	ctx.GetStub().GetCreator() // avoid ACL-001
	return ctx.GetStub().PutState(id, val)
}
