# RSS issues action

Create GitHub issues from an RSS or Atom feed. Existing issue titles are skipped,
including closed issues in the returned issue page. Outputs use `GITHUB_OUTPUT`.

The action builds its Dockerfile from the selected Git reference, so using a
commit SHA executes that commit's code. It needs a Linux runner with Docker.
Pin a reviewed commit SHA in production; `@master` below follows the default branch.

## Inputs

### `repo-token`

**Required** a GitHub token with `issues: write` permission on the target repository.

### `feed`

**Required** URL of the rss.

### `prefix`

Prefix added to issues.

### `lastTime`

Only include items published within this duration, such as `92h`. Defaults to `720h` (30 days); items without a publish date are included.

### `labels`

Labels to add, comma separated.

### `dry-run`

Log proposed issues without creating them. GitHub issue reads and feed downloads still occur.

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
uses: guilhem/rss-issues-action@master
with:
  repo-token: ${{ secrets.GITHUB_TOKEN }}
  feed: "https://cloud.google.com/feeds/kubernetes-engine-release-notes.xml"
```

### complete

```yaml
name: rss

on:
  workflow_dispatch:
  schedule:
    - cron: "0 * * * *"

permissions:
  contents: read
  issues: write

jobs:
  gke-release:
    runs-on: ubuntu-latest
    steps:
      - name: rss-issues
        id: rss
        uses: guilhem/rss-issues-action@master
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

## Development and publishing

Use Go 1.27.1 or newer:

```sh
go test -v ./...
go vet ./...
go build .
docker build -t rss-issues-action:local .
```

The tests use a local feed and mocked GitHub API to check duplicate detection,
individual issues, aggregation, dry-run, and environment-file output. No real
issues are created.

Pull requests build the image without logging into or publishing to GHCR. Pushes
to `master` and tags publish `ghcr.io/guilhem/rss-issues-action` using the job's
`packages: write` permission. Consumers of `action.yml` build the Dockerfile;
they do not depend on those published image tags. The runtime stays root so it
can write GitHub's mounted output files.

Duplicate detection currently checks the first page of matching issue titles.
Aggregation changes the title to include the run time, so it does not provide
per-item deduplication across aggregate runs.
