package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
)

func createResultContainer(title string, entries []dropbox.FolderInfo) fyne.CanvasObject {
	header := widget.NewLabel(fmt.Sprintf("%s (%d entries)", title, len(entries)))
	vbox := container.NewVBox(header)
	for _, entry := range entries {
		vbox.Add(widget.NewLabel(fmt.Sprintf("%s - %s", entry.Name, entry.LastModified)))
	}
	return vbox
}

func buttonContainer(btnListFolders, btnLastChanged, btnChanges24h *widget.Button) *fyne.Container {
	return container.NewHBox(btnListFolders, btnLastChanged, btnChanges24h)
}

func main() {
	a := app.New()
	w := a.NewWindow("Dropbox Monitor GUI")
	w.Resize(fyne.NewSize(800, 600))

	// Declare buttons as variables so they can be used in closures
	var btnListFolders, btnLastChanged, btnChanges24h *widget.Button

	btnListFolders = widget.NewButton("List Folders", func() {
		folderNames := dropbox.GetFolders()
		var infos []dropbox.FolderInfo
		for _, name := range folderNames {
			infos = append(infos, dropbox.FolderInfo{Name: name, LastModified: "N/A"})
		}
		content := createResultContainer("List Folders", infos)
		w.SetContent(container.NewBorder(
			buttonContainer(btnListFolders, btnLastChanged, btnChanges24h),
			nil,
			nil,
			nil,
			container.NewVScroll(content),
		))
	})

	btnLastChanged = widget.NewButton("Last Changed Dates", func() {
		infos := dropbox.GetLastChangedFolders()
		content := createResultContainer("Last Changed", infos)
		w.SetContent(container.NewBorder(
			buttonContainer(btnListFolders, btnLastChanged, btnChanges24h),
			nil,
			nil,
			nil,
			container.NewVScroll(content),
		))
	})

	btnChanges24h = widget.NewButton("Changes in Last 24 Hours", func() {
		infos := dropbox.GetChangesLast24Hours()
		content := createResultContainer("Changes Last 24 Hours", infos)
		w.SetContent(container.NewBorder(
			buttonContainer(btnListFolders, btnLastChanged, btnChanges24h),
			nil,
			nil,
			nil,
			container.NewVScroll(content),
		))
	})

	initialContent := container.NewVBox(
		buttonContainer(btnListFolders, btnLastChanged, btnChanges24h),
		widget.NewLabel("Select an action to view Dropbox info."),
	)
	w.SetContent(initialContent)
	w.ShowAndRun()
}
