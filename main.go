package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        APP_NAME,
		Description: "Utilize Windows notification for backlog API",
		Services: []application.Service{
			application.NewService(&Service{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	store, _ := loadStore()
	// If store doesn't exist, open window to enter API Key
	// Otherwise, hide the window and show system tray icon only
	app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Name:   APP_NAME,
		Title:  APP_TITLE,
		Width:  WINDOW_WIDTH,
		Height: WINDOW_HEIGHT,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundType: application.BackgroundType(application.Acrylic),
		URL:            "/",
		Hidden:         store.APIKey != "",
		ShouldClose: func(window *application.WebviewWindow) bool {
			window.Hide()
			return false
		},
	})
	addSystemTray(app)

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// Add system tray icon
func addSystemTray(app *application.App) {
	tray := app.NewSystemTray()
	tray.SetLabel(APP_NAME)
	tray.OnClick(func() {
		app.GetWindowByName(APP_TITLE).Show()
	})

	openMenuItem := application.NewMenuItem("Open")
	exitMenuItem := application.NewMenuItem("Exit")

	openMenuItem.OnClick(func(ctx *application.Context) {
		app.GetWindowByName(APP_TITLE).Show()
	})

	exitMenuItem.OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	menu := application.NewMenuFromItems(openMenuItem, exitMenuItem)
	tray.SetMenu(menu)
}
