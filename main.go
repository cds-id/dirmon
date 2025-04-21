package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
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
				Name:    "cleanup-advice",
				Aliases: []string{"ca"},
				Usage:   "Analyze directory and provide file deletion recommendations",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "age",
						Value: 90,
						Usage: "Age threshold in days for old file detection",
					},
					&cli.IntFlag{
						Name:  "size",
						Value: 100,
						Usage: "Size threshold in MB for large file detection",
					},
				},
				Action: func(c *cli.Context) error {
					path := "."
					if c.NArg() > 0 {
						path = c.Args().Get(0)
					}
					return provideCleanupAdvice(path, c.Int("age"), c.Int("size"))
				},
			},
			{
				Name:    "find-duplicates",
				Aliases: []string{"fd"},
				Usage:   "Find duplicate files in a directory",
				Action: func(c *cli.Context) error {
					path := "."
					if c.NArg() > 0 {
						path = c.Args().Get(0)
					}
					return findDuplicateFiles(path)
				},
			},
			{
				Name:    "disk-usage",
				Aliases: []string{"du"},
				Usage:   "Analyze disk usage in a directory",
				Action: func(c *cli.Context) error {
					path := "."
					if c.NArg() > 0 {
						path = c.Args().Get(0)
					}
					return analyzeDiskUsage(path)
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
		fmt.Println("8. Get cleanup advice")
		fmt.Println("9. Find duplicate files")
		fmt.Println("10. Analyze disk usage")
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
		case "8":
			fmt.Print("Enter directory path (or press Enter for current directory): ")
			path, _ := reader.ReadString('\n')
			path = strings.TrimSpace(path)
			if path == "" {
				path = "."
			}

			fmt.Print("Enter age threshold in days (or press Enter for default 90): ")
			ageStr, _ := reader.ReadString('\n')
			ageStr = strings.TrimSpace(ageStr)
			age := 90
			if ageStr != "" {
				if n, err := fmt.Sscanf(ageStr, "%d", &age); n != 1 || err != nil {
					fmt.Println("Invalid age, using default 90 days")
					age = 90
				}
			}

			fmt.Print("Enter size threshold in MB (or press Enter for default 100): ")
			sizeStr, _ := reader.ReadString('\n')
			sizeStr = strings.TrimSpace(sizeStr)
			size := 100
			if sizeStr != "" {
				if n, err := fmt.Sscanf(sizeStr, "%d", &size); n != 1 || err != nil {
					fmt.Println("Invalid size, using default 100 MB")
					size = 100
				}
			}

			err := provideCleanupAdvice(path, age, size)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')

		case "9":
			fmt.Print("Enter directory path (or press Enter for current directory): ")
			path, _ := reader.ReadString('\n')
			path = strings.TrimSpace(path)
			if path == "" {
				path = "."
			}

			err := findDuplicateFiles(path)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')

		case "10":
			fmt.Print("Enter directory path (or press Enter for current directory): ")
			path, _ := reader.ReadString('\n')
			path = strings.TrimSpace(path)
			if path == "" {
				path = "."
			}

			err := analyzeDiskUsage(path)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
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

// provideCleanupAdvice analyzes files in a directory and recommends which ones to delete
func provideCleanupAdvice(path string, ageThreshold, sizeThreshold int) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	fmt.Printf("Cleanup advice for %s:\n", absPath)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-40s %-15s %-20s %s\n", "FILENAME", "SIZE", "MODIFIED", "REASON")
	fmt.Println(strings.Repeat("-", 80))

	var totalPotentialSavings int64
	var recommendedFiles []string

	now := time.Now()
	ageThresholdDuration := time.Duration(ageThreshold*24) * time.Hour
	sizeThresholdBytes := int64(sizeThreshold * 1024 * 1024)

	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories for now
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		filePath := filepath.Join(path, file.Name())
		fileAge := now.Sub(info.ModTime())
		reason := ""

		// Check for temporary or log files
		if isTempFile(file.Name()) {
			reason = "Temporary file"
		} else if isLogFile(file.Name()) {
			reason = "Log file"
		} else if fileAge > ageThresholdDuration && info.Size() > 0 {
			reason = fmt.Sprintf("Not modified for %d days", int(fileAge.Hours()/24))
		} else if info.Size() > sizeThresholdBytes {
			reason = fmt.Sprintf("Large file (%s)", formatSize(info.Size()))
		}

		if reason != "" {
			fmt.Printf("%-40s %-15s %-20s %s\n",
				truncateString(file.Name(), 39),
				formatSize(info.Size()),
				info.ModTime().Format("2006-01-02"),
				reason)

			totalPotentialSavings += info.Size()
			recommendedFiles = append(recommendedFiles, filePath)
		}
	}

	if len(recommendedFiles) == 0 {
		fmt.Println("No files recommended for deletion.")
		return nil
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Potential space savings: %s\n", formatSize(totalPotentialSavings))

	fmt.Println("\nWould you like to delete these files? (y/N):")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		for _, filePath := range recommendedFiles {
			if err := os.Remove(filePath); err != nil {
				fmt.Printf("Error deleting %s: %v\n", filePath, err)
			} else {
				fmt.Printf("Deleted: %s\n", filePath)
			}
		}
	}

	return nil
}

// findDuplicateFiles identifies potential duplicate files in a directory
func findDuplicateFiles(path string) error {
	// First pass: get file sizes and organize by size
	filesBySize := make(map[int64][]string)

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			filesBySize[info.Size()] = append(filesBySize[info.Size()], filePath)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Second pass: compute MD5 hashes for potential duplicates (files with same size)
	duplicateGroups := make(map[string][]string)

	for size, files := range filesBySize {
		if len(files) > 1 && size > 0 {
			// Files with the same size are potential duplicates
			for _, file := range files {
				hash, err := calculateMD5(file)
				if err != nil {
					fmt.Printf("Error calculating hash for %s: %v\n", file, err)
					continue
				}

				duplicateGroups[hash] = append(duplicateGroups[hash], file)
			}
		}
	}

	// Display results
	duplicateCount := 0
	var totalWasted int64

	fmt.Println("Duplicate files:")
	fmt.Println(strings.Repeat("-", 80))

	for hash, files := range duplicateGroups {
		if len(files) > 1 {
			duplicateCount++

			// Get file size (all files in this group have the same size)
			info, err := os.Stat(files[0])
			if err != nil {
				continue
			}

			// Calculate wasted space
			wastedSpace := info.Size() * int64(len(files)-1)
			totalWasted += wastedSpace

			fmt.Printf("\nDuplicate Group %d (%s, wasted: %s):\n",
				duplicateCount, hash[:8], formatSize(wastedSpace))

			for i, file := range files {
				fmt.Printf("%d. %s\n", i+1, file)
			}
		}
	}

	if duplicateCount == 0 {
		fmt.Println("No duplicate files found.")
		return nil
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Found %d groups of duplicate files\n", duplicateCount)
	fmt.Printf("Potential space savings: %s\n", formatSize(totalWasted))

	return nil
}

// analyzeDiskUsage shows disk usage by file types and directories
func analyzeDiskUsage(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Collect stats by file type and directory
	typeStats := make(map[string]int64)
	dirStats := make(map[string]int64)

	var totalSize int64

	err = filepath.Walk(absPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip the root directory itself
		if filePath == absPath {
			return nil
		}

		if !info.IsDir() {
			// Update total size
			totalSize += info.Size()

			// Update file type stats
			ext := strings.ToLower(filepath.Ext(filePath))
			if ext == "" {
				ext = "[no extension]"
			}
			typeStats[ext] += info.Size()

			// Update directory stats (by parent directory)
			parentDir := filepath.Dir(filePath)
			dirStats[parentDir] += info.Size()
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Display results by file type
	fmt.Printf("Disk usage analysis for: %s\n\n", absPath)
	fmt.Println("Usage by file type:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-20s %-15s %s\n", "FILE TYPE", "SIZE", "% OF TOTAL")
	fmt.Println(strings.Repeat("-", 60))

	// Convert to slice for sorting
	type typeStat struct {
		ext  string
		size int64
	}

	typeStatsList := make([]typeStat, 0, len(typeStats))
	for ext, size := range typeStats {
		typeStatsList = append(typeStatsList, typeStat{ext, size})
	}

	// Sort by size (descending)
	sort.Slice(typeStatsList, func(i, j int) bool {
		return typeStatsList[i].size > typeStatsList[j].size
	})

	for _, stat := range typeStatsList {
		percentage := float64(stat.size) / float64(totalSize) * 100
		fmt.Printf("%-20s %-15s %.1f%%\n",
			stat.ext, formatSize(stat.size), percentage)
	}

	// Display results by directory
	fmt.Println("\nLargest directories:")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("%-50s %-15s\n", "DIRECTORY", "SIZE")
	fmt.Println(strings.Repeat("-", 70))

	// Convert to slice for sorting
	type dirStat struct {
		dir  string
		size int64
	}

	dirStatsList := make([]dirStat, 0, len(dirStats))
	for dir, size := range dirStats {
		dirStatsList = append(dirStatsList, dirStat{dir, size})
	}

	// Sort by size (descending)
	sort.Slice(dirStatsList, func(i, j int) bool {
		return dirStatsList[i].size > dirStatsList[j].size
	})

	// Show top 10 directories
	count := 0
	for _, stat := range dirStatsList {
		relPath, err := filepath.Rel(absPath, stat.dir)
		if err != nil {
			relPath = stat.dir
		}

		if relPath == "." {
			relPath = "[root directory]"
		}

		fmt.Printf("%-50s %-15s\n",
			truncateString(relPath, 49), formatSize(stat.size))

		count++
		if count >= 10 {
			break
		}
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Total size: %s\n", formatSize(totalSize))

	return nil
}

// Helper functions
func isTempFile(filename string) bool {
	lowerName := strings.ToLower(filename)
	return strings.HasSuffix(lowerName, ".tmp") ||
		strings.HasSuffix(lowerName, ".temp") ||
		strings.HasPrefix(lowerName, "~") ||
		strings.HasPrefix(lowerName, "temp_") ||
		strings.Contains(lowerName, "cache") ||
		strings.HasSuffix(lowerName, ".bak")
}

func isLogFile(filename string) bool {
	lowerName := strings.ToLower(filename)
	return strings.HasSuffix(lowerName, ".log") ||
		strings.HasSuffix(lowerName, ".log.gz") ||
		strings.HasSuffix(lowerName, ".logs") ||
		strings.Contains(lowerName, "debug")
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func calculateMD5(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a new hash
	hash := md5.New()

	// Copy file content to the hash
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// Get the hash sum
	hashInBytes := hash.Sum(nil)

	// Convert to string
	hashString := hex.EncodeToString(hashInBytes)

	return hashString, nil
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
