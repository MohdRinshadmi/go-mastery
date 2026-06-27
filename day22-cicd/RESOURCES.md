# Day 22 Resources — CI/CD for Go

- **`actions/setup-go`**
  https://github.com/actions/setup-go
  Installs Go and caches the module + build cache (`cache: true`).

- **GitHub Actions — building & testing Go**
  https://docs.github.com/en/actions/use-cases-and-examples/building-and-testing/building-and-testing-go
  Official starter workflows for Go projects.

- **golangci-lint**
  https://golangci-lint.run/
  Config (`.golangci.yml`), the linter catalog, and the GitHub Action.

- **`golangci/golangci-lint-action`**
  https://github.com/golangci/golangci-lint-action
  The CI action, with caching and version pinning guidance.

- **`go vet` / cmd/vet**
  https://pkg.go.dev/cmd/vet
  The full list of analyzers `go vet` runs.

- **Go blog — data race detector**
  https://go.dev/blog/race-detector
  What `-race` does and how to use it.

- **`go test` flags (incl. `-shuffle`, `-count`, `-race`)**
  https://pkg.go.dev/cmd/go#hdr-Testing_flags
  Reference for the test flags that fight flakiness.

- **GitHub — about protected branches**
  https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches
  Required status checks and merge gates.

- **`docker/build-push-action`**
  https://github.com/docker/build-push-action
  Build and push images (with SHA tags) from a workflow.
