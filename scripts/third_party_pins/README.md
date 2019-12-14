Fabric-sdk-go reuses fabric and fabric-ca code by first copying
desired files from upstream repos and then patching them to work
locally. The process of patching is a simple git merge which
replays changes from the last successful patch. For this to work,
the two steps (pulling files from upstream and patching) must
always be maintained in their own separate commits each time we
pull in the new upstream version.

Note that replaying changes from old commits using a simple
'git am' doesn't work well because a git patch created with
'git format-patch' doesn't have a knowledge of the common
ancestor of the master and the commit we want to replay,
so we end up with many unexpected conflicts.

Here are full steps for pulling code from upstream and patching:
```
> make thirdparty-pin
```
We must now commit upstream files first. This will keep the
subsequent patch in its own clean commit so we can use it in the
future.
```
> git add .
> git commit --signoff -m "Apply upstream"
```
We will refer to this commit later, let's say it is commit ```abcd```

Now we need to replay changes from the last correct patch.
We do it using git, with the help of a temporary branch where we
first copy the changes we want to replay. This example assumes
that the last correct patch was committed as ```5678```, and its
parent is the commit ```1234```.
```
> git format-patch --stdout 1234..5678 > ~/last.patch
> git checkout -b fix 1234
> git am ~/last.patch
> git checkout master
> git merge fix
```
If necessary, fix any conflicts and commit.

We can end up with two commits related to patching the upstream. To squash them
into a single commit, softly reset to ```abcd``` (the result of "Apply upstream",
see above) and commit all files to a single commit.
```
> git reset abcd
> git add .
> git commit --signoff -m "Patch upstream"
```
Note that ```git reset HEAD~2``` might not work as expected due to the merge.

Amend as required, and push all commits.

