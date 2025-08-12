# Linux File Operations Deck

A comprehensive deck for learning essential Linux file and directory operations through hands-on practice with spaced repetition.

## Overview

This deck teaches fundamental Linux command-line skills through 20 carefully sequenced cards that build upon each other. Each card represents a real-world task you'll encounter when working with files and directories in Linux environments.

## Learning Path

The cards are designed with prerequisites that create a logical learning progression:

1. **Basic Operations** (Difficulty 1-2)
   - Creating directories and files
   - Basic listing and navigation
   - Simple file operations

2. **Intermediate Skills** (Difficulty 2-3)  
   - File content manipulation
   - Detailed listings and hidden files
   - File copying and moving

3. **Advanced Operations** (Difficulty 3-4)
   - Permissions and execution
   - File searching and archiving
   - Symbolic links and disk usage

## Key Features

### State Management
Each card includes setup and cleanup commands to ensure:
- **Consistent starting state**: Every practice session starts fresh
- **Repeatable learning**: Commands work the same way each time
- **No interference**: Previous attempts don't affect current practice

### Self-Assessment Tools
Each card provides verification commands to help you check your work:
- **Verify command**: Shows you how to confirm your solution worked
- **Example**: After creating a directory, run `ls | grep my_project` to confirm it exists
- **Learning value**: You learn both the main command AND how to verify results

### Progressive Hints
Cards provide three levels of assistance:
- **Hint**: Guidance without revealing the answer ("Use the mkdir command")
- **Solution**: The exact command(s) to use (including alternatives)
- **Explanation**: What the output means and why it matters

### Command Dependencies
Prerequisites ensure you have the necessary context:
- `enter-dir` requires `create-dir` (need directory to enter)
- `write-content` requires `create-file` (need file to write to)
- `make-executable` requires `write-content` (need content to make executable)

## Example Cards

### Basic Directory Creation
```
Task: Create project directory
Your command: mkdir my_project
Verify with: ls | grep my_project
Explanation: Creates a directory named 'my_project'. No output on success follows Unix philosophy.
```

### File Permissions
```
Task: Make file executable
Your command: chmod +x main.py && ls -l main.py  
Verify with: ls -l main.py | grep x
Explanation: Adds execute permission. The 'x' in permissions shows it's executable.
```

### Self-Verification Pattern
The verify commands teach you essential checking skills:
- **File exists**: `ls | grep filename`
- **Directory exists**: `ls | grep dirname` 
- **File contents**: `cat filename`
- **Permissions**: `ls -l filename`
- **File removed**: `ls | grep -v filename || echo 'File removed'`

## Container Environment

- **Image**: Alpine Linux 3.18 (lightweight, fast startup)
- **Working Directory**: `/workspace` 
- **Security**: Rootless container with no network access
- **Cleanup**: Automatic state reset after each card

## Usage

1. **Install the deck**:
   ```bash
   ancli deck install examples/decks/linux-file-ops.ancli
   ```

2. **Start learning**:
   ```bash
   ancli review --deck linux-file-ops
   ```

3. **Check your progress**:
   ```bash
   ancli stats --deck linux-file-ops
   ```

## Skills Covered

- **Directory Operations**: create, navigate, nested structures
- **File Management**: create, copy, move, delete, permissions  
- **Content Handling**: write, append, view file contents
- **Information Gathering**: listings, permissions, disk usage
- **Self-Verification**: checking your work with verification commands
- **Advanced Features**: symbolic links, archives, file searching

## Learning Tips

1. **Use the verify commands**: They teach you how to check your own work
2. **Read the explanation**: Understanding output is as important as knowing commands
3. **Try variations**: Many cards show multiple solution approaches  
4. **Practice verification**: Learning to check results is a crucial skill
5. **Use hints wisely**: Try first, use hints when stuck, check solutions as last resort
6. **Build on basics**: Master simple commands before moving to complex ones

## Verification Commands Reference

| Task Type | Verification Pattern | Example |
|-----------|---------------------|---------|
| File created | `ls \| grep filename` | `ls \| grep main.py` |
| Directory created | `ls \| grep dirname` | `ls \| grep my_project` |
| File contents | `cat filename` | `cat main.py` |
| File permissions | `ls -l filename` | `ls -l main.py` |
| Hidden files | `ls -la \| grep .hidden` | `ls -la \| grep .hidden` |
| File removed | `ls \| grep -v filename` | `ls \| grep -v backup.py` |
| Directory structure | `find dirname -type d` | `find project -type d` |
| Archive created | `ls -lh archive.tar.gz` | `ls -lh backup.tar.gz` |

## Prerequisites

- Basic familiarity with command-line interface concepts
- No prior Linux experience required
- Understanding that commands are case-sensitive

## Estimated Time

- **First completion**: 45-60 minutes
- **Review sessions**: 15-20 minutes  
- **Total mastery**: 2-3 weeks with regular practice

This deck provides a solid foundation for Linux system administration, development work, and general command-line proficiency while teaching you essential verification and self-assessment skills.
