# Build up Symphony Agent, Piccolo

_(last edit: 11/20/2023)_

To build up Symphony Agent Piccolo. Please refer to [Symphony Agent](../docs/symphony-book/agent/symphony-agent.md) to see how to apply Piccolo.

## Prerequisites
 If you are doing this from WSL/Ubuntu, install `cargo` and `build-essential`.

   ```bash
   sudo apt-get update
   sudo apt-get install build-essential
   sudo apt install cargo
   ```
## Mage Commands
See all commands with mage -l
```
# under symphony/agent folder
> mage -l
Use this tool to quickly build Piccolo.

Targets:
  buildPiccolo    Build Symphony agent Piccolo with mode release or debug.
```
Samples
```
# under symphony/agent folder

# Build released Symphony agent Piccolo.
mage buildPiccolo release

# Build debug Symphony agent Piccolo.
mage buildPiccolo debug
```
File Structure

    .
    ├── src
    ├── target                  # artifacts would be in target folder
    │   ├── debug
    |   |   ├── piccolo         # debug piccolo
    |   |   ├── ...
    │   ├── release              
    |   |   ├── piccolo         # release piccolo
    |   |   ├── ...
    │   └── ...                 # etc.
    └── ...