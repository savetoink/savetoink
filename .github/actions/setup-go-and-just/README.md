# Setup Go and Just

Reusable composite action that sets up both the Go toolchain and the Just command runner.

## Usage

Basic usage (uses stable Go version and latest Just):

```yaml
- uses: ./.github/actions/setup-go-and-just
```

With specific Go version:

```yaml
- uses: ./.github/actions/setup-go-and-just
  with:
    go-version: "1.21"
```

With Go modules caching:

```yaml
- uses: ./.github/actions/setup-go-and-just
  with:
    cache: true
```

## Parameters

### Go Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `go-version` | string | `stable` | Go version to use |
| `go-version-file` | string | - | Path to go.mod file to use for Go version |
| `cache` | boolean | - | Whether to cache Go modules |
| `cache-dependency-path` | string | - | Path to dependency files for caching |

### Just Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `just-version` | string | - | Just version to install (latest if not specified) |
| `just-bin-dir` | string | - | Directory to install Just binary |

## Updating versions

To update the Go setup version or Just install action version, modify the versions in `.github/actions/setup-go-and-just/action.yml`.
