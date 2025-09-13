# 🚀 BlockCI-Q
**Blockchain + Quantum-Inspired CI/CD Engine**

---

## 📌 Overview
**BlockCI-Q** is a next-generation CI/CD system designed for a world where **immutability, transparency, and trust** matter as much as speed.  

Unlike traditional CI/CD tools (Jenkins, GitHub Actions, GitLab CI), BlockCI-Q:

- 🔒 Uses a **blockchain ledger** to make pipelines tamper-proof.  
- ✍️ **Cryptographically signs logs** per agent using Ed25519.  
- 📡 Follows a **Server–Agent architecture** (agents execute jobs, server dispatches).  
- ⚡ Provides a **CLI tool** to submit pipelines, verify integrity, and simulate tampering.  
- 🧪 Includes **unit + integration tests** for ledger, signing, and pipelines.  
- 🔮 Roadmap includes **quantum-resistant cryptography** and **hybrid-cloud deployments**.  

👉 This is not just another CI/CD tool — it’s CI/CD with **provable integrity**.

---

## 🏗 Architecture (Phase 10)

```text
            ┌─────────────────────────────┐
            │        Developer            │
            │   (writes pipeline.yaml)    │
            └───────────────┬─────────────┘
                            │
                            ▼
               ┌───────────────────────┐
               │        Server         │
               │  - Receives pipeline  │
               │  - Stores ledger      │
               │  - Dispatches jobs    │
               └───────────▲───────────┘
                           │
       ┌───────────────────┼───────────────────┐
       │                   │                   │
       ▼                   ▼                   ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│   Agent-1    │   │   Agent-2    │   │   Agent-N    │
│ - Poll jobs  │   │ - Execute    │   │ - Execute    │
│ - Run steps  │   │ - Return log │   │ - Return log │
└──────┬───────┘   └──────┬───────┘   └──────┬───────┘
       │                  │                  │
       └────────────── Logs + Results + Signatures ──┘
                            │
                            ▼
               ┌──────────────────────────────┐
               │      Blockchain Ledger       │
               │  - Immutable records         │
               │  - Tamper detection          │
               └──────────────────────────────┘

📂 Project Structure

blockci-q/
├── cmd/              # Entrypoints
│   ├── blockci/      # CLI tool (verify, tamper, submit)
│   ├── server/       # API server (pipelines, jobs, ledger)
│   └── agent/        # Worker agent (executes jobs)
│
├── internal/         # Main application logic
│   ├── blockchain/   # Immutable ledger (blocks, verify)
│   ├── core/         # CI/CD engine (parser, runner, scheduler)
│   ├── security/     # Key management & signatures
│   └── storage/      # Log storage
│
├── pkg/              # Utilities (hashing, config, logger)
├── configs/          # YAML configs
├── deployments/      # Docker + K8s manifests
├── scripts/          # DevOps helper scripts
├── tests/            # Unit + integration tests
├── pipeline.yaml     # Example pipeline
└── README.md         # Documentation


⚡ Features Achieved
✅ Blockchain-based Ledger → Stores every job log immutably.
✅ Digital Signatures (Ed25519) → Agents sign logs before committing.
✅ Ledger Verification → Detects tampering instantly.
✅ CLI Tool → Submit pipelines, verify ledger, simulate tampering.
✅ Server–Agent Communication → Agents register & poll for jobs.
✅ Job Dispatching → Server sends jobs → agents execute → return results.
✅ Unit & Integration Tests → For ledger, tampering detection, persistence.
✅ Pipeline YAML Parsing → Define pipeline as YAML (pipeline.yaml).

📖 Example: pipeline.yaml

agent: agent-1
stages:
  - name: Build
    steps:
      - run: echo "Compiling project..."
      - run: go build ./...
  - name: Test
    steps:
      - run: go test ./...
  - name: Deploy
    steps:
      - run: echo "Deploying application..."


🚀 Getting Started

* Build 
# Build CLI
go build -o blockci ./cmd/blockci

# Build Server
go build -o server ./cmd/server

# Build Agent
go build -o agent ./cmd/agent


* Run

# Start Server
./server   # Runs at http://localhost:8080

# Start Agent
./agent

# Submit Pipeline
./blockci submit pipeline.yaml

# Verify Ledger
./blockci verify ./ledger.jsonl

# Simulate Tampering (for testing)
./blockci tamper ./ledger.jsonl 0
./blockci verify ./ledger.jsonl   # should FAIL


🧪 Testing
Run all tests:  go test ./...

Example blockchain tampering test:  go test ./tests -run TestTamperingDetection -v


🔮 Roadmap
Phase 11 → Run real jobs with blockchain ledgering.
Phase 12 → Multi-agent scheduling & job distribution.
Phase 13 → Web Dashboard for pipelines & logs.
Phase 14 → Quantum-resistant cryptography.
Phase 15 → Hybrid-cloud deployment (Kubernetes + bare-metal).


👨‍💻 Contributing
We’d love contributors 🚀
Fork the repo.

Create a feature branch:
git checkout -b feature/my-feature

Commit changes:
git commit -m 'add my feature'

Push and open a PR.

Ways to contribute:
Improve blockchain consensus (future phases).
Add pipeline steps (Docker builds, artifact push, etc.).
Extend the CLI (blockci logs, blockci agents).
Build the dashboard UI.

📜 License
MIT License © 2025 BlockCI-Q Contributors


