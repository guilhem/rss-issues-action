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

### `aggregate`

Aggregate all items in a single issue

### `characterLimit`

Limit size of issue content

### `titleFilter`

Don't create an issue if the title matches the specified regular expression ([go regular expression syntax](https://github.com/google/re2/wiki/Syntax))

### `contentFilter`

Don't create an issue if the content matches the specified regular expression ([go regular expression syntax](https://github.com/google/re2/wiki/Syntax))

## Outputs

### `issues`

Issues number, comma separated.

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
      - uses: guilhem/rss-issues-action@0.2.0
        with:
          repo-token: ${{ secrets.GITHUB_TOKEN }}
          feed: "https://cloud.google.com/feeds/kubernetes-engine-release-notes.xml"
          prefix: "[GKE]"
          characterLimit: "255"
          dry-run: "false"
          lastTime: "92h"
          labels: "liens/Kubernetes"
```

### Real Usage

- [Create information feed](https://github.com/p7t/actus/issues)
