package main

import (
	"cricut-file-finder/util"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const APPLICATION_TITLE = "Cricut Design Space File Viewer"

var GitSHA = "unknown"

// OpenDirInExplorer opens the given path in the system's native file explorer
func OpenDirInExplorer(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// DynamicTable holds rows and renders them using containers
type DynamicTable struct {
	Rows [][]fyne.CanvasObject
}

// NewDynamicTable creates and returns a new container with the dynamic table layout
func NewDynamicTable(rows [][]fyne.CanvasObject) *DynamicTable {
	return &DynamicTable{Rows: rows}
}

// Render returns the fyne.CanvasObject containing the full table
func (dt *DynamicTable) Render() fyne.CanvasObject {
	var rowContainers []fyne.CanvasObject

	for _, row := range dt.Rows {
		rowContainer := container.NewHBox(row...)
		rowContainers = append(rowContainers, rowContainer)
	}

	return container.NewVBox(rowContainers...)
}

// Render returns a fyne.CanvasObject with equal column widths and dynamic row heights
func (dt *DynamicTable) Render2() fyne.CanvasObject {
	if len(dt.Rows) == 0 {
		return widget.NewLabel("No data")
	}

	colCount := len(dt.Rows[0])
	colWidths := make([]float32, colCount)

	// Step 1: Find max preferred width per column
	for _, row := range dt.Rows {
		for col, cell := range row {
			size := cell.MinSize()
			if size.Width > colWidths[col] {
				colWidths[col] = size.Width
			}
		}
	}

	// Step 2: Set min size for each cell to max width for that column
	var rowContainers []fyne.CanvasObject
	for _, row := range dt.Rows {
		if len(row) != colCount {
			continue // skip malformed rows
		}
		for i, cell := range row {
			cell.Resize(fyne.NewSize(colWidths[i], cell.MinSize().Height))
			cell.(*fyne.Container).Objects[0].Resize(fyne.NewSize(colWidths[i], cell.MinSize().Height))
			cell.(*fyne.Container).Objects[0].Refresh()
		}
		rowContainers = append(rowContainers, container.NewHBox(row...))
	}

	return container.NewVBox(rowContainers...)
}

// Convenience function to load an image (scaled to fit)
func LoadImageCell(path string, size fyne.Size) fyne.CanvasObject {
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("image missing: %s", path)
		return widget.NewLabel("No Image")
	}

	img := canvas.NewImageFromFile(path)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(size)
	return img
}

// Convenience function for a cell with padded label
func LabelCell(text string) fyne.CanvasObject {
	lbl := widget.NewLabel(text)
	return container.NewPadded(lbl)
}

// Example usage
func main() {
	a := app.New()
	w := a.NewWindow(APPLICATION_TITLE)
	w.Resize(fyne.NewSize(800, 600))

	// Sample data
	rows := [][]fyne.CanvasObject{
		{LabelCell("Image"), LabelCell("Name"), LabelCell("Path"), LabelCell("")},
	}

	if entries, err := util.GetCricutFiles(); err != nil {
		fmt.Printf("error getting Cricut files: %v", err)
	} else {
		util.SortByLastModifiedDesc(entries)

		for _, entry := range entries {
			rows = append(rows, []fyne.CanvasObject{
				LoadImageCell(entry.FullPath, fyne.NewSize(100, 100)),
				LabelCell(entry.Name),
				LabelCell(entry.Path),
				container.NewPadded(widget.NewButton("Open", func() {
					if err := OpenDirInExplorer(entry.Path); err != nil {
						fmt.Printf("error opening directory: %v", err)
					}
				})),
			})
		}

	}

	dt := NewDynamicTable(rows)
	table := dt.Render()

	scroll := container.NewScroll(table)
	scroll.SetMinSize(fyne.NewSize(800, 600))

	w.SetContent(scroll)
	w.ShowAndRun()
}

// CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -o cricut_finder.exe .
