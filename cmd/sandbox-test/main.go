package main

import (
	"fmt"
	"log"

	"github.com/antonioforte/chaincode-carnival/agents/exploiter/sandbox"
)

func main() {
	fmt.Println("Initializing Session Manager...")
	manager := sandbox.NewSessionManager()

	fmt.Println("Spawning mock session 'test-run-1'...")
	sess, err := manager.Spawn("test-run-1", "")
	if err != nil {
		log.Fatalf("Spawn failed: %v", err)
	}

	fmt.Printf("Successfully spawned Sandbox!\nSession ID: %s\nSubnet: %s\nCompose File: %s\n", sess.SessionID, sess.Subnet, sess.ComposePath)

	fmt.Println("Tearing down the sandbox...")
	if err := manager.Teardown(sess); err != nil {
		log.Fatalf("Teardown failed: %v", err)
	}

	fmt.Println("Teardown successful. All ephemeral networks and volumes cleared!")
}
