package main

import (
	"context"
	"embed"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed embed/icon.ico
var trayIcon []byte

var (
	modshell32        = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW = modshell32.NewProc("ShellExecuteW")
)

const (
	SW_SHOW = 1
)

func runAsAdmin() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("获取可执行文件路径失败: %v", err)
	}

	exePath, err := filepath.Abs(exe)
	if err != nil {
		log.Fatalf("获取绝对路径失败: %v", err)
	}

	verb, _ := syscall.UTF16PtrFromString("runas")
	exePtr, _ := syscall.UTF16PtrFromString(exePath)
	params, _ := syscall.UTF16PtrFromString("")
	dir, _ := syscall.UTF16PtrFromString(filepath.Dir(exePath))

	ret, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(exePtr)),
		uintptr(unsafe.Pointer(params)),
		uintptr(unsafe.Pointer(dir)),
		SW_SHOW,
	)

	if ret <= 32 {
		log.Fatalf("以管理员方式启动失败")
	}

	os.Exit(0)
}

func isDevMode() bool {
	exe, err := os.Executable()
	if err == nil {
		if strings.Contains(strings.ToLower(exe), "-dev") {
			log.Printf("检测到开发模式可执行文件: %s", exe)
			return true
		}
	}

	devMode := false
	if len(os.Args) > 1 && os.Args[1] == "dev" {
		devMode = true
	}
	if os.Getenv("WAILS_DEV") == "1" || os.Getenv("WAILS_DEVKIT") == "1" {
		devMode = true
	}
	if os.Getenv("WAILS_GO_BUILD_TAGS") != "" {
		devMode = true
	}
	if os.Getenv("WAILS_DEV") != "" {
		devMode = true
	}
	log.Printf("开发模式检测: %v (Args: %v, WAILS_DEV: %s)", devMode, os.Args, os.Getenv("WAILS_DEV"))
	return devMode
}

func main() {
	app := NewApp()

	if !app.environmentManager.IsAdmin() {
		if isDevMode() {
			log.Println("警告：开发模式下未以管理员权限运行，某些功能可能无法使用")
		} else {
			log.Println("检测到未以管理员权限运行，正在以管理员方式重新启动...")
			runAsAdmin()
			return
		}
	}

	if isWrapper, serviceName := IsServiceWrapperMode(); isWrapper {
		config, configErr := LoadServiceConfigFromRegistry(serviceName)
		if configErr != nil {
			log.Fatalf("加载服务配置失败: %v", configErr)
		}

		if err := RunAsWindowsService(serviceName, *config); err != nil {
			log.Fatalf("运行Windows服务失败: %v", err)
		}
		return
	}

	systrayManager := NewSystrayManager(app, trayIcon)

	err := wails.Run(&options.App{
		Title:     "Windows Service Manager",
		Width:     900,
		Height:    650,
		MinWidth:  750,
		MinHeight: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 239, G: 244, B: 249, A: 1},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "Windows-Service-Manager",
			OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
				runtime.Show(app.ctx)
				runtime.WindowUnminimise(app.ctx)
			},
		},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
		},
		OnDomReady: func(ctx context.Context) {
			go systrayManager.Start()
		},
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			runtime.WindowHide(ctx)
			return true
		},
		OnShutdown: func(ctx context.Context) {
			systrayManager.Cleanup()
			os.Exit(0)
		},
		Bind: []interface{}{
			app,
		},
		WindowStartState: options.Normal,
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisablePinchZoom:     false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
