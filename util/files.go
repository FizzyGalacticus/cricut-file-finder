package util

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Image files are stored in <home>/.cricut-design-space/LocalData/<numerical_id>/<random_id>/Images
const CRICUT_DIR_NAME = ".cricut-design-space"

type CricutFile struct {
	Name         string
	Path         string
	FullPath     string
	LastModified time.Time
}

func SortByLastModifiedDesc(entries []CricutFile) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastModified.After(entries[j].LastModified)
	})
}

func pathExists(path string) (bool, error) {
	if stat, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil // path does not exist
		}

		return false, err
	} else {
		return stat.IsDir(), nil
	}
}

func getHomeDir() string {
	// Try os/user first
	if u, err := user.Current(); err == nil && u.HomeDir != "" {
		return u.HomeDir
	}

	// Fallback to environment variables
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}

	return os.Getenv("HOME")
}

func createListing(filetype string) func(string) ([]os.FileInfo, error) {
	return func(path string) ([]os.FileInfo, error) {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		var dirs []os.FileInfo
		for _, entry := range entries {
			include := false

			if filetype == "file" {
				include = !entry.IsDir()
			} else if filetype == "dir" {
				include = entry.IsDir()
			}

			if include {
				info, err := entry.Info()

				if err != nil {
					return dirs, err
				}

				dirs = append(dirs, info)
			}
		}

		return dirs, nil
	}
}

var listDirectories = createListing("dir")
var listFiles = createListing("file")

func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	_, errInt := strconv.ParseInt(s, 10, 64)
	if errInt == nil {
		return true
	}

	_, errFloat := strconv.ParseFloat(s, 64)
	return errFloat == nil
}

func GetCricutFiles() ([]CricutFile, error) {
	ret := []CricutFile{}

	homeDir := getHomeDir()

	cricutDir := filepath.Join(homeDir, CRICUT_DIR_NAME)

	if exists, ok := pathExists(cricutDir); ok != nil {
		return ret, fmt.Errorf("error checking if cricut directory exists: %w", ok)
	} else if !exists {
		return ret, fmt.Errorf("Cricut directory does not exist at %s", cricutDir)
	} else {
		fmt.Printf("Cricut directory found at %s\n", cricutDir)
	}

	topDir := filepath.Join(cricutDir, "LocalData")

	if exists, ok := pathExists(topDir); ok != nil {
		return ret, fmt.Errorf("error checking if cricut directory exists: %w", ok)
	} else if !exists {
		return ret, fmt.Errorf("Cricut directory does not exist at %s", topDir)
	} else {
		fmt.Printf("Cricut LocalData directory found at %s\n", topDir)
	}

	projectDirs, err := listDirectories(topDir)

	if err != nil {
		return ret, err
	}

	for _, dir := range projectDirs {
		if !isNumeric(dir.Name()) {
			fmt.Printf("Skipping non-numeric directory: %s%s%s\n", topDir, string(os.PathSeparator), dir.Name())
			continue // Skip non-numeric directories
		}

		topCanvasDir := filepath.Join(topDir, dir.Name(), "Canvas")

		if exists, ok := pathExists(topCanvasDir); ok != nil {
			return ret, fmt.Errorf("error checking if canvas directory exists: %w", ok)
		} else if !exists {
			fmt.Printf("Canvas directory does not exist at %s, skipping\n", topCanvasDir)
			continue // Skip if the Canvas directory does not exist
		}

		canvasDirs, err := listDirectories(topCanvasDir)

		if err != nil {
			return ret, fmt.Errorf("error reading canvas directory %s: %w", topCanvasDir, err)
		}

		for _, dir := range canvasDirs {
			if !isNumeric(dir.Name()) {
				fmt.Printf("Skipping non-numeric canvas directory: %s%s%s\n", dir, string(os.PathSeparator), dir.Name())
				continue // Skip non-numeric directories
			}

			canvasDir := filepath.Join(topCanvasDir, dir.Name())

			files, err := listFiles(canvasDir)

			if err != nil {
				return ret, fmt.Errorf("error reading canvas directory %s: %w", canvasDir, err)
			}

			for _, file := range files {
				if strings.Contains(file.Name(), ".png") || strings.Contains(file.Name(), ".PNG") {
					ret = append(ret, CricutFile{
						Name:         file.Name(),
						Path:         canvasDir,
						FullPath:     filepath.Join(canvasDir, file.Name()),
						LastModified: file.ModTime(),
					})
				}
			}
		}
	}

	return ret, nil
}
