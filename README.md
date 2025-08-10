# AnCLI

**AnCLI** is a *CLI-first* flashcard platform that teaches real-world command-line skills with **spaced repetition**.  
Each card executes an actual shell command in a **rootless OCI container**, lets you grade yourself (`Again | Hard | Good | Easy`), and reschedules with the **FSRS 4-parameter** algorithm.

## Quick Start

### Prerequisites
- **Go 1.24+** for building from source
- **Podman** (rootless) or **Docker** for container execution
- **Linux/macOS** (Windows support via WSL2)

### Installation

```bash
# Clone and build
git clone https://github.com/justinlyon12/AnCLI.git
cd AnCLI
make build

# Move binary to PATH
sudo mv bin/ancli /usr/local/bin/
```

### Your First Review Session

```bash
# Start a review session (will create database on first run)
ancli review

# Review specific deck
ancli review --deck-id=1

# Limit session to 10 cards  
ancli review --max-cards=10

# Review only new/unlearned cards
ancli review --new-only
```

## Core Concepts

- **🎯 Cards**: Command-line challenges with expected execution
- **📚 Decks**: Collections of related cards (e.g., "Git Basics", "Linux Admin")  
- **⭐ Ratings**: `1=Again`, `2=Hard`, `3=Good`, `4=Easy` determine next review
- **🔒 Security**: Rootless containers with capability dropping by default
- **🧠 FSRS**: Optimized spaced repetition scheduling for efficient learning

## Usage Examples

```bash
# Basic review session
ancli review

# Advanced filtering
ancli review --deck-id=2 --due-only --max-cards=15 --shuffle=false

# Enable network for specific cards (shows security warning)
ancli review --no-network=false

# Debug logging
ancli review --log-level=debug

# Custom database location
ancli review --database-path=/tmp/test.db
```

## Documentation

- **[Getting Started Guide](docs/GETTING-STARTED.md)** - Step-by-step first-use tutorial
- **[Usage Guide](docs/USAGE.md)** - Complete command reference and workflows
- **[Deck Format](docs/DECK-FORMAT.md)** - Creating and structuring deck files
- **[Design Documentation](docs/)** - Architecture and design decisions

