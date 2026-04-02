#!/usr/bin/env bash
set -euo pipefail

required_files=(
  "container-runtime/go.mod"
  "container-runtime/Makefile"
  ".github/workflows/ci.yml"
  ".github/workflows/release.yml"
)

for path in "${required_files[@]}"; do
  if [[ ! -f "$path" ]]; then
    echo "::error title=Missing required file::$path"
    exit 1
  fi
done

if grep -R --include='*.go' -q '"mrunc/pkg/specs"' container-runtime; then
  if ! find container-runtime -path '*/pkg/specs/*.go' -print -quit | grep -q .; then
    echo "::warning title=Potentially incomplete checkout::Go files import mrunc/pkg/specs, but no files were found under container-runtime/pkg/specs in this checkout. Builds will fail until that package exists."
  fi
fi

if [[ -f container-runtime/internal/runtime/namespace.go && -f container-runtime/internal/runtime/namespace_test.go ]]; then
  if grep -q 'func CreateNamespaces(config ' container-runtime/internal/runtime/namespace.go && \
     grep -q 'CreateNamespaces()' container-runtime/internal/runtime/namespace_test.go; then
    echo "::warning title=Potential signature mismatch::namespace_test.go calls CreateNamespaces() without the config argument expected by the implementation."
  fi
fi

echo "--- Preflight checks completed."
