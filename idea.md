- Configure "focus" branches, these branches will always be displayed at the top
  of the list. These branches can have aliases supported, which will be respected
  when you run the program with like `git checkout`.
- it can currently be run as `git checkout`, but what about `git branch`?
- If you pass more than one argument, the arguments are passed to git branch?
- Make this configurable? Default behavior? Or each have their own flag?

```
sw -b <some `git branch` args>
sw ... <treated as `git checkout`, but wrapped with the branch selector>
sw -x <internal command, such as configuring focus branches>
sw -x focus <adds current branch to focus branches for repository>
sw -x unfocus <unfocuses current branch>
```