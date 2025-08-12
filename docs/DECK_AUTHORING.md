# AnCLI Deck Authoring Guide

A comprehensive guide to creating high-quality command-line learning decks with proper prerequisites, state management, and user-friendly design.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Deck Structure](#deck-structure)  
3. [Deck Configuration (deck.yaml)](#deck-configuration)
4. [Card Definitions (cards.csv)](#card-definitions)
5. [Prerequisites & Dependencies](#prerequisites--dependencies)
6. [State Management](#state-management)
7. [Container Configuration](#container-configuration)
8. [Validation & Testing](#validation--testing)
9. [Best Practices](#best-practices)
10. [Publishing & Distribution](#publishing--distribution)

---

## Quick Start

### 1. Create Your Deck Directory

```bash
mkdir my-awesome-deck
cd my-awesome-deck
```

### 2. Create deck.yaml

```yaml
name: my-awesome-deck
version: 1.0.0
author: Your Name
description: Learn awesome command-line skills
tags: [cli, tutorial, beginner]

container:
  image: alpine:3.18
  timeout: 30
  network: false

settings:
  prerequisite_mode: enforce
  show_solutions: true
  show_explanations: true
```

### 3. Create cards.csv

```csv
key,title,command,description,setup,cleanup,prerequisites,verify,hint,solution,explanation,difficulty,tags
hello,"Hello World","echo 'Hello, World!'","Your first command",,"",,"echo 'Hello, World!'","Use echo to print text","echo 'Hello, World!'","Prints text to screen. Single quotes preserve exact text. Output shows: Hello, World!",1,"basics"
```

### 4. Validate and Test

```bash
ancli deck lint .              # Validate deck structure
ancli deck test . --dry-run    # Test without containers  
ancli deck pack .              # Create .ancli package
```

---

## Deck Structure

A complete deck follows this structure:

```
my-deck/
├── deck.yaml          # Required: Deck metadata and configuration
├── cards.csv          # Required: Card definitions
├── README.md          # Optional: Deck documentation
└── assets/            # Optional: Supporting files
    ├── scripts/       # Shell scripts
    ├── configs/       # Configuration files  
    └── data/          # Sample data files
```

### Required Files

- **deck.yaml**: Metadata, container settings, learning configuration
- **cards.csv**: Individual card definitions with commands and hints

### Optional Files

- **README.md**: User-facing documentation
- **assets/**: Files that cards can reference (mounted at `/assets` in container)

---

## Deck Configuration

The `deck.yaml` file defines deck-wide settings and defaults.

### Basic Metadata

```yaml
name: linux-basics           # Unique deck identifier (lowercase, hyphens)
version: 1.2.0              # Semantic versioning (x.y.z)
author: Linux Academy        # Author name
description: Essential Linux commands for beginners
tags: [linux, cli, basics]  # Searchable tags
license: MIT                 # License (optional)
difficulty_range: [1, 4]    # Min/max difficulty levels
```

### Container Configuration

```yaml
container:
  image: alpine:3.18                # Container image to use
  timeout: 30                       # Per-command timeout (seconds)
  network: false                    # Network access (secure by default)
  environment:                      # Environment variables
    USER: student
    HOME: /home/student
    TERM: xterm-256color
  working_dir: /workspace           # Starting directory
```

### Cleanup Behavior

```yaml
cleanup:
  mode: auto                 # auto | manual | none
  preserve_on_fail: true     # Keep state if user rates "Again"
  timeout: 10                # Cleanup command timeout
```

### Learning Settings

```yaml
settings:
  shuffle_cards: false        # Randomize card order
  prerequisite_mode: enforce  # enforce | link | none  
  show_solutions: true        # Allow solution display
  show_explanations: true     # Allow explanation display
  auto_cleanup: true          # Run cleanup automatically
```

### FSRS Parameters (Optional)

```yaml
fsrs:
  request_retention: 0.9      # Target retention rate (0.8-0.95)
  maximum_interval: 365       # Max days between reviews
  initial_difficulty: 3.4     # Starting difficulty
```

---

## Card Definitions

The `cards.csv` file defines individual flashcards using a structured format.

### CSV Header (Required)

```csv
key,title,command,description,setup,cleanup,prerequisites,verify,hint,solution,explanation,difficulty,tags
```

### Field Descriptions

| Field | Required | Purpose | Example |
|-------|----------|---------|---------|
| `key` | ✅ | Unique card identifier | `create-dir` |
| `title` | ✅ | Human-readable card name | `"Create Directory"` |
| `command` | ✅ | The challenge/task for the user | `mkdir my_project` |
| `description` | ✅ | What the user should accomplish | `"Create a directory called my_project"` |
| `setup` | ❌ | Command to run before main command | `rm -rf my_project` |
| `cleanup` | ❌ | Command to run after rating | `rm -rf my_project` |
| `prerequisites` | ❌ | Comma-separated list of required cards | `basic-ls,create-file` |
| `verify` | ❌ | Command to help user check their work | `ls \| grep my_project` |
| `hint` | ❌ | Guidance without giving the answer | `"Use the mkdir command"` |
| `solution` | ❌ | Correct command(s), pipe-separated | `mkdir my_project` |
| `explanation` | ❌ | What the output means and why | `"Creates directory. No output means success."` |
| `difficulty` | ✅ | Difficulty level (1-6) | `2` |
| `tags` | ❌ | Comma-separated category tags | `"directories,basic"` |

### Understanding the Verify Field

The `verify` field contains commands that **help users self-assess** their work. These commands are **shown to the user** to teach them verification skills:

```csv
key,command,verify
create-dir,"mkdir my_project","ls | grep my_project"
create-file,"touch README.md","ls | grep README"
set-permissions,"chmod +x script.sh","ls -l script.sh | grep x"
```

**Important**: Verify commands are educational tools, not automated tests. They teach users how to check their own work.

### Example Card

```csv
key,title,command,description,setup,cleanup,prerequisites,verify,hint,solution,explanation,difficulty,tags
create-dir,"Create project directory","mkdir my_project","Create a new directory called my_project","rm -rf my_project 2>/dev/null","rm -rf my_project",,"ls | grep my_project","Use the mkdir command","mkdir my_project","Creates a directory named 'my_project' in current location. No output on success follows Unix philosophy.",1,"directories"
```

---

## Prerequisites & Dependencies

Prerequisites ensure commands execute in the right context by establishing required state.

### Understanding Prerequisites

In AnCLI, prerequisites are about **command dependencies**, not learning order. They ensure the container has the required state before executing a command.

**Example**: File operations chain
```csv
key,command,prerequisites
mkdir-project,"mkdir my_project",
cd-project,"cd my_project && pwd","mkdir-project"
create-file,"touch main.py","cd-project"
```

Each card depends on the state created by previous cards:
- `cd-project` needs the directory from `mkdir-project`
- `create-file` needs to be in the directory from `cd-project`

### Simple Prerequisites

```csv
key,command,prerequisites
basic,"mkdir workspace",
advanced,"cd workspace && pwd","basic"
```

### Multiple Prerequisites

```csv
key,command,prerequisites
final,"ls workspace/file.txt","basic,create-file"
```

### Prerequisite Modes

#### enforce (Recommended)
```yaml
prerequisite_mode: enforce
```
- Cards locked until prerequisites completed successfully
- Ensures proper command context
- Prevents errors from missing dependencies

#### link  
```yaml
prerequisite_mode: link
```
- Prerequisites auto-executed to establish state
- Allows skipping around but maintains functionality
- Good for advanced users

#### none
```yaml 
prerequisite_mode: none
```
- Prerequisites ignored
- User responsible for command context
- Only for expert users

### Dependency Best Practices

1. **Keep chains short**: Maximum 3-4 levels deep
2. **Logical progression**: Each step builds naturally on previous
3. **Clear dependencies**: Prerequisites should be obvious from the command
4. **No cycles**: Prerequisites must form a directed acyclic graph (DAG)

### Common Dependency Patterns

#### File Operations Chain
```csv
key,command,prerequisites,setup,cleanup
mkdir,"mkdir project",,"rm -rf project","rm -rf project"
cd,"cd project && pwd","mkdir","mkdir project","rm -rf project"
touch,"touch README.md","cd","mkdir project && cd project","cd .. && rm -rf project"
edit,"echo 'Hello' > README.md","touch","mkdir project && cd project && touch README.md","cd .. && rm -rf project"
```

#### Git Workflow  
```csv
key,command,prerequisites,setup,cleanup
init,"git init",,"rm -rf .git","rm -rf .git"
add,"git add .","init","git init && touch file","rm -rf .git file"
commit,"git commit -m 'Initial'","add","git init && touch file && git add .","rm -rf .git file"
```

---

## State Management

Proper state management ensures cards work consistently every time.

### Setup Commands

Run **before** the main command to establish required state:

```csv
key,command,setup
test-file,"cat config.txt","echo 'setting=value' > config.txt"
```

### Cleanup Commands

Run **after** rating to reset state for next review:

```csv
key,command,cleanup  
create-temp,"touch /tmp/test","rm -f /tmp/test"
```

### Verify Commands

Help users check their work (displayed to user):

```csv
key,command,verify
create-dir,"mkdir mydir","ls | grep mydir"
```

### State Management Patterns

#### Progressive State Building
```csv
key,command,setup,cleanup
base,"mkdir workspace",,"rm -rf workspace"
populate,"touch workspace/file","mkdir -p workspace","rm -rf workspace"  
modify,"echo 'data' > workspace/file","mkdir -p workspace && touch workspace/file","rm -rf workspace"
```

#### Isolated Operations
```csv
key,command,setup,cleanup
test-copy,"cp file1 file2","echo 'content' > file1","rm -f file1 file2"
test-move,"mv file1 renamed","echo 'content' > file1","rm -f file1 renamed"
```

#### Reset Points
```csv
key,command,setup,cleanup
checkpoint,"git status","git init && touch file && git add .","rm -rf .git file"
```

### Setup/Cleanup Guidelines

- ✅ **Always pair setup with cleanup** for reproducible practice
- ✅ **Make cleanup comprehensive** - remove all created state
- ✅ **Handle errors gracefully** - use `|| true` for non-critical cleanup
- ✅ **Test repeatability** - card should work identically on repeat
- ❌ **Don't assume clean slate** - always clean up previous state

---

## Container Configuration

### Security-First Approach

AnCLI uses secure defaults and opt-in for elevated permissions:

```yaml
container:
  image: alpine:3.18          # Minimal, secure base image
  timeout: 30                 # Reasonable timeout
  network: false              # No network by default
  # Automatic security:
  # - All capabilities dropped
  # - Read-only root filesystem  
  # - No privilege escalation
  # - Isolated from host
```

### Network Access (Use Sparingly)

```yaml
container:
  network: true              # ⚠️  Shows security warning
```

For specific cards only:
```csv
key,command,network
ping-test,"ping -c 1 google.com",true
download,"curl -s https://api.github.com",true
```

### Custom Images

```yaml
container:
  image: python:3.11-alpine   # Language-specific image
  image: ubuntu:22.04         # If you need specific tools
  image: myregistry/custom    # Private registry
```

### Environment Variables

```yaml
container:
  environment:
    LANG: en_US.UTF-8
    EDITOR: nano
    CUSTOM_SETTING: value
```

### Resource Limits

```yaml
resources:
  memory: 256m               # Memory limit
  cpu: 0.5                   # CPU limit (0.5 = half core)
  disk: 1g                   # Disk space limit
```

---

## Validation & Testing

### Deck Linting

```bash
# Basic validation
ancli deck lint .

# Verbose output with details  
ancli deck lint . --verbose

# JSON output for automation
ancli deck lint . --json
```

### Validation Categories

#### Structure Errors (Block Usage)
- Missing required files (deck.yaml, cards.csv)
- Invalid YAML/CSV format  
- File encoding issues

#### Card Errors (Block Usage)  
- Duplicate card keys
- Missing required fields (key, title, command, description)
- Invalid prerequisite references
- Circular dependencies

#### Security Warnings
- Network enabled globally
- Dangerous commands detected (sudo, rm -rf /, etc.)
- Privileged operations

#### Usability Warnings
- Missing hints/explanations
- Poor difficulty progression
- Long prerequisite chains
- Setup without cleanup

### Common Validation Messages

**CARD001: Duplicate card key**
```
Duplicate card key 'create-dir' (also used at line 5)
```
→ Ensure all card keys are unique

**CARD003: Invalid prerequisite reference**
```
Card 'advanced' references non-existent prerequisite 'basic'
```
→ Check spelling, ensure prerequisite card exists

**CARD004: Circular dependency**  
```
Circular dependency detected involving card 'card-a'
```
→ Review prerequisite chain, remove cycles

**SEC001: Network enabled globally**
```
Network access is enabled globally
```
→ Consider enabling network only for specific cards

**UX002: No hint provided**
```
Card has no hint provided
```
→ Add helpful hint without giving away the answer

### Testing Commands

```bash
# Dry run (no containers)
ancli deck test . --dry-run

# Test specific card
ancli deck test . --card create-dir

# Test with prerequisites 
ancli deck test . --card advanced --with-prereqs

# Test repeatability
ancli deck test . --card create-dir --repeat 3

# Full integration test
ancli deck test . --comprehensive
```

### Manual Testing Checklist

- [ ] All cards execute successfully
- [ ] Prerequisites work correctly  
- [ ] Setup/cleanup maintains state properly
- [ ] Verify commands help user assessment
- [ ] Solutions are accurate and complete
- [ ] Explanations add learning value
- [ ] Difficulty progression feels natural

---

## Best Practices

### Card Design

#### DO ✅
- **One concept per card** - Focus on single skill
- **Clear, specific titles** - "Create Directory" not "Files Part 1"  
- **Actionable descriptions** - Tell user exactly what to accomplish
- **Progressive difficulty** - Build skills incrementally
- **Include verification** - Teach users how to check their work
- **Explain output** - Help users understand what they see
- **Test thoroughly** - Ensure cards work reliably

#### DON'T ❌
- **Assume prior state** - Always use setup/cleanup
- **Chain too deeply** - Keep prerequisite chains under 5 levels
- **Mix concepts** - Don't combine unrelated skills in one card  
- **Give away answers** - Hints should guide, not solve
- **Ignore security** - Avoid dangerous commands without good reason
- **Skip explanations** - Output understanding is crucial for learning

### Content Guidelines

#### Hints
Guide without revealing the answer:

```csv
hint
"Use the command that 'makes directories'"
"What flag makes parent directories?"
"Hidden files start with what character?"
```

#### Solutions  
Show exact command(s) to type. Use pipe (|) to separate alternatives:

```csv
solution
"mkdir my_project"
"ls -la | ls --all -l | ls -a -l"  
"find . -name '*.py' | find . -type f -name '*.py'"
```

#### Explanations
Describe what output shows and why it matters:

```csv
explanation
"Shows all files including hidden ones starting with dot (.). The . and .. entries are current and parent directories."
"Exit code 0 means success. No output follows Unix philosophy of 'silence is golden' for successful operations."
"Creates directory structure recursively. -p flag creates parents as needed and doesn't error if directory exists."
```

#### Verify Commands
Teach users how to check their work:

```csv
verify
"ls | grep my_project"          # Check directory exists
"cat filename"                  # Check file contents  
"ls -l filename | grep x"       # Check file permissions
"ps aux | grep process"         # Check process running
"curl -s localhost:8080"        # Check service responding
```

### Difficulty Progression

| Level | Description | Example Commands |
|-------|-------------|------------------|
| 1 | Trivial - Basic commands | `ls`, `pwd`, `echo hello` |
| 2 | Easy - Simple operations | `mkdir dir`, `touch file`, `cat file` |
| 3 | Medium - Moderate complexity | `chmod +x file`, `find . -name "*.py"` | 
| 4 | Hard - Advanced features | `tar -czf backup.tar.gz dir/` |
| 5 | Expert - Complex operations | `awk '{sum+=$1} END {print sum}' file` |
| 6 | Insane - Multi-step mastery | Complex pipelines, system administration |

### Tagging Strategy

Use consistent, hierarchical tags:

```csv
tags
"basics,files"           # Category, subcategory
"git,branching"          # Tool, feature
"permissions,security"   # Concept, domain
"advanced,networking"    # Difficulty, area
```

---

## Solution and Explanation Guidelines

### Solution Field

The solution field contains the correct command(s). Use pipe (|) to separate multiple valid alternatives:

#### Single Solution
```csv
key,solution
create-dir,"mkdir my_project"
```

#### Multiple Valid Solutions
```csv
key,solution
list-all,"ls -la | ls -al | ls --all -l"
make-executable,"chmod +x file | chmod 755 file | chmod u+x file"
```

#### Solution Best Practices
- List most common/idiomatic solution first
- Include variations that teach different approaches
- Don't include wrong-but-similar commands
- Test all solutions actually work

### Explanation Field

The explanation describes what the output means and teaches understanding:

#### Good Explanation Example
```csv
key,command,solution,explanation
check-perms,"ls -l file","ls -l file","Shows permissions as -rwxr-xr-- format: first dash is file type, then three triplets for user/group/other permissions (r=read=4, w=write=2, x=execute=1)"
```

#### Explanation Components
1. **What appears**: Describe the actual output
2. **What it means**: Explain each part
3. **Why it matters**: Context for understanding
4. **Common variations**: Note platform differences

#### Examples by Category

**File Listings**
```csv
explanation
"Lists all files including hidden ones starting with dot (.). Columns show: permissions, links, owner, group, size, date, name. Total shows disk blocks used."
```

**Process Output**
```csv
explanation  
"Shows running processes. PID is process ID, %CPU/MEM show resource usage, COMMAND is the running program. TIME is cumulative CPU time used."
```

**Error Messages**
```csv
explanation
"'Permission denied' means you lack required privileges. 'No such file' means path doesn't exist. Check spelling and use tab completion."
```

**Network Commands**
```csv
explanation
"Shows 5 ICMP packets sent/received. Time=XXms is round-trip latency. 0% packet loss means connection is stable. High times indicate network issues."
```

### Hint vs Solution vs Explanation

#### Hint
Guides without revealing answer:
```csv
hint
"Use the command that 'makes directories'"
"Remember the flag for recursive operations"
"Hidden files start with a special character"
```

#### Solution  
The exact command(s) to type:
```csv
solution
"mkdir -p parent/child/grandchild"
"find . -type f -name '*.txt' | find . -name '*.txt' -type f"
```

#### Explanation
Understanding the result:
```csv
explanation
"Creates nested directories. -p flag creates parent directories as needed and doesn't error if directory exists. No output on success follows Unix philosophy."
```

### Complete Example Card

```csv
key,title,command,hint,solution,explanation,verify
disk-usage,"Check disk usage","du -sh /var/*","Use du with human-readable flag","du -sh /var/* | du -h --summarize /var/*","Shows disk usage per directory in /var. Format: size (K/M/G) then path. Largest consumers appear with M/G suffixes. -s summarizes subdirectories into single value. Useful for finding what's using disk space.","du -sh /var/*"
```

---

## Publishing & Distribution

### Packaging

```bash
# Create distributable package
ancli deck pack .

# This creates my-deck.ancli tarball containing:
# - deck.yaml
# - cards.csv  
# - README.md (if present)
# - assets/ (if present)
```

### Installation

```bash
# Install from local package
ancli deck install my-deck.ancli

# Install from URL (future)
ancli deck install https://decks.ancli.dev/linux-basics.ancli

# Install from registry (future)  
ancli deck install linux-basics
```

### Version Management

Follow semantic versioning:

- **Major** (1.0.0 → 2.0.0): Breaking changes, incompatible updates
- **Minor** (1.0.0 → 1.1.0): New cards, features, backward compatible  
- **Patch** (1.0.0 → 1.0.1): Bug fixes, typos, small improvements

### Documentation Requirements

Every published deck should include:

- **README.md** with:
  - Learning objectives
  - Prerequisites (human knowledge)
  - Estimated time commitment
  - Example cards
  - Usage instructions

- **Clear licensing** in deck.yaml

- **Comprehensive validation** - no errors, minimal warnings

### Quality Standards

Before publishing:

- [ ] 100% cards work correctly
- [ ] All validation errors resolved
- [ ] Comprehensive testing completed
- [ ] Documentation written
- [ ] Learning objectives clear
- [ ] Appropriate difficulty progression
- [ ] Good hint/explanation coverage
- [ ] Security review completed

---

## Common Patterns & Examples

### File Operations Deck
```csv
key,command,setup,cleanup,prerequisites,verify
create,"mkdir workspace",,"rm -rf workspace",,"ls | grep workspace"
enter,"cd workspace && pwd","mkdir workspace","rm -rf workspace","create","pwd | grep workspace"  
file,"touch README.md","mkdir workspace && cd workspace","cd .. && rm -rf workspace","enter","ls | grep README"
```

### Git Workflow Deck
```csv  
key,command,setup,cleanup,prerequisites,verify
init,"git init",,"rm -rf .git",,"test -d .git && echo 'Git repo initialized'"
add,"git add README.md","git init && touch README.md","rm -rf .git README.md","init","git status | grep README"
commit,"git commit -m 'Initial'","git init && touch README.md && git add .","rm -rf .git README.md","add","git log --oneline | head -1"
```

### System Administration Deck
```csv
key,command,setup,cleanup,prerequisites,verify
processes,"ps aux | head -10",,"",,"ps aux | head -1"
disk,"df -h | head -5",,"",,"df -h | head -1"  
memory,"free -h",,"",,"free -h | grep Mem"
network,"netstat -tuln | head -10",,"",,"netstat -tuln | head -1"
```

---

## Troubleshooting

### Common Issues

**Cards fail unexpectedly**
- Check setup/cleanup commands work in isolation
- Verify prerequisites are completed
- Test in clean container environment

**Prerequisites don't work**
- Check for circular dependencies
- Ensure prerequisite cards exist
- Verify setup establishes correct state

**Validation errors**
- Read error messages carefully
- Check CSV header matches exactly
- Verify YAML syntax is correct

**Container issues**
- Ensure image exists and is accessible
- Check timeout values are reasonable
- Verify commands work in target image

### Getting Help

1. **Validation output**: Use `ancli deck lint . --verbose` for detailed error information
2. **Test mode**: Use `ancli deck test . --dry-run` to check card logic without containers
3. **Documentation**: Reference this guide and example decks
4. **Community**: Join AnCLI community for deck authoring support

---

## Reference

### Complete CSV Template

```csv
key,title,command,description,setup,cleanup,prerequisites,verify,hint,solution,explanation,difficulty,tags
example-card,"Example Card","echo 'Hello'","Print greeting message",,"",,"echo 'Hello'","Use echo command","echo 'Hello'","Prints Hello to terminal. Single quotes preserve literal text.",1,"basics,examples"
```

### Validation Error Codes

| Code | Category | Description |
|------|----------|-------------|
| STRUCT001 | Structure | Missing required file |
| STRUCT002 | Structure | Invalid file format |
| DECK001 | Deck | Missing required field |
| CARD001 | Card | Duplicate card key |
| CARD002 | Card | Missing required field |
| CARD003 | Card | Invalid prerequisite |
| CARD004 | Card | Circular dependency |
| SEC001 | Security | Network enabled globally |
| UX001 | Usability | Missing explanation |
| UX002 | Usability | Missing hint |

### Useful Commands

```bash
# Validation and testing
ancli deck lint .                    # Validate deck
ancli deck test . --card KEY         # Test specific card
ancli deck pack .                    # Create .ancli package
ancli deck install DECK.ancli       # Install deck

# Development workflow
ancli deck lint . --watch           # Auto-lint on changes
ancli deck test . --verbose         # Detailed test output
```

This guide provides everything needed to create professional, educational, and secure AnCLI decks. Start with the quick start section, then dive deeper into specific areas as needed.
