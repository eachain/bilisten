package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"bilisten/backend"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/dock"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

//go:embed frontend/public/favicon.png
var icon []byte

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
func main() {
	log.SetFlags(log.Flags() | log.Lshortfile | log.Lmicroseconds)

	const AppName = "bilisten"
	DataDir := filepath.Join(application.Path(application.PathDataHome), AppName)
	CacheDir := filepath.Join(application.Path(application.PathCacheHome), AppName)

	log.Printf("Cache Home: %v", CacheDir)
	log.Printf("Data Home: %v", DataDir)

	backend.InitLogger(CacheDir)
	backend.InitImageCacheManager(DataDir, CacheDir)
	backend.InitAudioCacheManager(DataDir, CacheDir)
	backend.InitStorageManager(DataDir)

	var window *application.WebviewWindow

	dockService := dock.New()
	open := func() {
		if window != nil {
			window.Show()
			window.Restore() // 若窗口最小化则还原
			window.Focus()   // 置于最前
			dockService.ShowAppIcon()
		}
	}

	app := application.New(application.Options{
		Name:        "Bilisten",
		Description: "Listen music from bilibili audio.",
		Services: []application.Service{
			application.NewService(dockService),
		},
		Assets: application.AssetOptions{
			Handler: backend.NewRouter(application.AssetFileServerFS(assets)),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		SingleInstance: &application.SingleInstanceOptions{
			// 必填：唯一标识，推荐反向域名格式
			UniqueID: "com.eachain.bilisten",
			// 关键回调：第二个实例启动时触发
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				open()
			},
		},
	})

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	window = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Bilisten",
		// Window sized to the golden ratio (1000 / 618 ≈ 1.618).
		Width:     1000,
		Height:    618,
		MinWidth:  1000,
		MinHeight: 618,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 0,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour:           application.NewRGB(6, 7, 15),
		URL:                        "/",
		ZoomControlEnabled:         false,
		DefaultContextMenuDisabled: true,
		DevToolsEnabled:            false,
	})

	InitMenu(app, window)
	InitSystray(app, open)

	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		window.Hide()             // 隐藏窗口
		dockService.HideAppIcon() // 隐藏 Dock 图标
		e.Cancel()                // 阻止窗口关闭
	})

	workdir, _ := os.Getwd()
	backend.NewLogger("main").Info("application try start on: %v", workdir)

	err := app.Run()
	if err != nil {
		backend.NewLogger("main").Error("application start failed: %v", err)
		os.Exit(1)
	}

}

func InitMenu(app *application.App, window *application.WebviewWindow) {
	menu := app.NewMenu()

	appMenu := menu.AddSubmenu("必听")

	appMenu.AddRole(application.About)
	appMenu.FindByRole(application.About).SetLabel("关于必听")
	appMenu.AddSeparator()

	appMenu.Add("设置").SetAccelerator("Cmd+,").OnClick(func(ctx *application.Context) {
		app.Event.Emit("open-settings")
	})
	appMenu.AddSeparator()

	appMenu.AddRole(application.Hide)
	appMenu.FindByRole(application.Hide).SetLabel("隐藏")

	appMenu.AddRole(application.CloseWindow)
	appMenu.FindByRole(application.CloseWindow).SetLabel("关闭")

	appMenu.AddRole(application.Quit)
	appMenu.FindByRole(application.Quit).SetLabel("退出")

	editMenu := menu.AddSubmenu("编辑")
	editMenu.AddRole(application.SelectAll)
	editMenu.FindByRole(application.SelectAll).SetLabel("全选")
	editMenu.AddRole(application.Copy)
	editMenu.FindByRole(application.Copy).SetLabel("复制")
	editMenu.AddRole(application.Paste)
	editMenu.FindByRole(application.Paste).SetLabel("粘贴")
	editMenu.AddSeparator()

	// Set the application menu
	app.Menu.Set(menu)
}

func InitSystray(app *application.App, open func()) {
	systray := app.SystemTray.New()
	systray.SetIcon(icon)

	menu := app.NewMenu()
	menu.Add("显示主界面").OnClick(func(ctx *application.Context) {
		open()
	})
	menu.AddSeparator()

	menu.Add("播放/暂停").OnClick(func(ctx *application.Context) {
		app.Event.Emit("play-toggle")
	})
	menu.Add("下一曲").OnClick(func(ctx *application.Context) {
		app.Event.Emit("play-next")
	})
	menu.Add("上一曲").OnClick(func(ctx *application.Context) {
		app.Event.Emit("play-prev")
	})
	menu.AddSeparator()

	menu.Add("退出").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	systray.SetMenu(menu)
}
