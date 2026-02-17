package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// Service 表示一个后台服务
type Service struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	ExePath    string    `json:"exePath"`
	Args       string    `json:"args"`
	WorkingDir string    `json:"workingDir"`
	Status     string    `json:"status"` // "running", "stopped", "error"
	PID        int       `json:"pid"`
	AutoStart  bool      `json:"autoStart"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// ServiceConfig 用于创建新服务的配置
type ServiceConfig struct {
	Name       string `json:"name"`
	ExePath    string `json:"exePath"`
	Args       string `json:"args"`
	WorkingDir string `json:"workingDir"`
}

type App struct {
	ctx                context.Context
	serviceManager     *WindowsServiceManager
	environmentManager *EnvironmentManager
}

func NewApp() *App {
	return &App{
		serviceManager:     NewWindowsServiceManager(),
		environmentManager: NewEnvironmentManager(),
	}
}

// startup 在应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.serviceManager.SetContext(ctx)
	a.serviceManager.loadServices()
}

// GetServices 获取所有服务列表
func (a *App) GetServices() []*Service {
	services, err := a.serviceManager.GetServices()
	if err != nil {
		return []*Service{}
	}
	return services
}

// CreateService 创建新的服务
func (a *App) CreateService(config ServiceConfig) (*Service, error) {
	return a.serviceManager.CreateService(config)
}

// StartService 启动服务
func (a *App) StartService(serviceID string) error {
	return a.serviceManager.StartService(serviceID)
}

// StopService 停止服务
func (a *App) StopService(serviceID string) error {
	return a.serviceManager.StopService(serviceID)
}

// DeleteService 删除服务
func (a *App) DeleteService(serviceID string) error {
	return a.serviceManager.DeleteService(serviceID)
}

// SelectFile 选择文件对话框
func (a *App) SelectFile() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择可执行文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "可执行文件 (*.exe)",
				Pattern:     "*.exe",
			},
			{
				DisplayName: "所有文件 (*.*)",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		return "", err
	}

	return selection, nil
}

// SelectDirectory 选择目录对话框
func (a *App) SelectDirectory() (string, error) {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择工作目录",
	})

	if err != nil {
		return "", err
	}

	return selection, nil
}

// CheckAdminPrivileges 检查管理员权限
func (a *App) CheckAdminPrivileges() bool {
	return a.environmentManager.IsAdmin()
}

// SetAutoStart 设置开机自启动
func (a *App) SetAutoStart(enabled bool) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %v", err)
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %v", err)
	}
	defer key.Close()

	appName := "WindowsServiceManager"

	if enabled {
		err = key.SetStringValue(appName, execPath)
		if err != nil {
			return fmt.Errorf("设置启动项失败: %v", err)
		}
	} else {
		err = key.DeleteValue(appName)
		if err != nil && err != registry.ErrNotExist {
			return fmt.Errorf("删除启动项失败: %v", err)
		}
	}

	return nil
}

// GetAutoStartStatus 获取开机自启动状态
func (a *App) GetAutoStartStatus() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue("WindowsServiceManager")
	return err == nil
}

// RestartAsAdmin 以管理员权限重启应用
func (a *App) RestartAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	args := strings.Join(os.Args[1:], " ")

	verbPtr, err := syscall.UTF16PtrFromString("runas")
	if err != nil {
		return err
	}
	exePtr, err := syscall.UTF16PtrFromString(exe)
	if err != nil {
		return err
	}
	cwdPtr, err := syscall.UTF16PtrFromString(cwd)
	if err != nil {
		return err
	}
	argPtr, err := syscall.UTF16PtrFromString(args)
	if err != nil {
		return err
	}

	var showCmd int32 = 1

	err = windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func (a *App) ShowWindow() {
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowCenter(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
}

func (a *App) HideWindow() {
	runtime.WindowHide(a.ctx)
}

// SetServiceAutoStart 设置服务开机自启动
func (a *App) SetServiceAutoStart(serviceID string, enabled bool) error {
	return a.serviceManager.SetServiceAutoStart(serviceID, enabled)
}

// GetServiceAutoStart 获取服务开机自启动状态
func (a *App) GetServiceAutoStart(serviceID string) bool {
	return a.serviceManager.GetServiceAutoStart(serviceID)
}

// AddSystemEnvironmentVariable 添加系统环境变量
func (a *App) AddSystemEnvironmentVariable(varName, varValue string) error {
	return a.environmentManager.AddSystemEnvironmentVariable(varName, varValue)
}

// AddPathVariable 添加PATH环境变量
func (a *App) AddPathVariable(pathValue string) error {
	return a.environmentManager.AddPathVariable(pathValue)
}

// OpenSystemEnvironmentSettings 打开系统环境变量设置
func (a *App) OpenSystemEnvironmentSettings() error {
	return a.environmentManager.OpenSystemEnvironmentSettings()
}

// ValidatePathExists 验证路径是否存在
func (a *App) ValidatePathExists(path string) bool {
	return a.environmentManager.ValidatePathExists(path)
}

// DiagnoseEnvironmentAccess 诊断环境变量访问权限
func (a *App) DiagnoseEnvironmentAccess() (map[string]interface{}, error) {
	return a.environmentManager.DiagnoseEnvironmentAccess()
}

// GetServiceLogs 获取服务日志内容（最新日志）
func (a *App) GetServiceLogs(serviceID string) (string, error) {
	logDir := filepath.Join(os.Getenv("ProgramData"), "windows_service_logs")

	log.Printf("GetServiceLogs: 查找服务 %s 的日志文件", serviceID)

	files, err := filepath.Glob(filepath.Join(logDir, fmt.Sprintf("%s_*.log", serviceID)))
	if err != nil {
		log.Printf("GetServiceLogs: 查找日志文件失败: %v", err)
		return "", fmt.Errorf("查找日志文件失败: %v", err)
	}

	if len(files) == 0 {
		log.Printf("GetServiceLogs: 未找到日志文件")
		return "", fmt.Errorf("未找到日志文件")
	}

	latestFile := files[len(files)-1]
	log.Printf("GetServiceLogs: 找到 %d 个日志文件，使用最新的: %s", len(files), latestFile)

	content, err := os.ReadFile(latestFile)
	if err != nil {
		log.Printf("GetServiceLogs: 读取日志文件失败: %v", err)
		return "", fmt.Errorf("读取日志文件失败: %v", err)
	}

	log.Printf("GetServiceLogs: 成功读取日志文件，大小: %d 字节", len(content))
	return string(content), nil
}

// GetServiceLogsPath 获取服务日志文件路径（最新日志）
func (a *App) GetServiceLogsPath(serviceID string) (string, error) {
	logDir := filepath.Join(os.Getenv("ProgramData"), "windows_service_logs")

	log.Printf("GetServiceLogsPath: 查找服务 %s 的日志文件", serviceID)

	files, err := filepath.Glob(filepath.Join(logDir, fmt.Sprintf("%s_*.log", serviceID)))
	if err != nil {
		log.Printf("GetServiceLogsPath: 查找日志文件失败: %v", err)
		return "", fmt.Errorf("查找日志文件失败: %v", err)
	}

	if len(files) == 0 {
		log.Printf("GetServiceLogsPath: 未找到日志文件")
		return "", fmt.Errorf("日志文件不存在")
	}

	latestFile := files[len(files)-1]
	log.Printf("GetServiceLogsPath: 找到 %d 个日志文件，返回最新的: %s", len(files), latestFile)
	return latestFile, nil
}

// OpenLogsDirectory 打开日志目录
func (a *App) OpenLogsDirectory(serviceID string) error {
	logDir := filepath.Join(os.Getenv("ProgramData"), "windows_service_logs")

	log.Printf("OpenLogsDirectory: 打开日志目录: %s", logDir)

	if _, err := os.Stat(logDir); err != nil {
		log.Printf("OpenLogsDirectory: 日志目录不存在: %v", err)
		return fmt.Errorf("日志目录不存在")
	}

	runtime.BrowserOpenURL(a.ctx, "file:///"+logDir)
	log.Printf("OpenLogsDirectory: 已打开日志目录")
	return nil
}
