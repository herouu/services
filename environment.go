package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// EnvironmentManager 环境变量管理器
type EnvironmentManager struct{}

func NewEnvironmentManager() *EnvironmentManager {
	return &EnvironmentManager{}
}

// IsAdmin 检查是否以管理员权限运行
func (em *EnvironmentManager) IsAdmin() bool {
	if _, err := os.Open("\\\\.\\PHYSICALDRIVE0"); err == nil {
		return true
	}

	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token, err := windows.OpenCurrentProcessToken()
	if err != nil {
		token, err = openCurrentThreadTokenSafe()
		if err != nil {
			return false
		}
	}
	defer token.Close()

	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}

	return member
}

// openCurrentThreadTokenSafe 安全地获取当前线程的访问令牌
func openCurrentThreadTokenSafe() (windows.Token, error) {
	if err := impersonateSelf(); err != nil {
		return 0, err
	}
	defer revertToSelf()

	thread, err := getCurrentThread()
	if err != nil {
		return 0, err
	}

	var token windows.Token
	err = openThreadToken(thread, windows.TOKEN_QUERY, true, &token)
	if err != nil {
		return 0, err
	}

	return token, nil
}

// Windows API 函数声明
var (
	modadvapi32 = windows.NewLazySystemDLL("advapi32.dll")
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procGetCurrentThread = modkernel32.NewProc("GetCurrentThread")
	procOpenThreadToken  = modadvapi32.NewProc("OpenThreadToken")
	procImpersonateSelf  = modadvapi32.NewProc("ImpersonateSelf")
	procRevertToSelf     = modadvapi32.NewProc("RevertToSelf")
)

// getCurrentThread 获取当前线程的伪句柄
func getCurrentThread() (windows.Handle, error) {
	r0, _, e1 := syscall.Syscall(procGetCurrentThread.Addr(), 0, 0, 0, 0)
	handle := windows.Handle(r0)
	if handle == 0 {
		if e1 != 0 {
			return 0, error(e1)
		}
		return 0, syscall.EINVAL
	}
	return handle, nil
}

// openThreadToken 打开线程令牌
func openThreadToken(thread windows.Handle, access uint32, openAsSelf bool, token *windows.Token) error {
	r0, _, e1 := syscall.Syscall6(procOpenThreadToken.Addr(), 4,
		uintptr(thread), uintptr(access), uintptr(0), uintptr(0), 0, 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	*token = windows.Token(r0)
	return nil
}

// impersonateSelf 模拟自身
func impersonateSelf() error {
	r0, _, e1 := syscall.Syscall(procImpersonateSelf.Addr(), 1, 2, 0, 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}

// revertToSelf 恢复自身
func revertToSelf() error {
	r0, _, e1 := syscall.Syscall(procRevertToSelf.Addr(), 0, 0, 0, 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}

// AddSystemEnvironmentVariable 添加系统级环境变量
func (em *EnvironmentManager) AddSystemEnvironmentVariable(varName, varValue string) error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("无法打开系统环境变量注册表 (需要管理员权限): %v", err)
	}
	defer key.Close()

	var valueType uint32
	if strings.ToUpper(varName) == "PATH" || strings.Contains(varValue, "%") {
		valueType = registry.EXPAND_SZ
	} else {
		valueType = registry.SZ
	}

	// 如果是PATH变量，需要特殊处理
	if strings.ToUpper(varName) == "PATH" {
		var existingPath string
		var readErr error

		existingPath, _, readErr = key.GetStringValue("PATH")
		if readErr != nil && readErr != registry.ErrNotExist {
			return fmt.Errorf("无法读取现有PATH变量: %v", readErr)
		}

		if existingPath != "" {
			pathEntries := strings.Split(existingPath, ";")
			for _, entry := range pathEntries {
				if strings.EqualFold(strings.TrimSpace(entry), strings.TrimSpace(varValue)) {
					return fmt.Errorf("PATH中已存在该路径: %s", varValue)
				}
			}
		}

		if existingPath != "" {
			if !strings.HasSuffix(existingPath, ";") {
				varValue = existingPath + ";" + varValue
			} else {
				varValue = existingPath + varValue
			}
		}
	}

	// 设置注册表值
	if valueType == registry.EXPAND_SZ {
		err = key.SetExpandStringValue(varName, varValue)
	} else {
		err = key.SetStringValue(varName, varValue)
	}

	if err != nil {
		return fmt.Errorf("无法设置环境变量: %v", err)
	}

	// 立即通知系统环境变量已更改
	err = em.broadcastEnvironmentChange()
	if err != nil {
		return fmt.Errorf("环境变量设置成功，但通知系统失败: %v", err)
	}

	return nil
}

// AddPathVariable 专门用于添加PATH环境变量
func (em *EnvironmentManager) AddPathVariable(pathValue string) error {
	pathValue = strings.Trim(pathValue, "\"")

	if !filepath.IsAbs(pathValue) {
		return fmt.Errorf("必须提供绝对路径")
	}

	if strings.HasSuffix(strings.ToLower(pathValue), ".exe") {
		pathValue = filepath.Dir(pathValue)
	}

	return em.AddSystemEnvironmentVariable("PATH", pathValue)
}

// broadcastEnvironmentChange 广播环境变量更改消息
func (em *EnvironmentManager) broadcastEnvironmentChange() error {
	const (
		HWND_BROADCAST   = 0xffff
		WM_SETTINGCHANGE = 0x001A
		SMTO_ABORTIFHUNG = 0x0002
	)

	user32 := windows.NewLazySystemDLL("user32.dll")
	sendMessageTimeoutW := user32.NewProc("SendMessageTimeoutW")

	environmentPtr, _ := syscall.UTF16PtrFromString("Environment")

	ret, _, err := sendMessageTimeoutW.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(environmentPtr)),
		uintptr(SMTO_ABORTIFHUNG),
		uintptr(5000), // 5秒超时
		uintptr(0),
	)

	if ret == 0 {
		return fmt.Errorf("广播环境变量更改失败: %v", err)
	}

	return nil
}

// OpenSystemEnvironmentSettings 打开系统环境变量设置
func (em *EnvironmentManager) OpenSystemEnvironmentSettings() error {
	cmd := exec.Command("rundll32.exe", "sysdm.cpl,EditEnvironmentVariables")
	return cmd.Start()
}

// ValidatePathExists 验证路径是否存在
func (em *EnvironmentManager) ValidatePathExists(path string) bool {
	path = strings.Trim(path, "\"")
	if _, err := windows.GetFileAttributes(windows.StringToUTF16Ptr(path)); err != nil {
		return false
	}
	return true
}

// GetSystemEnvironmentVariable 获取系统环境变量值
func (em *EnvironmentManager) GetSystemEnvironmentVariable(varName string) (string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("无法打开系统环境变量注册表: %v", err)
	}
	defer key.Close()

	value, _, err := key.GetStringValue(varName)
	if err != nil {
		if err == registry.ErrNotExist {
			return "", fmt.Errorf("环境变量不存在: %s", varName)
		}
		return "", fmt.Errorf("无法读取环境变量: %v", err)
	}

	return value, nil
}

// DiagnoseEnvironmentAccess 诊断环境变量访问权限
func (em *EnvironmentManager) DiagnoseEnvironmentAccess() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE)
	if err != nil {
		result["registry_read"] = false
		result["registry_read_error"] = err.Error()
	} else {
		result["registry_read"] = true
		key.Close()
	}

	key, err = registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.SET_VALUE)
	if err != nil {
		result["registry_write"] = false
		result["registry_write_error"] = err.Error()
	} else {
		result["registry_write"] = true
		key.Close()
	}

	key, err = registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.ALL_ACCESS)
	if err != nil {
		result["registry_full"] = false
		result["registry_full_error"] = err.Error()
	} else {
		result["registry_full"] = true
		key.Close()
	}

	pathValue, err := em.GetSystemEnvironmentVariable("PATH")
	if err != nil {
		result["path_read"] = false
		result["path_read_error"] = err.Error()
	} else {
		result["path_read"] = true
		result["path_length"] = len(pathValue)
	}

	return result, nil
}
