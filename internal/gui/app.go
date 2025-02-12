package gui

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	dicontainer "github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/container"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/lifecycle"
)

// App represents the GUI application
type App struct {
	*lifecycle.BaseComponent
	guiContainer *fyne.Container
	app          fyne.App
	window       fyne.Window
	monContainer *dicontainer.Container
}

// NewApp creates a new GUI application
func NewApp(monContainer *dicontainer.Container) (*App, error) {
	return &App{
		BaseComponent: lifecycle.NewBaseComponent("GUIApp"),
		monContainer:  monContainer,
		app:           app.New(),
	}, nil
}

// Start starts the GUI application
func (a *App) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	// Start container first
	if err := a.monContainer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Create main window
	a.window = a.app.NewWindow("Dropbox Monitor")

	// Create status label
	statusLabel := widget.NewLabel("Status: Running")

	// Create content
	a.guiContainer = container.NewVBox(
		widget.NewLabel("Dropbox Monitor"),
		statusLabel,
	)

	// Set window content
	a.window.SetContent(a.guiContainer)

	// Show and run
	a.window.Show()
	go a.app.Run()

	a.SetState(lifecycle.StateRunning)
	return nil
}

// Stop stops the GUI application
func (a *App) Stop(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	// TODO: Clean up GUI components here

	// Stop container
	if err := a.monContainer.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	if a.window != nil {
		a.window.Close()
	}

	a.SetState(lifecycle.StateStopped)
	return nil
}

// Health checks the health of the GUI application
func (a *App) Health(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	if a.State() != lifecycle.StateRunning {
		return lifecycle.ErrNotRunning
	}

	return a.monContainer.Health(ctx)
}
