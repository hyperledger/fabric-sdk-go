# Reusing fabric and fabric-ca Code

Fabric-sdk-go reuses fabric and fabric-ca code (upstream) by first copying desired files from upstream repos and then patching them to work locally. The first part, copying desired pieces of upstream (applying upstream), is captured in scripts of this package. To apply upstream (and overwrite any local changes), we simply run
```
> make thirdparty-pin
```

We then proceed by patching the upstream code to make it work locally, and committing the new version.

We don't worry about this until there is a need to modify the upstream code. "make thirdparty-pin" is not a part of "make all".

# Modifying Local Upstream Copy

If changes to local upstream copy don't require build changes (which would affect the execution of "make thirdparty-pin"), they are simply committed as any other change to fabric-sdk-go code.

# Upstream Code Changes

The procedure described here is required when what we pull from the upstream repositories is changed. The use cases are:
- upgrade upstream.
- change what parts of upstream we pull, e.g. when adding or dropping SDK features which require upstream functionality.
- downgrade upstream, e.g. if a simple rollback to some previous version is not feasible for any reason.

Steps:
1. Calculate a git patch which captures all changes to the upstream code we had to make locally in order to make it work.
2. Modify build to pull the desired upstream version. This might include changing the scripts in this package.
3. Run "make thirdparty-pin" to pull the new upstream version.
3. Apply the patch created in the first step, and any changes necessary for the new code to work locally.

The following sections describe each step in more detail.

## Calculate Upstream Patch

The objective of this step is to calculate a git patch which captures all changes to the upstream code we had to make locally to make it work.

First, we apply upstream. This will copy over upstream files and thus wipe out all changes we made locally to make them work.

```
> make thirdparty-pin
> git add .
> git commit --signoff -m "Apply upstream"
```
We will refer to this commit later, let's say it is commit ```abcd```

Next, we create a git commit which simply reverts changes from ```abcd```. This commit will captures all changes to the upstream code we made locally to make it work.
```
> git diff HEAD..HEAD~1 > ~/last.diff
> git apply ~/last.diff 
> git add .
> git commit --signoff -m "Patch upstream"
```
Verify that the patch commit is correct.
```
>git diff HEAD HEAD~2
```
The output should be empty.

Finally, we create a patch from the patch commit.
```
> git format-patch --stdout HEAD~1..HEAD > ~/upstream.patch
```

## Modify Build to Pull New Upstream Version(s)

In this step we modify the Makefile and make any required changes to the scripts in this package to pull the desired parts of upstream. All changes must be committed before proceeding.

## Pull the New Upstream Version
```
> make thirdparty-pin
> git add .
> git commit --signoff -m "Apply upstream"
```
## Fix the Code to Work Locally

The first step in fixing the upstream code to work locally is to replay all changes we made in the past for the same reason. These changes were previously calculated and captured in **~/upstream.patch** (see above).

Note that replaying changes from old commits using a simple 'git am' doesn't work well because a git patch created with 'git format-patch' doesn't have a knowledge of the common ancestor of the master and the commit we want to replay, so we end up with many unexpected conflicts.

We replay old changes using git, with the help of a temporary branch where we
first copy the changes we want to replay. Note that here we refer to commit which we labeled in this document as ```1234``` (see above). This commit is the parent of the commit we used to create the patch we will apply here.
```
> git checkout -b fix 1234
> git am ~/upstream.patch
> git checkout master
> git merge fix
```
If necessary, fix any conflicts and commit.

From here, we can proceed with any desired changes to any SDK file, including changes to any upstream file.
