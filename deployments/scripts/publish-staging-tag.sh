#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: publish-staging-tag.sh [options]

Creates and pushes a lightweight git tag for staging in the format from the docs:
  stage-<short-sha>  → triggers deploy:stage in GitLab CI
  (see deployments/README.md, deployments/docs/README.md)

By default HEAD must be an ancestor of origin/develop (as in .gitlab-ci.yml).

Options:
  -n, --dry-run      only show the tag name and commands, without git tag / push
  -s, --skip-verify  do not verify that HEAD ⊂ develop (not recommended)
  -h, --help         this help
EOF
}

DRY_RUN=0
SKIP_VERIFY=0
while [ $# -gt 0 ]; do
  case "$1" in
    -n | --dry-run)
      DRY_RUN=1
      shift
      ;;
    -s | --skip-verify)
      SKIP_VERIFY=1
      shift
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

cd "$(git rev-parse --show-toplevel)"

if ! git diff-index --quiet HEAD -- 2>/dev/null; then
  echo "Warning: there are uncommitted changes in the index or working tree." >&2
  echo "The tag will still point only at the current HEAD commit." >&2
fi

git fetch --no-tags origin develop

if [ "$SKIP_VERIFY" -eq 0 ]; then
  if ! git merge-base --is-ancestor HEAD "refs/remotes/origin/develop"; then
    echo "Error: current commit (HEAD) is not on origin/develop." >&2
    echo "CI will reject the staging deploy. Switch to a commit from develop or merge your changes." >&2
    echo "Bypass: $0 --skip-verify (at your own risk)" >&2
    exit 1
  fi
fi

SHORT_SHA=$(git rev-parse --short HEAD)
TAG="stage-${SHORT_SHA}"

COMMIT=$(git rev-parse HEAD)
if git rev-parse "$TAG" >/dev/null 2>&1; then
  EXISTING=$(git rev-parse "$TAG^{commit}")
  if [ "$EXISTING" != "$COMMIT" ]; then
    echo "Error: tag $TAG already exists and points at another commit ($EXISTING)." >&2
    exit 1
  fi
  echo "Local tag $TAG is already on this commit."
else
  echo "Creating lightweight tag $TAG → $COMMIT"
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "DRY-RUN: git tag $TAG"
  else
    git tag "$TAG"
  fi
fi

echo "Pushing tag to origin..."
if [ "$DRY_RUN" -eq 1 ]; then
  echo "DRY-RUN: git push origin $TAG"
else
  git push origin "$TAG"
fi

echo "Done: $TAG (staging CI: deploy on rule ^stage-)"
