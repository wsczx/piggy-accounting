package main

import (
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"piggy-accounting/backend/service"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	// 提前初始化所有 service（Bind 需要非 nil 的结构体指针）
	service.InitServices()

	// macOS 使用非 Frameless 模式以保留系统圆角
	// Windows 使用 Frameless 模式去掉系统边框并使用自定义标题栏
	isWindows := runtime.GOOS == "windows"

	err := wails.Run(&options.App{
		Title:            "猪猪记账",
		Width:            900,
		Height:           780,
		MinWidth:         900,
		MinHeight:        780,
		Frameless:        isWindows,
		DisableResize:    true,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHidden(),
			WindowIsTranslucent:  true,
			WebviewIsTransparent: true,
			About: &mac.AboutInfo{
				Title:   "猪猪记账",
				Message: "一款精美的个人记账工具\n\n© 2025 By 孤鸿",
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			DisableWindowIcon:                 true,
			BackdropType:                      windows.Auto,
			DisableFramelessWindowDecorations: true,
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []any{
			app,
			service.Records,
			service.Categories,
			service.Budgets,
			service.ExportImport,
			service.SmartRecognize,
			service.Tags,
			service.Reminders,
			service.Recurring,
			service.Accounts,
			service.Transfers,
			service.Backups,
			service.Ledgers,
			service.Tasks,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
