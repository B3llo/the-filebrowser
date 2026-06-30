<p align="center">
  <img src="https://raw.githubusercontent.com/filebrowser/filebrowser/master/branding/banner.png" width="550"/>
</p>

<p align="center">
  <a href="https://github.com/B3llo/the-filebrowser/actions/workflows/ci.yaml"><img src="https://github.com/B3llo/the-filebrowser/actions/workflows/ci.yaml/badge.svg" alt="Build"></a>
  <a href="https://goreportcard.com/report/github.com/filebrowser/filebrowser/v2"><img src="https://goreportcard.com/badge/github.com/filebrowser/filebrowser/v2" alt="Go Report Card"></a>
  <a href="https://github.com/B3llo/the-filebrowser/releases/latest"><img src="https://img.shields.io/github/release/B3llo/the-filebrowser.svg" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/B3llo/the-filebrowser.svg" alt="License"></a>
</p>

<p align="center">
  <strong>📁 A modern, beautiful web file manager for existing filesystems</strong><br>
  <em>Browse, manage, and share your files from anywhere.</em>
</p>

---

## What is FileBrowser?

FileBrowser gives you a clean, fast web interface to access your files on any server. Point it to a directory and you're ready to go, upload, download, preview, edit, and share files from any device with a browser.

It's not a cloud platform or a storage service. It's a **file manager for your existing filesystem**.

## Features

| Feature | Description |
|---|---|
| 🎨 **Modern UI** | Redesigned interface with light/dark themes, smooth animations, and a clean layout |
| 📂 **Multi-Source** | Browse multiple directories and mount points from a single interface |
| 🔍 **Command Palette** | Instant search with `Cmd+K` / `Ctrl+K`, find files and folders fast |
| 📤 **Upload & Share** | Drag-and-drop uploads, share links with expiration, and folder creation |
| 👁️ **Preview 20+ Formats** | Preview images, PDFs, videos, Markdown, code files, and more, without downloading |
| 🌍 **i18n** | Translated into multiple languages (Portuguese, English, and more via Transifex) |
| 🔐 **Auth Providers** | JSON, No-Auth, Proxy, and custom authentication providers |
| ⚡ **Single Binary** | Go backend + embedded Vue frontend, one binary, zero dependencies |
| 🐳 **Docker Ready** | Docker and Docker Compose support out of the box |

## Quick Start

### Docker (recommended)

```bash
docker pull b3llo/the-filebrowser
docker run -v /path/to/your/files:/srv -p 8080:80 b3llo/the-filebrowser
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.

### Binary

Download the latest release for your platform from [Releases](https://github.com/B3llo/the-filebrowser/releases), then:

```bash
./filebrowser -r /path/to/your/files
```

### Build from Source

```bash
git clone https://github.com/B3llo/the-filebrowser.git
cd the-filebrowser
task build
```

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | [Go](https://golang.org/) |
| Frontend | [Vue 3](https://vuejs.org/) + TypeScript + Pinia |
| Database | [Bolt](https://github.com/etcd-io/bbolt) (embedded) |
| Build | [Taskfile](https://taskfile.dev/) + [Vite](https://vitejs.dev/) |
| Styling | oklch design tokens, CSS custom properties |

## Documentation

Full documentation is available at [filebrowser.org](https://filebrowser.org).

## Contributing

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) before submitting a pull request.

## License

[Apache License 2.0](LICENSE) © File Browser Contributors
