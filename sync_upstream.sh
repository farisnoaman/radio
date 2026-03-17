#!/bin/bash

# sync_upstream.sh
# A script to sync your fork with the official ToughRadius repository

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Checking repo status...${NC}"

# Ensure we are in the main branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo -e "${RED}Error: You are on branch '$CURRENT_BRANCH'. Please switch to 'main' before syncing.${NC}"
    exit 1
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo -e "${RED}Error: You have uncommitted changes. Please commit or stash them first.${NC}"
    exit 1
fi

# Check if upstream exists
if ! git remote | grep -q "upstream"; then
    echo -e "${YELLOW}Adding upstream remote...${NC}"
    git remote add upstream https://github.com/talkincode/toughradius.git
fi

echo -e "${GREEN}Fetching latest changes from official repo (upstream)...${NC}"
git fetch upstream

echo -e "${GREEN}Merging updates into your local main branch...${NC}"
if git merge upstream/main; then
    echo -e "${GREEN}Successfully synced!${NC}"
    echo -e "${YELLOW}You can now push these updates to your GitHub fork using: git push fork main${NC}"
else
    echo -e "${RED}Conflict detected!${NC}"
    echo -e "Please resolve the conflicts manually, then commit the changes."
fi
