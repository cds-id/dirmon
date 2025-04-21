# DirMon - Directory Monitoring CLI Tool

[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

DirMon is a command-line application written in Go that monitors directories, lists their contents, and allows you to delete files. It provides both interactive and command-line interfaces with persistent configuration storage.

## Features

- **Directory Monitoring**: Watch directories for file creation, modification, deletion, and other changes
- **File Management**: List directory contents and delete files
- **Interactive Mode**: User-friendly menu-driven interface
- **Persistent Configuration**: Save directories to monitor for later use
- **Batch Monitoring**: Monitor all saved directories simultaneously

## Installation

### Prerequisites

- Go 1.18 or higher

### Building from Source

```bash
# Clone the repository
git clone git@github.com:cds-id/dirmon.git
cd dirmon

# Install dependencies
go mod tidy

# Build the application
go build -o dirmon
```

### Installation to System Path (Optional)

```bash
# For Linux/macOS
sudo cp dirmon /usr/local/bin/

# For Windows
# Move dirmon.exe to a directory in your PATH
```

## Usage

### Interactive Mode

The easiest way to use DirMon is through its interactive interface:

```bash
dirmon interactive
# or the shorter version
dirmon i
```

This will present a menu with the following options:
1. List directory contents
2. Delete a file
3. Monitor a directory
4. View monitored directories
5. Add directory to monitored list
6. Remove directory from monitored list
7. Monitor all saved directories

### Command Line Usage

You can also use DirMon directly from the command line:

```bash
# List directory contents (defaults to current directory)
dirmon list [path]
dirmon ls [path]

# Delete a file (with confirmation)
dirmon delete filename
dirmon rm filename

# Monitor a specific directory for changes
dirmon monitor [path]
dirmon mon [path]

# Add a directory to the monitored list
dirmon add-dir /path/to/directory

# Show all monitored directories
dirmon show-dirs

# Monitor all saved directories
dirmon monitor-all
```

## Configuration

DirMon stores its configuration in a JSON file. By default, it looks for configuration in the following locations:

1. `/opt/dirmon_config.json` (system-wide)
2. `~/.dirmon_config.json` (user's home directory)

The configuration file stores the list of directories to monitor, which can be managed through the interactive interface or with the `add-dir` command.

## Example Usage

### Adding Directories to Monitor

```bash
# Add important directories to monitor
dirmon add-dir /var/log
dirmon add-dir /home/user/projects
dirmon add-dir /opt/application/data
```

### Monitoring Multiple Directories

```bash
# Start monitoring all saved directories
dirmon monitor-all
```

Sample output:
```
Adding /var/log to watch list
Adding /home/user/projects to watch list
Adding /opt/application/data to watch list

Starting monitoring of all directories... (Press Ctrl+C to stop)
--------------------------------------------------------------------------------
[14:32:15] [/var/log] MODIFIED - syslog
[14:32:23] [/home/user/projects] CREATED - newfile.txt
[14:32:35] [/opt/application/data] DELETED - oldfile.dat
```

## License

MIT License - See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Troubleshooting

If you encounter any issues:

- Make sure you have sufficient permissions to access the directories you're monitoring
- Check that the Go version is 1.18 or higher
- For Linux users, you may need to run with sudo to monitor system directories
- On some systems, you might need to increase the inotify watches limit:
  ```bash
  echo fs.inotify.max_user_watches=524288 | sudo tee -a /etc/sysctl.conf && sudo sysctl -p
  ```
  ---

  Repository: [github.com/cds-id/dirmon](https://github.com/cds-id/dirmon)
