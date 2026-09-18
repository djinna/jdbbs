# Checkpoints

Lightweight convention for marking known-good states in this repo.

## What a checkpoint is

A checkpoint is an annotated git tag that points to a commit we can safely return to.

Use checkpoints to:
- mark verified deploys/hotfixes
- set phase boundaries before risky work
- simplify rollback and comparison

## Tag naming convention

Format:

`checkpoint-YYYY-MM-DD-<scope>`

Examples:
- `checkpoint-2026-04-05-transmittal-ui`
- `checkpoint-2026-04-10-email-delivery`

## When to create one

Create a checkpoint:
1. after a change is verified in the target environment
2. before major refactors or migrations
3. at end of a milestone/phase

Avoid creating checkpoints for every tiny commit.

## Standard workflow

1) Make sure you are on the commit to mark:

```bash
git rev-parse --short HEAD
```

2) Create annotated tag (required):

```bash
git tag -a checkpoint-YYYY-MM-DD-<scope> -m "Checkpoint: <what was verified and where>"
```

3) Push tag:

```bash
git push origin checkpoint-YYYY-MM-DD-<scope>
```

## Useful commands

List checkpoints:

```bash
git tag --list "checkpoint-*"
```

Inspect a checkpoint:

```bash
git show checkpoint-YYYY-MM-DD-<scope>
```

Diff from checkpoint to current HEAD:

```bash
git diff checkpoint-YYYY-MM-DD-<scope>..HEAD
```

Create recovery branch from checkpoint:

```bash
git checkout -b recovery/<name> checkpoint-YYYY-MM-DD-<scope>
```

## Current example

- `checkpoint-2026-04-05-transmittal-ui`
  - verified transmittal UI cosmetics in prod
  - includes section header rename/order updates and removal of "Other Instructions"

- `checkpoint-2026-09-18-pre-workshop-v2` (supersedes `checkpoint-2026-09-18-pre-workshop` at `0cd67fd`)
  - end of the Fri 18 Sep punch-list block, ahead of the 21–22 Sep workshop
  - generated front-matter pages default on (5.14); `@`-in-link fix (5.16);
    colour images → tuned grey for print, EPUB keeps colour (5.15);
    compscore continuation-line fix; punch-list page polling fix
  - verified: `go test ./srv/` green, free Ghosts build 61 pp, three-image
    test book through the real pipeline, smoke on /factory, /admin/*,
    /pinstitute/*/factory + transmittal
