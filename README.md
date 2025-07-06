# 🏔️ PEAKS - Beautiful Terminal Bandwidth Monitor

A modern, real-time bandwidth monitoring tool for your terminal, built with the latest Charm ecosystem tools for a stunning user experience.

## ✨ Features

- **Real-time Monitoring**: Monitor network bandwidth with high-resolution split-axis braille charts
- **Beautiful TUI**: Built with Bubble Tea and Lip Gloss for a modern terminal interface
- **Cross-platform**: Works on Linux, macOS, and Windows
- **Split-Axis Charts**: Clear separation with upload below and download above the axis line
- **Braille Charts**: High-resolution area charts using Unicode braille characters
- **Color Coding**: 
  - 🔴 Red for upload traffic (below axis)
  - 🟢 Green for download traffic (above axis)
- **Interactive Controls**: Pause, reset, toggle stats, and more
- **Detailed Statistics**: Track uptime, peaks, and totals
- **Responsive Design**: Adapts to terminal size automatically
- **1-Minute History**: Shows 60 seconds of bandwidth data at 500ms intervals

## 🚀 Installation

### Prerequisites
- Go 1.21 or higher
- A terminal with Unicode and color support

### Build from Source
```bash
git clone https://github.com/marco/peaks
cd peaks
go build -o peaks
./peaks
```

## 🎮 Controls

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `p` / `Space` | Pause/Resume monitoring |
| `r` | Reset chart and statistics |
| `s` | Toggle statistics panel |
| `?` | Toggle help |

## 🖥️ Screenshots

The tool displays:
- Real-time split-axis braille chart with upload below and download above the axis
- Current bandwidth rates in the footer
- Optional statistics panel with uptime, peaks, and totals
- Live/paused status indicator
- Beautiful color-coded interface with clear traffic separation

## 🛠️ Technical Details

### Built With
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - The Elm Architecture for Go TUI apps
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** - Style definitions for terminal layouts
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - Common UI components
- **[gopsutil](https://github.com/shirou/gopsutil)** - Cross-platform system information

### Architecture
- `main.go` - Main application and Bubble Tea model
- `bandwidth.go` - Cross-platform bandwidth monitoring
- `chart.go` - Braille chart rendering with color overlays
- `ui.go` - Enhanced UI components and statistics

### Chart Rendering
The tool uses Unicode braille characters (U+2800–U+28FF) for high-resolution terminal charts. Each character provides 8 dots in a 2×4 matrix, allowing for detailed visualization of bandwidth patterns.

### Color Mixing
When upload and download traffic overlap in the same chart position, the visualization intelligently blends colors to show yellow, indicating simultaneous activity.

## 🎨 Customization

The tool uses modern terminal colors and should work well with most terminal themes. Colors are automatically adapted based on your terminal's color support.

## 📊 Performance

- Updates at 2 FPS for smooth, easy-to-follow visualization
- Minimal CPU usage through efficient rendering
- Maintains 1 minute of history by default
- Automatic scaling based on observed peak values

## 🐛 Troubleshooting

### Braille Characters Not Displaying
- Ensure your terminal font supports Unicode braille characters
- Try fonts like: Cascadia Code, Fira Code, or DejaVu Sans Mono
- On Windows, use Windows Terminal or a modern terminal emulator

### Colors Not Showing
- Verify your terminal supports ANSI colors
- Modern terminals (Terminal.app, iTerm2, Windows Terminal) should work fine
- Legacy terminals may show limited colors

### Permission Issues
- The tool only reads network interface statistics
- No special permissions required on most systems
- If issues persist, try running as administrator/sudo

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- The amazing [Charm](https://charm.sh) team for the beautiful TUI libraries
- The Go community for excellent cross-platform system libraries
- Terminal art enthusiasts who pioneered braille-based visualization

---

*Made with ❤️ and lots of ☕ by marco*
