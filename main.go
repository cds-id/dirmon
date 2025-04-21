package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli/v2"
)

// Config stores the application configuration
type Config struct {
	MonitoredDirs []string `json:"monitored_dirs"`
}

// Global variables
var (
	configFile = "dirmon_config.json"
	appConfig  Config
)

func main() {
	// Load configuration
	loadConfig()

	app := &cli.App{
		Name:  "dirmon",
		Usage: "Monitor directories and manage files",
		Commands: []*cli.Command{
			{
				Name:    "interactive",
				Aliases: []string{"i"},
				Usage:   "Run in interactive mode",
				Action: func(c *cli.Context) error {
					return runInteractiveMode()
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List directory contents",
				Action: func(c *cli.Context) error {
					path := "."
					if c.NArg() > 0 {
						path = c.Args().Get(0)
					}
					return listDirectory(path)
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"rm"},
				Usage:   "Delete a file",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("please specify a file to delete")
					}
					return deleteFile(c.Args().Get(0))
				},
			},
			{
				Name:    "monitor",
				Aliases: []string{"mon"},
				Usage:   "Monitor a directory for changes",
				Action: func(c *cli.Context) error {
					path := "."
					if c.NArg() > 0 {
						path = c.Args().Get(0)
					}
					return monitorDirectory(path)
				},
			},
			{
				Name:  "add-dir",
				Usage: "Add a directory to monitored list",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("please specify a directory to add")
					}
					return addDirectory(c.Args().Get(0))
				},
			},
			{
				Name:  "show-dirs",
				Usage: "Show all monitored directories",
				Action: func(c *cli.Context) error {
					viewMonitoredDirectories()
					return nil
				},
			},
			{
				Name:  "monitor-all",
				Usage: "Monitor all saved directories",
				Action: func(c *cli.Context) error {
					return monitorAllDirectories()
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// loadConfig loads the application configuration
func loadConfig() {
	// Look for config in /opt directory first
	optConfigPath := "/opt/dirmon_config.json"

	// Check if file exists in /opt
	if _, err := os.Stat(optConfigPath); err == nil {
		configFile = optConfigPath
	} else {
		// Fall back to user's home directory
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configFile = filepath.Join(homeDir, ".dirmon_config.json")
		}
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		// Config file doesn't exist yet - create empty config
		appConfig = Config{
			MonitoredDirs: []string{},
		}
		return
	}

	if err := json.Unmarshal(data, &appConfig); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		appConfig = Config{
			MonitoredDirs: []string{},
		}
		return
	}
}

// saveConfig saves the application configuration
func saveConfig() error {
	data, err := json.MarshalIndent(appConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

// clearScreen clears the terminal screen
func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// runInteractiveMode starts the interactive CLI mode
func runInteractiveMode() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		clearScreen()
		fmt.Println("===== Directory Monitor =====")
		fmt.Println("1. List directory contents")
		fmt.Println("2. Delete a file")
		fmt.Println("3. Monitor a directory")
		fmt.Println("4. View monitored directories")
		fmt.Println("5. Add directory to monitored list")
		fmt.Println("6. Remove directory from monitored list")
		fmt.Println("7. Monitor all saved directories")
		fmt.Println("0. Exit")
		fmt.Println("=============================")
		fmt.Print("Enter your choice: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		clearScreen()

		switch choice {
		case "0":
			return nil
		case "1":
			fmt.Print("Enter directory path (or press Enter for current directory): ")
			path, _ := reader.ReadString('\n')
			path = strings.TrimSpace(path)
			if path == "" {
				path = "."
			}

			err := listDirectory(path)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')

		case "2":
			fmt.Print("Enter file path to delete: ")
			path, _ := reader.ReadString('\n')
			path = strings.TrimSpace(path)

			if path != "" {
				err := deleteFile(path)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			}

			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')

		case "3":
			fmt.Print("Enter directory path to monitor (or press Enter for current directory): ")
			path, _ := reader.ReadString('\n')
			path = strings.TrimSpace(path)
			if path == "" {
				path = "."
			}

			fmt.Println("Monitoring directory. Press Ctrl+C to stop...")
			err := monitorDirectory(path)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "4":
			viewMonitoredDirectories()
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')

		case "5":
			fmt.Print("Enter directory path to add: ")
			path, _ := reader.ReadString('\n')
			path = strings.TrimSpace(path)

			if path != "" {
				err := addDirectory(path)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			}

			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')

		case "6":
			err := removeDirectory()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')

		case "7":
			fmt.Println("Monitoring all directories. Press Ctrl+C to stop...")
			err := monitorAllDirectories()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		default:
			fmt.Println("Invalid choice")
			time.Sleep(1 * time.Second)
		}
	}
}

func listDirectory(path string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	fmt.Printf("Contents of %s:\n", absPath)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-10s %-40s %-15s %s\n", "TYPE", "NAME", "SIZE", "MODIFIED")
	fmt.Println(strings.Repeat("-", 80))

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return err
		}

		fileType := "FILE"
		if file.IsDir() {
			fileType = "DIR"
		}

		size := fmt.Sprintf("%d bytes", info.Size())

		fmt.Printf("%-10s %-40s %-15s %s\n",
			fileType,
			file.Name(),
			size,
			info.ModTime().Format("2006-01-02 15:04:05"),
		)
	}
	return nil
}

func deleteFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory, not a file", path)
	}

	fmt.Printf("Are you sure you want to delete '%s'? (y/N): ", path)
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Operation cancelled")
		return nil
	}

	err = os.Remove(path)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully deleted file: %s\n", path)
	return nil
}

func monitorDirectory(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// First list the current contents
	fmt.Printf("Current contents of %s:\n", absPath)
	err = listDirectory(absPath)
	if err != nil {
		fmt.Printf("Error listing directory: %v\n", err)
	}

	fmt.Println("\nStarting monitoring... (Press Ctrl+C to stop)")
	fmt.Println(strings.Repeat("-", 80))

	// Start listening for events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				eventType := ""
				switch {
				case event.Op&fsnotify.Create == fsnotify.Create:
					eventType = "CREATED"
				case event.Op&fsnotify.Write == fsnotify.Write:
					eventType = "MODIFIED"
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					eventType = "DELETED"
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					eventType = "RENAMED"
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					eventType = "CHMOD"
				}

				fmt.Printf("[%s] %s - %s\n",
					time.Now().Format("15:04:05"),
					eventType,
					filepath.Base(event.Name),
				)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("[ERROR] %v\n", err)
			}
		}
	}()

	// Add a path to watch
	err = watcher.Add(absPath)
	if err != nil {
		return err
	}

	// Wait forever
	<-make(chan struct{})
	return nil
}

func addDirectory(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Verify it's a directory
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", absPath)
	}

	// Check if already in the list
	for _, dir := range appConfig.MonitoredDirs {
		if dir == absPath {
			fmt.Printf("Directory %s is already in the monitored list\n", absPath)
			return nil
		}
	}

	// Add to the list
	appConfig.MonitoredDirs = append(appConfig.MonitoredDirs, absPath)
	err = saveConfig()
	if err != nil {
		return err
	}

	fmt.Printf("Directory %s has been added to the monitored list\n", absPath)
	return nil
}

func viewMonitoredDirectories() {
	if len(appConfig.MonitoredDirs) == 0 {
		fmt.Println("No directories are being monitored")
		return
	}

	fmt.Printf("Monitored Directories (saved in %s):\n", configFile)
	fmt.Println(strings.Repeat("-", 80))
	for i, dir := range appConfig.MonitoredDirs {
		fmt.Printf("%d. %s\n", i+1, dir)
	}
}
func removeDirectory() error {
	if len(appConfig.MonitoredDirs) == 0 {
		fmt.Println("No directories are being monitored")
		return nil
	}

	viewMonitoredDirectories()
	fmt.Print("\nEnter the number of the directory to remove (0 to cancel): ")
	var choice int
	fmt.Scanln(&choice)

	if choice <= 0 || choice > len(appConfig.MonitoredDirs) {
		return nil
	}

	removedDir := appConfig.MonitoredDirs[choice-1]
	appConfig.MonitoredDirs = append(appConfig.MonitoredDirs[:choice-1], appConfig.MonitoredDirs[choice:]...)

	err := saveConfig()
	if err != nil {
		return err
	}

	fmt.Printf("Directory %s has been removed from the monitored list\n", removedDir)
	return nil
}

func monitorAllDirectories() error {
	if len(appConfig.MonitoredDirs) == 0 {
		return fmt.Errorf("no directories to monitor")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Add all paths to watch
	for _, dir := range appConfig.MonitoredDirs {
		fmt.Printf("Adding %s to watch list\n", dir)
		err = watcher.Add(dir)
		if err != nil {
			fmt.Printf("Error watching %s: %v\n", dir, err)
		}
	}

	fmt.Println("\nStarting monitoring of all directories... (Press Ctrl+C to stop)")
	fmt.Println(strings.Repeat("-", 80))

	// Start listening for events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				eventType := ""
				switch {
				case event.Op&fsnotify.Create == fsnotify.Create:
					eventType = "CREATED"
				case event.Op&fsnotify.Write == fsnotify.Write:
					eventType = "MODIFIED"
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					eventType = "DELETED"
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					eventType = "RENAMED"
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					eventType = "CHMOD"
				}

				// Get directory path for the event
				dirPath := filepath.Dir(event.Name)

				fmt.Printf("[%s] [%s] %s - %s\n",
					time.Now().Format("15:04:05"),
					dirPath,
					eventType,
					filepath.Base(event.Name),
				)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("[ERROR] %v\n", err)
			}
		}
	}()

	// Wait forever
	<-make(chan struct{})
	return nil
}
