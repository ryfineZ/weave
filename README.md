# Weave

A high-performance, visual proxy client for macOS.

Build proxy chains by dragging nodes on a canvas — no config files needed.

> **Status:** 🚧 v0.1.0 in development

## Tech Stack

- **Daemon** — Go + [sing-box](https://github.com/SagerNet/sing-box) as library
- **RPC** — [connect-go](https://connectrpc.com) + [buf](https://buf.build)
- **Desktop UI** — [Tauri 2](https://tauri.app) + React + [React Flow](https://reactflow.dev)

## Project Structure

```
cmd/proxyd/          Go daemon entry point
internal/            Daemon internals
proto/               Protobuf definitions (source of truth)
gen/                 Generated code (buf generate)
apps/desktop/        Tauri + React desktop app
build/               macOS packaging scripts
```

## Development

```bash
# Generate proto code
buf generate

# Run daemon
go run ./cmd/proxyd

# Run desktop app (requires daemon running)
npm run desktop
```

## License

GPL-3.0 (daemon links sing-box which is GPL-3.0)
