set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

build:
    wails build -clean -tags webkit2_41

rpm:
    #!/usr/bin/env bash
    set -eu -o pipefail
    version="$(node -p 'JSON.parse(require("fs").readFileSync("frontend/package.json", "utf8")).version')"
    rpm_arch="$(uname -m)"
    command -v nfpm >/dev/null || { echo "nfpm not found in PATH" >&2; exit 1; }
    case "$rpm_arch" in
      x86_64) nfpm_arch=amd64 ;;
      aarch64) nfpm_arch=arm64 ;;
      *) echo "Unsupported RPM architecture: $rpm_arch" >&2; exit 1 ;;
    esac
    mkdir -p dist/rpm
    sed \
      -e "s/__VERSION__/$version/g" \
      -e "s/__ARCH__/$nfpm_arch/g" \
      packaging/nfpm.yaml.tmpl > dist/rpm/nfpm.yaml
    wails build -clean -tags webkit2_41
    nfpm package \
      --config dist/rpm/nfpm.yaml \
      --packager rpm \
      --target "dist/rpm/novella-${version}-1.${rpm_arch}.rpm"
