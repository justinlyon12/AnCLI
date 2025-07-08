# AnCLI

## Vision  
**AnCLI** is a *CLI-first* flash-card platform that teaches real-world command-line skills with **spaced repetition**.  
Each card executes an actual shell command in a **rootless OCI container**, grades the outcome (regex match or exit-status), and reschedules the card with the **FSRS 4-parameter** algorithm. 