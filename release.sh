#!/bin/bash

set -e

git fetch --tags

LATEST_TAG=$(git describe --tags --abbrev=0)
echo "Latest tag: $LATEST_TAG"

VERSION_PARTS=($(echo $LATEST_TAG | sed 's/v//g' | tr '.' ' '))

PATCH_VERSION=${VERSION_PARTS[2]}
NEW_PATCH_VERSION=$((PATCH_VERSION + 1))

NEW_VERSION="v${VERSION_PARTS[0]}.${VERSION_PARTS[1]}.$NEW_PATCH_VERSION"
echo "New version: $NEW_VERSION"

git add .
git commit -m "Release $NEW_VERSION"

git tag $NEW_VERSION


git push origin main
git push origin $NEW_VERSION

echo "Release $NEW_VERSION pushed successfully!"
