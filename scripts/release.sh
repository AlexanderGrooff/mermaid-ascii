#!/bin/bash

set -e
set -x

if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi
NEXT_TAG=$1
CURRENT_TAG=$(git tag | tail -n1)

# Replace all occurrences of the current version with the next version
rg -F -l "${CURRENT_TAG}" | grep -v 'go\.mod\|go\.sum' | xargs sed -i "s/${CURRENT_TAG}/${NEXT_TAG}/g"
git add .
git commit -am "release version ${NEXT_TAG}"

# Show changes of last commit
git show HEAD

# Ask for confirmation before pushing
read -p "Push to origin? [y/N] " -n 1 -r

git tag ${NEXT_TAG}
git push
git push --tags
