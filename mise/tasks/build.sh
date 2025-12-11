#!/usr/bin/env bash
#MISE description="Build go code"
set -euo pipefail

go build -o bin/github-backup .
