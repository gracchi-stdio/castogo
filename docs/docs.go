// Package docs embeds the markdown documentation tree for runtime serving.
//
// The embed directive must live in this file (inside docs/) because go:embed can
// only reach the declaring package's own subtree. The dev/ and user/ directories
// must each contain at least one file at build time, or compilation fails with
// "pattern dev: no matching files found".
package docs

import "embed"

// FS is the embedded documentation filesystem. Top-level subdirectories are
// "dev" (developer docs) and "user" (user docs). Read entries with
// fs.ReadDir(FS, "dev") or a single file with FS.ReadFile("dev/getting-started.md").
//
//go:embed dev user
var FS embed.FS
