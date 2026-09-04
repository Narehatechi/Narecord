#!/usr/bin/env bash
# Copy every folder under Narecord plugins/ into Equicord src/userplugins/<Name>/.
# Used by .github/workflows/release.yml (build-desktop-asar).
# Preserve all files (index.ts/tsx, style.css, fallback.css, …).
set -euo pipefail

usage() {
  echo "usage: $0 <narecord-root> <equicord-root>" >&2
  exit 2
}

[[ $# -eq 2 ]] || usage

NARECORD_ROOT=$(cd "$1" && pwd)
EQUICORD_ROOT=$(cd "$2" && pwd)
SRC="$NARECORD_ROOT/plugins"
DEST="$EQUICORD_ROOT/src/userplugins"

if [[ ! -d "$SRC" ]]; then
  echo "missing plugins dir: $SRC" >&2
  exit 1
fi

mkdir -p "$DEST"

copied=()
for dir in "$SRC"/*/; do
  [[ -d "$dir" ]] || continue
  name=$(basename "$dir")
  target="$DEST/$name"
  mkdir -p "$target"
  cp -a "$dir." "$target/"
  copied+=("$name")
done

expected=(Abyss Hideout Incinerator NareMotion NarehateBadge Narelogs NareNotes Nnaa narePerf)
missing=()
for name in "${expected[@]}"; do
  if [[ ! -d "$DEST/$name" ]]; then
    missing+=("$name")
    continue
  fi
  if [[ ! -e "$DEST/$name/index.ts" && ! -e "$DEST/$name/index.tsx" ]]; then
    echo "missing index.ts/tsx for $name in $DEST/$name" >&2
    exit 1
  fi
done

if ((${#missing[@]})); then
  echo "missing expected userplugins: ${missing[*]}" >&2
  exit 1
fi

if ((${#copied[@]} == 0)); then
  echo "copied zero plugin folders from $SRC" >&2
  exit 1
fi

echo "Copied ${#copied[@]} plugin folder(s) into $DEST:"
printf '  %s\n' "${copied[@]}"
