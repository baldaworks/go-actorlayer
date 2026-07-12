# Releasing go-actorlayer

1. Ensure `main` is green and the README examples still pass under `go test ./...`.
2. Update `CHANGELOG.md` for the release.
3. Create and push a semantic version tag such as `v0.1.0`.
4. GitHub Actions will rerun tests and publish a GitHub release from the tag.

This module is a library, so release artifacts are source-based tags and release
notes rather than compiled binaries.
