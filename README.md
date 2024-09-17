# wal
golang: distributed WAL (Write Ahead Log)

# git commands
```console
git clone git@github.com:percybear/wal.git --config core.sshCommand="ssh -i ~/.ssh/github_ed25519"
```
## delete local branch
```console
git branch -d branch_name
```
* Before deleting a local branch, make sure to switch to another branch that you do NOT want to delete, with the git checkout command.
* If the branch contains unmerged changes and unpushed commits, the -d flag will not allow the local branch to be deleted.

## delete remote branch
```console
git push origin -d branch name
```
