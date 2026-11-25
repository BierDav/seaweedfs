# The SeaweedFS backup fork

`patched` is the newest SeaweedFS release plus the patches listed in
`.fork-kit.toml`, both for `weed filer.backup` writing to a local mirror.

Every push to `patched` tests the patches, builds `weed` and publishes
`ghcr.io/bierdav/seaweedfs-backup:<release>-custom.<n>` on the upstream runtime
image `chrislusf/seaweedfs:<release>` (`custom-release.yml`). `<release>` is the
newest SeaweedFS release merged into `patched`. The cluster pins the image by
digest in selfhosted-platform (`manifests/backup/seaweedfs-backup.yaml`).

## Keeping up with upstream

[fork-kit](https://github.com/BierDav/fork-kit) (`.github/workflows/fork-kit.yml`)
merges the newest SeaweedFS release every six hours into the branch
`fork-kit/sync` and keeps one pull request into `patched` open for it.

- **Clean merge, green checks, no upstream `.github/` changes:** the PR merges
  itself, and the push to `patched` publishes a new image.
- **Conflicts:** the PR is labelled `fork-kit:conflict` and lists the
  conflicting files with the patches they belong to. Resolve them on
  `fork-kit/sync` (`npx github:BierDav/fork-kit#v0.3.1 resolve`, or GitHub's
  conflict editor) and merge the PR yourself.
- **Upstream changed `.github/`:** nothing is pushed until a human has reviewed
  the change. Run the fork-kit workflow with `allow-workflow-changes` once you
  have.
- **Holding a PR:** add the label `fork-kit:hold` to any PR into `patched`.
  The required check `fork-kit/hold` then fails and auto-merge goes off, so the
  PR cannot merge until the label is removed; the next fork-kit run turns
  auto-merge back on for a sync PR. It holds that PR only, not releases.

## Repository settings this relies on

- `patched` is the default branch, so the schedule runs.
- **Upstream's workflows are disabled** in this fork; only `custom-release.yml`
  and `fork-kit.yml` run. Disable any workflow upstream adds when you review its
  `.github/` change.
- **The GitHub App:** its key is the secret `FORK_KIT_APP_KEY` in the
  environment `fork-kit`, which only `patched` can deploy to; its ID is the
  repository variable `FORK_KIT_APP_ID`.
- **A ruleset on `patched`:** a pull request and the checks `test`, `build` and `fork-kit/hold`;
  no force pushes or deletion. Repository admins may bypass it.
- **Auto-merge** is allowed in the repository settings.
