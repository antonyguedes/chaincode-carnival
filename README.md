# 🏴‍☠️ Chaincode Carnival: Adversarial Arena

![Chaincode Carnival Arena Logo](public/django.jpeg)

**Chaincode Carnival** is an enterprise-grade AI security orchestration pipeline designed to autonomously audit, exploit, and secure Hyperledger Fabric smart contracts. 

Unlike traditional static scanners, Chaincode Carnival uses a **Multi-Agent Rivalry Engine**. It pits four distinct AI personalities against each other in a virtual "Cyber Arena" to ensure your chaincode is mathematically sound before it ever touches a production ledger.

---

## 🏛️ The Arena Architecture

The system orchestrates four specialized agents, each of which can be powered by different LLM providers (Gemini, Groq, OpenAI, Ollama, etc.):

1.  **🔍 The Analyzer**: Scans the source code using dynamic AST parsing and semantic reasoning to find deep logic flaws (MVCC collisions, unbounded loops, access control leaks).
2.  **👺 The Red Team**: Takes the findings and generates malicious exploit payloads. It spins up **ephemeral Docker Sandboxes** to physically prove the exploit works.
3.  **🛡️ The Blue Team**: Receives the successful breach report and generates a **Unified Diff Patch** to secure the code recursive rounds until the Red Team's payloads bounce.
4.  **⚖️ The Skeptic**: A senior auditor judge that scores the final patch on a 4-dimension rubric (Security, Readability, Idiomatic Go, and Completeness).

---

## 🚀 Installation

### 1. Prerequisites
- **Go**: 1.20 or higher
- **Docker**: For running the exploit sandboxes
- **Fabric Samples**: A copy of `fabric-samples` at the project root (used for the final CI/CD deployment stage)

### 2. Environment Setup
Create a `.env` file based on `.env.example`. You can mix and match providers for each agent:
```env
# Example using Gemini for everything
ANALYZER_PROVIDER="gemini"
ANALYZER_MODEL="gemini-1.5-flash"
ANALYZER_API_KEY="your_api_key_here"

# ... (repeat for REDTEAM, BLUETEAM, SKEPTIC)
```

### 3. Build & Global Command
To build the binary in the project root:
```bash
go build -o hfchance ./cmd/hfchance
```

#### Running locally:
```bash
./hfchance carnival start
```

#### To run as `hfchance` (Global PATH):
**Bash / Zsh**:
```bash
export PATH=$PATH:.
hfchance carnival start
```

**Fish Shell**:
```fish
set -px PATH .
hfchance carnival start
```

---

## 🎮 Running the Carnival

Start the interactive arena by running:
```bash
hfchance carnival start
```

### The Workflow:
1.  **Interactive Menu**: Select which chaincode file to drop into the arena (sorted by most recent).
2.  **The Rumble**: High-intensity AI banter as they battle over the code logic (with dramatic 5-second pauses).
3.  **CI/CD Gate**: If the **Blue Team wins** (Score > 90/100), the system automatically:
    - Generates a secured `.go` version of your code.
    - Boots the internal Fabric Test Network.
    - **Deploys the secured chaincode** to the channel.
4.  **Failure**: If the **Red Team wins**, the pipeline blocks the deployment and alerts the developers of the unpatched risks.

---

## 🛠️ Advanced Usage

**Target specific files manually:**
```bash
./hfchance carnival start /path/to/my/custom_chaincode.go
```

**Clean Docker subnet overlaps (if sandboxes stall):**
```bash
docker network prune -f
```

---

## 🏆 Project Accomplishments
- Native Go AST parsing for Zero-Config chaincode discovery.
- Resilient JSON extraction for chatty LLMs.
- Multi-provider support (Anthropic, Cohere, Deepseek, Groq, Gemini, Ollama).
- Physical CI/CD integration with Hyperledger Fabric `test-network`.
