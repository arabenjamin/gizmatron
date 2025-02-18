#!/bin/sh

case "$(uname -m)" in
  x86_64)
    exec /gizmatron-amd64 "$@"  # Execute the amd64 binary
    ;;
  aarch64)
    exec /gizmatron-arm64 "$@"  # Execute the arm64 binary
    ;;
  *)
    echo "Unsupported architecture: $(uname -m)"
    exit 1
    ;;
esac