package cicd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
)

// Deployer represents the production CI/CD gateway bridging the AI Auditor to the real Fabric runtime
type Deployer struct {
	NetworkPath string
	ChannelName string
}

// NewDeployer initializes the CI connection to the fabric-samples target
func NewDeployer(networkPath string, channelName string) *Deployer {
	return &Deployer{
		NetworkPath: networkPath,
		ChannelName: channelName,
	}
}

// Deploy physically spins up the test-network and commits the secured Chaincode to the ledger
func (d *Deployer) Deploy(chaincodeName string, targetFile string) error {
	fmt.Printf("\n[CI/CD Pipeline] Booting physical Hyperledger Fabric Network...\n")

	// 1. Wipe any ghost instances cleanly
	downCmd := exec.Command("./network.sh", "down")
	downCmd.Dir = d.NetworkPath
	_ = downCmd.Run() // Ignore errors if it was already down

	time.Sleep(2 * time.Second)

	// 2. Start Network and establish Crypto channels
	fmt.Printf("[CI/CD Pipeline] Standing up CAs and creating ledger channel '%s'...\n", d.ChannelName)
	upCmd := exec.Command("./network.sh", "up", "createChannel", "-c", d.ChannelName)
	upCmd.Dir = d.NetworkPath
	
	if out, err := upCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Network boot collapsed: %v\nOutput: %s", err, string(out))
	}

	// 3. Absolute path mapping for the Go compiler
	targetDir := filepath.Dir(targetFile)
	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("Invalid compiler path: %v", err)
	}

	// 4. Executing secure Production Deployment onto the Peer Network
	fmt.Printf("[CI/CD Pipeline] Committing AI-Patched Source Code. Triggering Peer Chaincode Container Build for '%s'...\n", chaincodeName)
	fmt.Printf("[CI/CD Pipeline] (Note: The Go Compiler and Fabric Docker abstractions may take 30-90+ seconds!)\n")
	
	deployCmd := exec.Command("./network.sh", "deployCC", "-ccn", chaincodeName, "-ccp", absPath, "-ccl", "go")
	deployCmd.Dir = d.NetworkPath
	
	if out, err := deployCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Chaincode deployment collapsed. Build failed: %v\nOutput: %s", err, string(out))
	}

	fmt.Printf("[CI/CD Pipeline] [SUCCESS] Target executable digitally signed, endorsed, and committed to %s!\n", d.ChannelName)
	return nil
}
