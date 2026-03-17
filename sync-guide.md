# Maintaining Your Fork (Syncing with Upstream)

Since your `kart` repository is a fork of [talkincode/toughradius](https://github.com/talkincode/toughradius), you should periodically "sync" to get new features and security fixes.

## Automated Update Script
I have created a script called `sync_upstream.sh` in your project root. To use it:

1. **Run the script**:
   ```bash
   ./sync_upstream.sh
   ```

## What the script does:
1. **Fetches** the latest code from the official ToughRadius (`upstream`).
2. **Merges** those changes into your local [main](file:///home/faris/Downloads/toughradius/toughradius/main.go#57-123) branch.
3. **Identifies Conflicts**: If the official project changed a file that we also changed (like the Arabic UI or Billing fixes), Git will ask you to resolve it.

## Best Practice for Customizations
To make updates easier in the future:
- **Small Fixes**: It's okay to keep them in [main](file:///home/faris/Downloads/toughradius/toughradius/main.go#57-123).
- **Large New Features**: Create a new branch (e.g., `git checkout -b feat/my-new-feature`). This keeps your [main](file:///home/faris/Downloads/toughradius/toughradius/main.go#57-123) branch clean and easier to sync.

## Useful Commands
- `git remote -v`: Check your connections.
- `git log upstream/main..main`: See what custom changes you have that are NOT in the official version.
- `git log main..upstream/main`: See what new official changes you are missing.