# SCPN: Simplicity Contract Propagation on Nostr

> A decentralized discovery and interaction layer for Liquid Smart Contracts, powered by Nostr.

This project is a proof-of-concept building a foundational layer for smart contract interoperability on the Liquid Network. It uses the Nostr protocol as a public, censorship-resistant "bulletin board" where users can publish, discover, and interact with Simplicity smart contracts.

## 🚀 Project Goal

Currently, deploying a complex contract on Liquid (using Simplicity or Miniscript) is a "fire-and-forget" action. There is no standard way for other users or wallets to:
1.  Discover that the contract exists.
2.  Understand its spending conditions (witness data).
3.  Find all the necessary components (contract hex, taproot data) to build a valid spending transaction.

This project solves the discovery problem by creating a **standardized metadata format** for Simplicity contracts and **propagating it as a `kind: 1` event on Nostr**. This allows any compatible client to find and interpret these contracts, paving the way for a more open and interactive smart contract ecosystem.

## 🛠️ Architecture Overview

This project is built as a monorepo with two main components:

* **Frontend (SvelteKit):**
    * Written in **SvelteKit** with **TypeScript** and styled with **Tailwind CSS**.
    * Uses **NDK (Nostr Dev Kit)** to communicate with Nostr relays (publishing and subscribing).
    * Integrates with **NIP-07** browser extensions (like Alby or nos2x) for user authentication and event signing.
    * Provides a UI for:
        1.  Dynamically creating a new contract (e.g., an oracle-based vault).
        2.  Publishing the contract's metadata to Nostr.
        3.  Viewing a live-updating table of all contracts published with this protocol.

* **Backend (Go):**
    * A simple REST API built with **Go** and the **Fiber** web framework.
    * Exposes endpoints to receive contract parameters from the frontend.
    * Executes local **shell scripts** that wrap calls to the `simc` (SimplicityHL) compiler and `hal-simplicity` CLI.
    * This process compiles the human-readable Simplicity code into raw hex, generates the taproot address, and computes all necessary metadata (like the control block) required for spending.
    * Returns the complete contract data as a JSON object to the frontend, ready for publishing.

## Prerequisites

Before you begin, ensure you have the following tools installed and configured on your system.

### 1. Nostr
* **A Nostr Account:** You need a keypair (public `npub...` and private `nsec...`).
* **NIP-07 Browser Extension:** This is required for logging into the web app.
    * [nos2x](https://github.com/nobs-lol/nos2x) (Recommended)
    * [Alby](https://getalby.com/) 
    * [Fina](https://fina.cash/)

### 2. Core Dependencies
* **Rust & Cargo:** Required to build the Simplicity tools.
    ```bash
    curl --proto '=https' --tlsv1.2 -sSf [https://sh.rustup.rs](https://sh.rustup.rs) | sh
    source "$HOME/.cargo/env"
    ```
* **Go (latest stable):** Powers the backend API.
    * Visit the [Official Go Website](https://go.dev/doc/install) for installation instructions.

* **Node.js (v22+):** Required for the SvelteKit frontend.
    * We recommend using `nvm` (Node Version Manager) to manage Node versions.
    ```bash
    curl -o- [https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh](https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh) | bash
    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
    [ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"
    
    # Install and use Node 22
    nvm install 22
    nvm use 22
    ```

## 🚀 Setup and Installation

Follow these steps to build the dependencies and run the project.

### 1. Clone the Simplicity Tools

You must compile `simplicity-hl` (which provides `simc`) and `hal-simplicity` from source.

```bash
# Clone and build simplicity-hl
git clone [https://github.com/BlockstreamResearch/SimplicityHL.git](https://github.com/BlockstreamResearch/SimplicityHL.git)
cd SimplicityHL
cargo build --release
# Make sure the binary is in your PATH. Example:
# sudo cp target/release/simc /usr/local/bin/

# Go back to your main projects directory
cd ..

# Clone and build hal-simplicity
git clone [https://github.com/BlockstreamResearch/hal-simplicity.git](https://github.com/BlockstreamResearch/hal-simplicity.git)
cd hal-simplicity
cargo build --release
# Make sure the binary is in your PATH. Example:
# sudo cp target/release/hal-simplicity /usr/local/bin/

# Verify that both are in your PATH
which simc
which hal-simplicity
# Both commands should return a path (e.g., /usr/local/bin/simc)
```

### 2. Clone the MareNostrum repository
```bash
git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
cd YOUR_REPO_NAME
```

### 3. Run front-end Svelte
```bash
npm install
npm run dev
```

### 4. Run back-end Go Fiber
```bash
cd backend-go
go get .
go run .
```

It should now be runnig everything and you can access MareNostrum at http://localhost:5173