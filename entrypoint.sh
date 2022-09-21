#!/bin/bash
set -e

echo "validate environment"

REQUIRED_VARS="GITHUB_TOKEN SSH_PUBLIC_KEY SSH_PRIVATE_KEY REMOTE_USER REMOTE_HOST REMOTE_PATH"

for VAR in $REQUIRED_VARS
do
  if [ -z "${!VAR}" ]; then
    echo "missing required environment variable: $VAR"
    exit 1
  fi
done

git config --global credential.helper store
echo "https://github-backup:$GITHUB_TOKEN@github.com" > "$HOME/.git-credentials"

BACKUP_DIR=/data/backups/github.com

echo "remove local backup copies"

rm -f $BACKUP_DIR/*.tar.gz

echo "start backup script"

python "$HOME/backup.py" \
       --concurrent 50 \
       --config-file "$HOME/config.json" \
       --backup-dir $BACKUP_DIR

echo "backup done"

cd $BACKUP_DIR
TIMESTAMP=$(date +"%Y-%m-%d")

for ORG in $(jq --raw-output '.orgs[].name' < "$HOME/config.json")
do
  echo "compressing $ORG"
  tar cfz "$ORG-$TIMESTAMP.tar.gz" "$ORG"
done

echo "compression done"

# Temporarily disable rsync

# echo "prepare SSH keys"

# mkdir -p "$HOME/.ssh"
# chmod 0700 "$HOME/.ssh"

# echo "$SSH_PRIVATE_KEY" > "$HOME/.ssh/id"
# echo "$SSH_PUBLIC_KEY" > "$HOME/.ssh/id.pub"
# chmod 600 "$HOME/.ssh/id" "$HOME/.ssh/id.pub"

# echo "syncing files"
# rsync ./*.tar.gz "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH"
# echo "sync done"
