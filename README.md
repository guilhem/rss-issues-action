# RSS issues action

This action create issues from a syndication feed (RSS or Atom).

## Inputs

### `repo-token`

**Required** the GITHUB_TOKEN secret.

### `feed`

**Required** URL of the rss.

### `prefix`

Prefix added to issues.

### `lastTime`

Limit items date.

### `labels`

Labels to add, comma separated.

### `dry-run`

Log issue creation but do nothing

## Outputs

### `issues`

Issues id, comma separated.

## Example

### step

```yaml
uses: guilhem/rss-issues-action
with:
  repo-token: ${{ secrets.GITHUB_TOKEN }}
  feed: "https://cloud.google.com/feeds/kubernetes-engine-release-notes.xml"
```

### complete

```yaml
name: rss

on:
  schedule:
    - cron: "0 * * * *"

jobs:
  gke-release:
    runs-on: ubuntu-latest
    steps:
      - uses: guilhem/rss-issues-action@0.0.1
        with:
          repo-token: ${{ secrets.GITHUB_TOKEN }}
          feed: "https://cloud.google.com/feeds/kubernetes-engine-release-notes.xml"
          prefix: "[GKE]"
          dry-run: "false"
          lastTime: "92h"
          labels: "liens/Kubernetes"
```

### Real Usage

- [Create information feed](https://github.com/p7t/actus/issues)
