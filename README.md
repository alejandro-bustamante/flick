# 🎞️ Flick

**Flick** is a Text User Interface (TUI) tool for automatically organizing movie and TV show files, designed for easy integration with media servers like **Plex** or **Jellyfin**.

It intelligently renames files, moves them to their final destination, and logs all changes made, allowing for auditing or reverting operations if necessary. It's a lightweight and portable alternative to tools like **FileBot**, written in Go.

## ✨ Features

- **Friendly TUI interface** built with [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Automatic file renaming** based on file names and metadata obtained using `ffprobe`
- **Queries external databases** such as The Movie Database (TMDb)
- **Moves files to their destination folder** with the following structure:
  - `Movies/NAME (YEAR)/NAME (YEAR).ext`
  - `TV Shows/NAME/Season X/NAME - SXXEXX.ext`
- **Supports multiple video formats**
- **Client/server architecture**: a daemon runs in the background, and the TUI communicates via Unix sockets
- **All binaries can be statically compiled** (including ffprobe and SQLite)
- **Operation history with rollback capability**

---

## 📦 Installation

_Coming soon._
`flick` will be distributed as a single binary for Linux.

---

## 🖥️ Usage

To launch the TUI:

```bash
flick
```
