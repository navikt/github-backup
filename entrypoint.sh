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

BACKUP_DIR=/tmp/backups/github.com

echo "remove local copies"
rm -f $BACKUP_DIR/*.tar.gz

echo "prepare keys"

echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id
echo "${SSH_PUBLIC_KEY}" > ~/.ssh/id.pub
chmod 600 ~/.ssh/id ~/.ssh/id.pub
echo "IdentityFile ~/.ssh/id" > ~/.ssh/config
ssh-keyscan github.com > ~/.ssh/known_hosts 2>&1
ssh-keyscan $REMOTE_HOST >> ~/.ssh/known_hosts 2>&1

echo "start backup script"

python /home/backup/backup.py \
       --concurrent 50 \
       --config-file /home/backup/config.json \
       --backup-dir $BACKUP_DIR

echo "backup done"

cd $BACKUP_DIR
TIMESTAMP=$(date +"%Y-%m-%d")

for ORG in $(cat /home/backup/config.json | jq --raw-output '.orgs[].name')
do
  echo "compressing $ORG"
  tar cfz $ORG-$TIMESTAMP.tar.gz $ORG
done

echo "syncing files"
rsync *.tar.gz $REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH

echo "done"
