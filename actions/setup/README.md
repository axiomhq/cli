# Setup Axiom CLI

Install the [Axiom CLI](https://github.com/axiomhq/cli) and add it to `PATH`.

```yaml
- uses: axiomhq/cli/actions/setup@v0.19.0
- run: axiom query "['my-dataset']"
  env:
    AXIOM_TOKEN: ${{ secrets.AXIOM_TOKEN }}
```

## Inputs

| Name | Default | Description |
|------|---------|-------------|
| `version` | `latest` | Version to install: a tag like `v0.17.0`, a bare `0.17.0`, or `latest`. |

## Outputs

| Name | Description |
|------|-------------|
| `version` | The installed version, as a tag, e.g. `v0.17.0`. |

## Details

The action downloads the release archive for the runner's platform, verifies it
against the release `checksums.txt`, and puts the binary on `PATH`. It does not
configure authentication. Set `AXIOM_TOKEN`, `AXIOM_ORG_ID`, and `AXIOM_URL` in
the environment of the steps that need them.

Linux, macOS, and Windows runners are supported on `amd64` and `arm64`. There is
no `windows/arm64` build. The archive is cached in the runner tool cache, which
persists on self-hosted runners but not on GitHub-hosted ones.
