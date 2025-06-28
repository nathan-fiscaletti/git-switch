# git-switch

A fast, interactive terminal UI for switching between git branches. Built with [tcell](https://github.com/gdamore/tcell) for a smooth, cross-platform experience.

Switching branches with `git checkout` or `git switch` can be slow, especially in repositories with many branches. **git-switch** makes it easy to quickly find and switch branches with just a few keystrokes.

## Features

- **Fuzzy search**: Instantly filter branches as you type.
- **Keyboard navigation**: Use arrow keys to move, Enter to switch, and Esc/Ctrl+C to quit.
- **Direct checkout**: Pass a branch name as an argument to switch without the UI.

## Installation

```sh
go install github.com/nathan-fiscaletti/git-switch@latest
```

> I highly recommend that you alias this to `sw` in your shell for eas-of-use.

## Usage

### Interactive Mode

Just run:

```sh
git-switch
```

- Start typing to filter branches.
- Use **Up/Down** arrows to select.
- Press **Enter** to checkout the selected branch.
- Press **Esc** or **Ctrl+C** to exit.

### Direct Checkout

To checkout a branch directly (no UI):

```powershell
git-switch <branch-name>
```

If the branch exists, you’ll be switched immediately.

## License

MIT (See [LICENSE](./LICENSE))