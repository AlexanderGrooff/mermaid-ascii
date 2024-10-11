#!/bin/bash

set -e
set -x

if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi
NEXT_TAG=$1
CURRENT_TAG=$(git describe --tags)

# Replace all occurrences of the current version with the next version
find . -type f -exec sed -i "s/${CURRENT_TAG}/${NEXT_TAG}/g" {} +
git add .
git commit -am "release version ${NEXT_TAG}"
git tag ${NEXT_TAG}

# Show changes of last commit
git show HEAD

# Ask for confirmation before pushing
read -p "Push to origin? [y/N] " -n 1 -r

git push
git push --tags
