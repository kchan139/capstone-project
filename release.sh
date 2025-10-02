#!/bin/bash
# Release script for mrunc
# Usage: ./release.sh <version>

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check arguments
if [ $# -ne 1 ]; then
    echo -e "${RED}Error: Version number required${NC}"
    echo "Usage: $0 <version>"
    echo "Example: $0 0.6.9"
    exit 1
fi

VERSION=$1
TAG="v${VERSION}"

# Validate version format
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Invalid version format${NC}"
    echo "Version must be in format: MAJOR.MINOR.PATCH)"
    exit 1
fi

echo -e "${GREEN}Creating release ${TAG}${NC}"

# Check if we're in the project root
if [ ! -f "container-runtime/go.mod" ]; then
    echo -e "${RED}Error: Must run from project root${NC}"
    exit 1
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo -e "${YELLOW}Warning: You have uncommitted changes${NC}"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if tag already exists
if git rev-parse "$TAG" >/dev/null 2>&1; then
    echo -e "${RED}Error: Tag ${TAG} already exists${NC}"
    exit 1
fi

# Show what will be released
echo -e "${YELLOW}Changes since last release:${NC}"
LAST_TAG=$(git tag --sort=-creatordate | grep '^v[0-9]' | head -n1)
if [ -n "$LAST_TAG" ]; then
    git log --oneline "${LAST_TAG}..HEAD"
else
    echo "First release - showing all commits:"
    git log --oneline
fi

echo
read -p "Create tag ${TAG} and push? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
fi

# Create annotated tag
echo -e "${GREEN}Creating tag ${TAG}...${NC}"
git tag -a "$TAG" -m "Release ${TAG}"

# Push tag
echo -e "${GREEN}Pushing tag to GitHub...${NC}"
git push origin "$TAG"

echo
echo -e "${GREEN}✓ Release ${TAG} created!${NC}"
echo
echo "GitHub Actions will now:"
echo "  1. Build the binary with version ${VERSION}"
echo "  2. Create GitHub release"
echo "  3. Upload release artifacts"
echo
echo "Monitor the release at:"
echo "  https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[\/:]\(.*\)\.git/\1/')/actions"
echo
echo "To undo (if needed):"
echo "  git tag -d ${TAG}"
echo "  git push origin :refs/tags/${TAG}"
