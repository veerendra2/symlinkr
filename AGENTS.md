# Repository Guidelines

## Recursive Removal Safety

- Remove recursive symlinks only when the current source tree contains the corresponding path and the destination symlink points to that exact source path.
- If a source file was deleted or moved, preserve its dangling destination symlink. This is intentional: Symlinkr is stateless and cannot safely confirm ownership without the source.
- Preserve unrelated symlinks and destination directories. Do not infer ownership by scanning the destination tree.
