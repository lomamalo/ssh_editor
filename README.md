# SSH Editor

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

Professional code editor accessible via browser for editing remote files via SSH/SFTP. Modern interface inspired by VSCode.

## ✨ Features

- **Modern and intuitive web interface**
- **Secure SSH/SFTP connection**
- **File explorer** with complete tree structure
- **Code editor** with multi-language support
- **Real-time saving** (Ctrl+S)
- **Sudo support** for editing system files
- **Create/delete** files and folders
- **Professional dark theme**
- **Lightweight and fast** - single binary, no dependencies

## Installation

### Prerequisites

- Go 1.21 or higher (to compile from sources)
- SSH access to a remote server

### Compile from sources
```bash
# Clone the repository
git clone https://github.com/your-username/ssh_editor.git
cd ssh_editor

# Install dependencies
go mod tidy
```

### Cross-platform compilation

#### Windows (PowerShell)
```powershell
# Depuis PowerShell
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o ssh-editor.exe
```

#### Linux/macOS (Bash)
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o ssh-editor
chmod +x ssh-editor

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o ssh-editor-macos
chmod +x ssh-editor-macos

# macOS ARM
GOOS=darwin GOARCH=arm64 go build -o ssh-editor-arm64
chmod +x ssh-editor-arm64
```

## 📖 Usage

### Startup

#### Windows
```cmd
ssh-editor.exe
```

#### Linux/macOS
```bash
# Make executable (first time only)
chmod +x ssh-editor-linux

# Launch
./ssh-editor-linux
```

### Access

1. Open your browser at **http://localhost:8080**
2. An SSH connection window appears automatically

### SSH Connection

Fill in the information:

- **Type**: Folder or Single file
- **Host**: Server IP (e.g., `192.168.1.100`)
- **Port**: SSH port (default: `22`)
- **Username**: SSH username
- **Password**: SSH password
- **Path**: Folder/file path (e.g., `/home/user/project`)
- **Use sudo**: Check if editing system files

### Keyboard shortcuts

- **Ctrl + S**: Save file
- **Tab**: Indentation (4 spaces)
- **Right click**: Context menu (delete)

### Advanced features

#### Create a file/folder
Click the `+` or `□` buttons in the explorer header.

#### Delete a file/folder
Right click on the item → Delete

#### Editing with sudo
Check "Use sudo" when connecting to edit system files requiring root privileges.

## Go Dependencies
```go
github.com/pkg/sftp
golang.org/x/crypto/ssh
```

## Security

**Important**:

- SSH passwords are stored temporarily in memory
- Use HTTPS in production (reverse proxy recommended)
- Limit access to port 8080 via firewall
- Do not expose directly on the Internet without additional authentication
