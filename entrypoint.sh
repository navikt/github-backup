#!/bin/sh
set -e

echo "validate environment"

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

if [[ -z "${REMOTE_USER}" ]]; then
  echo "REMOTE_USER is required"
  exit 1
fi

if [[ -z "${REMOTE_HOST}" ]]; then
  echo "REMOTE_HOST is required"
  exit 1
fi

if [[ -z "${REMOTE_PATH}" ]]; then
  echo "REMOTE_PATH is required"
  exit 1
fi

BACKUP_DIR=/data/backups/github.com
TMP_HOME=/tmp/home/backup

echo "remove local copies"

rm -f $BACKUP_DIR/*.tar.gz

echo "prepare keys"

mkdir -p $TMP_HOME/.ssh
chmod 0700 $TMP_HOME/.ssh

echo "create private and public keys"

echo "$SSH_PRIVATE_KEY" > $TMP_HOME/.ssh/id
echo "$SSH_PUBLIC_KEY" > $TMP_HOME/.ssh/id.pub
chmod 600 $TMP_HOME/.ssh/id $TMP_HOME/.ssh/id.pub

echo "gather SSH public keys"

ssh-keyscan github.com > $TMP_HOME/.ssh/known_hosts
# ssh-keyscan "$REMOTE_HOST" >> $TMP_HOME/.ssh/known_hosts

echo "start backup script"

python ~/backup.py \
       --concurrent 50 \
       --config-file ~/config.json \
       --backup-dir $BACKUP_DIR

echo "backup done"

cd $BACKUP_DIR
TIMESTAMP=$(date +"%Y-%m-%d")

for ORG in $(jq --raw-output '.orgs[].name' < ~/config.json)
do
  echo "compressing $ORG"
  tar cfz "$ORG-$TIMESTAMP".tar.gz "$ORG"
done

exit 0

echo "syncing files"
rsync ./*.tar.gz "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH"

echo "sync done"
