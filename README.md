GoMP3
=====

GoMP3 is a tiny YouTube-to-MP3 web app written in Go. It streams the converted MP3 back to the browser so the user downloads the audio immediately. The UI is built with gomponents, gomui, and htmx; the server uses Leapkit.

Features
- Paste a YouTube URL and download the MP3 in one step
- Renders a single, htmx-driven page with a responsive layout and dark mode toggle
- Streams the generated file directly (no temp storage beyond conversion)
- Ships with Tailwind-based styles compiled via `tailo` and Docker support with ffmpeg preinstalled

Requirements
- Go 1.24+
- ffmpeg available on your PATH (for local runs)
- Tailwind CLI helper `tailo` for rebuilding CSS in development
	- Install with `go tool tailo download`

Quickstart (local)
1) Install deps: `go mod download`
2) Ensure ffmpeg is installed (macOS: `brew install ffmpeg`).
3) Run the app: `go tool dev --watch.extensions=.go,.css,.js `
4) Visit http://localhost:3000 and paste a YouTube link.

Docker
- Build: `docker build -t gomp3 .`
- Run: `docker run --rm -p 3000:3000 gomp3`
- ffmpeg is included in the image.

Configuration
- `HOST` (default `0.0.0.0`)
- `PORT` (default `3000`)
- `SESSION_SECRET` (default random string)
- `SESSION_NAME` (default `leapkit_session`)


Notes
- The converter picks the best available audio stream from YouTube, runs ffmpeg, then streams the result to the client and removes the temp file.
