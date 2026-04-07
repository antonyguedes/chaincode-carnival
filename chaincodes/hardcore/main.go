package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// HardcoreAuctionContract implements an extremely complex and vulnerable chaincode
type HardcoreAuctionContract struct {
	contractapi.Contract
}

type Auction struct {
	ID          string    `json:"id"`
	Item        string    `json:"item"`
	HighestBid  int       `json:"highest_bid"`
	HighestUser string    `json:"highest_user"`
	IsOpen      bool      `json:"is_open"`
	EndTime     time.Time `json:"end_time"`
}

// CreateAuction initializes a bidding event
// Vulnerability [DET-001]: Non-Deterministic Consensus overlap! Relying on time.Now() will completely corrupt cross-peer endorsement policies.
func (c *HardcoreAuctionContract) CreateAuction(ctx contractapi.TransactionContextInterface, id string, item string) error {
	auction := Auction{
		ID:          id,
		Item:        item,
		HighestBid:  0,
		IsOpen:      true,
		EndTime:     time.Now().Add(24 * time.Hour), // Non-Deterministic flaw!
	}

	auctionJSON, _ := json.Marshal(auction)
	return ctx.GetStub().PutState(id, auctionJSON)
}

// PlaceBid executes a volatile bid update
// Vulnerability [MVCC-001]: Time-of-Check to Time-of-Use. Massive concurrency overlap! Without structural locking or implicit version checks, rapid continuous bids will overwrite each other randomly!
func (c *HardcoreAuctionContract) PlaceBid(ctx contractapi.TransactionContextInterface, id string, bidAmount int, user string) error {
	auctionJSON, err := ctx.GetStub().GetState(id)
	if err != nil || auctionJSON == nil {
		return fmt.Errorf("Auction not found")
	}

	var auction Auction
	json.Unmarshal(auctionJSON, &auction)

	if !auction.IsOpen {
		return fmt.Errorf("Auction is closed")
	}

	if bidAmount <= auction.HighestBid {
		return fmt.Errorf("Bid must be strictly higher than %d", auction.HighestBid)
	}

	// Standard State Overlap Concurrency Failure
	auction.HighestBid = bidAmount
	auction.HighestUser = user

	updatedJSON, _ := json.Marshal(auction)
	return ctx.GetStub().PutState(id, updatedJSON)
}

// Refund removes balances
// Vulnerability [ACL-001]: Fatal Access Control leak! Anyone can type 'user' string into the CLI to drain another person's balance because it ignores GetCreator()/ClientID validation!
func (c *HardcoreAuctionContract) Refund(ctx contractapi.TransactionContextInterface, targetUser string) error {
	balanceJSON, err := ctx.GetStub().GetState("balance_" + targetUser)
	if err != nil || balanceJSON == nil {
		return fmt.Errorf("Balance unrecoverable")
	}

	// Arbitrarily modifies an external user's object!
	return ctx.GetStub().PutState("balance_" + targetUser, []byte("0"))
}

func main() {
	chaincode, _ := contractapi.NewChaincode(&HardcoreAuctionContract{})
	chaincode.Start()
}
