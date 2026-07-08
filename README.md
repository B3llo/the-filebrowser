<h1 align="center">📁 The File Browser</h1>

<p align="center">
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

## Acknowledgements

This project is based on [filebrowser/filebrowser](https://github.com/filebrowser/filebrowser). Huge thanks to the original File Browser Contributors for the foundational work.

## License

[Apache License 2.0](LICENSE) © File Browser Contributors
