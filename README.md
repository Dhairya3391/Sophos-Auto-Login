# Sophos Auto Login (Go)

This repository now focuses on the Go implementation that automates login and logout for Sophos captive portals—commonly deployed on institutional networks. The binary maintains an authenticated session, periodically verifies connectivity, and gracefully shuts down when requested.

## Features

- Captive portal login/logout via HTTP form posts.
- Duration control: run forever or for a specified number of minutes/hours.
- Retry logic: checks connectivity every 30 seconds and re-authenticates when needed.
- Graceful shutdown on `q`/`Q`, `Ctrl+C`, or `SIGTERM`.
- Connection-pooled HTTP client with sane timeouts for long-running stability.

## Prerequisites

- Go 1.20+ (tested on macOS)

## Configuration

Edit the constants near the top of `main.go`:

```go
const (
    Username = "your_username"
    Password = "your_password"
)
```

For production, prefer environment variables, encrypted files, or a secret manager. The hard-coded constants are convenient for local builds only.

## Build

```bash
cd /Users/dhairya/Development/Sophos-auto-login
go build -o sophos-login main.go
```

## Run

```bash
./sophos-login [flags]
```

### Useful Flags

- `-forever` — Run until you quit manually.
- `-minutes <n>` — Run for `<n>` minutes.
- `-hours <n>` — Run for `<n>` hours.

If you skip all duration flags, the program defaults to 60 minutes. Use `./sophos-login -h` to see the full flag list.

### Stopping the Program

- Press `q` or `Q` (followed by Enter) to log out and exit.
- Press `Ctrl+C` to trigger cleanup and logout.

## Notes

- Keep credentials out of version control. `.gitignore` already excludes common secret files—extend it as needed.
- Test the binary on your target network before relying on it unattended.
- Use this utility only on networks where you have explicit permission.
