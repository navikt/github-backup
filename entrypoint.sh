#!/bin/sh
set -e

if [[ -z "${GITHUB_TOKEN}" ]]; then
  echo "GITHUB_TOKEN is required"
  exit 1
fi

if [[ -z "${SSH_PUBLIC_KEY}" ]]; then
  echo "SSH_PUBLIC_KEY is required"
  exit 1
fi

if [[ -z "${SSH_PRIVATE_KEY}" ]]; then
  echo "SSH_PRIVATE_KEY is required"
  exit 1
fi

echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id
echo "${SSH_PUBLIC_KEY}" > ~/.ssh/id.pub
chmod 600 ~/.ssh/id ~/.ssh/id.pub

echo "IdentityFile ~/.ssh/id" > ~/.ssh/config

ssh-keyscan github.com > ~/.ssh/known_hosts 2>&1

python /home/backup/backup.py \
       --concurrent 50 \
       --config-file /home/backup/config.json \
       --backup-dir /tmp/backups/github.com
