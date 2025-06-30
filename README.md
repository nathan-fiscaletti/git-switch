# git-switch

A fast, interactive terminal UI for switching between git branches. Built with [tcell](https://github.com/gdamore/tcell) for a smooth, cross-platform experience.

Switching branches with `git checkout` or `git switch` can be slow, especially in repositories with many branches. **git-switch** makes it easy to quickly find and switch branches with just a few keystrokes.

## Features

- **Fuzzy search**: Instantly filter branches as you type.
- **Pinned Branches**: A configurable list of branches that will always show at the top of the list.
- **Keyboard navigation**: Use arrow keys to move, Enter to switch, and Esc/Ctrl+C to quit.
- **Direct checkout**: Pass a branch name as an argument to switch without the UI.

## Installation

```sh
go install github.com/nathan-fiscaletti/git-switch@latest
```

> ℹ️ I highly recommend that you alias the `git-switch` command to `sw` in your shell for eas-of-use. The rest of this documentation will make the assumption that you have. If not, use `git-switch` instead of `sw` for each command.
> 
> ```powershell
> # Windows Powershell
> "`nset-alias sw git-switch" | out-file -append -encoding utf8 $profile; . > $profile
> 
> # Bash (use .zshrc for zsh, etc.)
> echo "alias sw='git-switch'" >> ~/.bashrc && source ~/.bashrc
> ```

## Usage

### Interactive Mode

Just run:

```sh
sw
```

- Start typing to filter branches.
- Use **Up/Down** arrows to select.
- Press **Enter** to checkout the selected branch.
- Press **Esc** or **Ctrl+C** to exit.

### Direct Checkout

To checkout a branch directly (no UI):

```sh
sw <branch-name>
```

> Unless prefixed with `-x`, any arguments passed to `sw` will be forwarded to `git checkout`.

If the branch exists, you’ll be switched immediately.

### Pinned Branches

A pinned branch always shows at the top of the list of branches in the switcher.

```sh
# While a branch is checked out
sw -x pin
sw -x unpin
```

### Using git-switch as a general branch selector

You can use git-switch to select a branch and have the selected branch returned to the caller. 

When using this mode the branch will not be automatically checked out, but instead printed to stdout.

```powershell
# Windows Powershell
$branch = sw -x pipe
echo $branch

# Bash
branch=$(sw -x pipe)
echo $branch
```

### Using the interactive branch selector in your own project

The interactive branch selector is exposed in the [`pkg`](./pkg) package.

```go
package main

import (
    sw "github.com/nathan-fiscaletti/git-switch/pkg"
)

func main() {
    branchSelector, err := sw.NewBranchSelector(sw.BranchSelectorArguments{
        CurrentBranch:      currentBranch,
        Branches:           branches,
        PinnedBranches:     pinnedBranches,
        PinnedBranchPrefix: "★",
        WindowSize:         10,
        SearchLabel:        "search branch",
    })
    if err != nil {
        panic(err)
    }

    b, err := branchSelector.PickBranch()
    if err != nil {
        panic(err)
    }

    // ... use b ...
}
```

## License

MIT (See [LICENSE](./LICENSE))
