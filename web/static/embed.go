package static

import "embed"

// Embed all static assets needed by the site (CSS, favicons, manifest)
// Keep the pattern explicit to avoid accidentally pulling large/unneeded files later.
//
//go:embed styles.css favicon.ico favicon-16x16.png favicon-32x32.png apple-touch-icon.png android-chrome-192x192.png android-chrome-512x512.png site.webmanifest
var Files embed.FS
