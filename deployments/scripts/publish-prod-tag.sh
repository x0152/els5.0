#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: publish-prod-tag.sh [options] [version]

Creates and pushes a lightweight git tag for production:
  v<semver>  -> triggers deploy:prod in GitLab CI
  Examples: v1.2.3, 1.2.3, v1.2.3-rc.1
  (tag format is described in deployments/docs/ci-cd-gitflow.md)

Without a version argument the script runs interactively: shows the latest
version and offers a patch / minor / major bump or manual input.

By default HEAD must be an ancestor of origin/main (as in .gitlab-ci.yml).

Options:
  -n, --dry-run            only show the tag name and commands, without git tag / push
  -s, --skip-verify        do not verify that HEAD ⊂ main (not recommended)
  -v, --version <version>  release version (v prefix optional), disables the dialog
  -h, --help               this help
EOF
}

DRY_RUN=0
SKIP_VERIFY=0
VERSION_INPUT=""

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
    -v | --version)
      if [ $# -lt 2 ]; then
        echo "Error: $1 requires a version value." >&2
        usage >&2
        exit 2
      fi
      VERSION_INPUT="$2"
      shift 2
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    -*)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
    *)
      if [ -n "$VERSION_INPUT" ]; then
        echo "Extra positional argument: $1" >&2
        usage >&2
        exit 2
      fi
      VERSION_INPUT="$1"
      shift
      ;;
  esac
done

# Full semver (with optional pre-release / build), as in ci-cd-gitflow.md.
SEMVER_RE='^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z][0-9A-Za-z.-]*)?(\+[0-9A-Za-z.-]+)?$'

normalize_version() {
  # Prints a normalized semver (without v prefix) or exits with an error.
  local raw="${1#v}"
  if [[ ! "$raw" =~ $SEMVER_RE ]]; then
    echo "Error: version '$1' does not look like semver." >&2
    echo "Expected a format like: 1.2.3, v1.2.3, 1.2.3-rc.1" >&2
    exit 2
  fi
  printf '%s' "$raw"
}

cd "$(git rev-parse --show-toplevel)"

# Fetch main and tags to correctly compute the latest version and branch check.
git fetch --tags --force origin main >/dev/null 2>&1 || git fetch --tags origin

# Latest stable v-tag by semver sort (pre-releases are excluded).
LATEST_TAG="$(git tag -l 'v[0-9]*' --sort=-v:refname \
  | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' \
  | head -n1 || true)"

if [ -n "$LATEST_TAG" ]; then
  BASE_CORE="${LATEST_TAG#v}"
else
  BASE_CORE="0.0.0"
fi
IFS='.' read -r BASE_MAJ BASE_MIN BASE_PAT <<<"$BASE_CORE"

NEXT_PATCH="${BASE_MAJ}.${BASE_MIN}.$((BASE_PAT + 1))"
NEXT_MINOR="${BASE_MAJ}.$((BASE_MIN + 1)).0"
NEXT_MAJOR="$((BASE_MAJ + 1)).0.0"

if [ -z "$VERSION_INPUT" ]; then
  if [ ! -t 0 ]; then
    echo "Error: no version and stdin is not interactive." >&2
    echo "Pass the version as an argument, e.g.: $0 1.2.3" >&2
    exit 2
  fi

  echo "Latest production version: ${LATEST_TAG:-<no v* tags>}"
  echo
  echo "What are we releasing?"
  echo "  1) patch -> v${NEXT_PATCH}   (bugfixes)"
  echo "  2) minor -> v${NEXT_MINOR}   (new features, backward compatible)"
  echo "  3) major -> v${NEXT_MAJOR}   (breaking changes)"
  echo "  4) enter version manually"
  echo

  CHOICE=""
  while [ -z "$CHOICE" ]; do
    printf "Choice [1/2/3/4]: "
    read -r CHOICE
    case "$CHOICE" in
      1) VERSION_INPUT="$NEXT_PATCH" ;;
      2) VERSION_INPUT="$NEXT_MINOR" ;;
      3) VERSION_INPUT="$NEXT_MAJOR" ;;
      4)
        printf "Enter version (e.g. 1.2.3): "
        read -r VERSION_INPUT
        ;;
      *)
        echo "Enter 1, 2, 3, or 4." >&2
        CHOICE=""
        ;;
    esac
  done
fi

VERSION="$(normalize_version "$VERSION_INPUT")"
TAG="v${VERSION}"

if [ -t 0 ]; then
  printf "Release %s? [y/N]: " "$TAG"
  read -r CONFIRM
  case "$CONFIRM" in
    y | Y | yes | YES) ;;
    *)
      echo "Cancelled."
      exit 0
      ;;
  esac
fi

if ! git diff-index --quiet HEAD -- 2>/dev/null; then
  echo "Warning: there are uncommitted changes in the index or working tree." >&2
  echo "The tag will still point only at the current HEAD commit." >&2
fi

if [ "$SKIP_VERIFY" -eq 0 ]; then
  if ! git merge-base --is-ancestor HEAD "refs/remotes/origin/main"; then
    echo "Error: current commit (HEAD) is not on origin/main." >&2
    echo "CI will reject deploy:prod. Switch to a commit from main or merge your changes." >&2
    echo "Bypass: $0 --skip-verify $VERSION (at your own risk)" >&2
    exit 1
  fi
fi

COMMIT=$(git rev-parse HEAD)
if git rev-parse "$TAG" >/dev/null 2>&1; then
  EXISTING=$(git rev-parse "$TAG^{commit}")
  if [ "$EXISTING" != "$COMMIT" ]; then
    echo "Error: tag $TAG already exists and points at another commit ($EXISTING)." >&2
    exit 1
  fi
  echo "Local tag $TAG is already on this commit."
else
  echo "Creating lightweight tag $TAG -> $COMMIT"
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

echo "Done: $TAG (production CI: deploy on rule ^v\\d)"
