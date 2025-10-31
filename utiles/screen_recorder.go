package utiles

import (
	"os/exec"
	"runtime"
	"fmt"
)

// StartScreenRecording 启动屏幕录制 (使用Xbox Game Bar)
func StartScreenRecording() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	// 使用快捷键 Win+Alt+R 来开始录制
	// 这需要用户已经在设置中启用了 Xbox Game Bar
	cmd := exec.Command("powershell", "-Command", `
		Add-Type -AssemblyName System.Windows.Forms
		[System.Windows.Forms.SendKeys]::SendWait("^{%{r}}")
	`)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("无法启动屏幕录制: %v", err)
	}

	return nil
}

// StopScreenRecording 停止屏幕录制 (使用Xbox Game Bar)
func StopScreenRecording() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	// 使用快捷键 Win+Alt+R 来停止录制
	cmd := exec.Command("powershell", "-Command", `
		Add-Type -AssemblyName System.Windows.Forms
		[System.Windows.Forms.SendKeys]::SendWait("^{%{r}}")
	`)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("无法停止屏幕录制: %v", err)
	}

	return nil
}

// StartOBSRecording 启动 OBS 录制（如果已安装）
func StartOBSRecording() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	// 尝试启动 OBS 并开始录制
	// 这需要 OBS 已安装并配置了命令行参数
	cmd := exec.Command("obs64.exe", "--startrecording")
	
	err := cmd.Start()
	if err != nil {
		// 如果找不到 obs64.exe，尝试其他可能的路径
		cmd = exec.Command("C:\\Program Files\\obs-studio\\bin\\64bit\\obs64.exe", "--startrecording")
		err = cmd.Start()
		if err != nil {
			cmd = exec.Command("C:\\Program Files (x86)\\obs-studio\\bin\\32bit\\obs32.exe", "--startrecording")
			err = cmd.Start()
			if err != nil {
				return fmt.Errorf("无法启动 OBS 录制: %v", err)
			}
		}
	}

	return nil
}

// StopOBSRecording 停止 OBS 录制
func StopOBSRecording() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	// 发送停止录制命令
	cmd := exec.Command("obs64.exe", "--stoprecording")
	
	err := cmd.Start()
	if err != nil {
		// 如果找不到 obs64.exe，尝试其他可能的路径
		cmd = exec.Command("C:\\Program Files\\obs-studio\\bin\\64bit\\obs64.exe", "--stoprecording")
		err = cmd.Start()
		if err != nil {
			cmd = exec.Command("C:\\Program Files (x86)\\obs-studio\\bin\\32bit\\obs32.exe", "--stoprecording")
			err = cmd.Start()
			if err != nil {
				return fmt.Errorf("无法停止 OBS 录制: %v", err)
			}
		}
	}

	return nil
}

// StartWindowRecording 启动指定窗口录制（通过OBS）
func StartWindowRecording(windowTitle string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	// 使用OBS录制特定窗口
	// 注意：这要求OBS已经配置好了对应的场景和窗口捕获源
	cmd := exec.Command("obs64.exe", "--startrecording", "--profile", "window_capture_"+windowTitle)
	
	err := cmd.Start()
	if err != nil {
		// 如果找不到 obs64.exe，尝试其他可能的路径
		cmd = exec.Command("C:\\Program Files\\obs-studio\\bin\\64bit\\obs64.exe", "--startrecording", "--profile", "window_capture_"+windowTitle)
		err = cmd.Start()
		if err != nil {
			cmd = exec.Command("C:\\Program Files (x86)\\obs-studio\\bin\\32bit\\obs32.exe", "--startrecording", "--profile", "window_capture_"+windowTitle)
			err = cmd.Start()
			if err != nil {
				return fmt.Errorf("无法启动针对窗口 '%s' 的录制: %v", windowTitle, err)
			}
		}
	}

	return nil
}

// StartSpecificWindowRecording 通过窗口标题启动特定窗口录制
func StartSpecificWindowRecording(windowTitle string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	// 创建PowerShell脚本来查找并录制特定窗口
	script := fmt.Sprintf(`
		Add-Type @"
			using System;
			using System.Runtime.InteropServices;
			public class WindowHelper {
				[DllImport("user32.dll")]
				public static extern IntPtr FindWindow(string lpClassName, string lpWindowName);
				[DllImport("user32.dll")]
				public static extern bool SetForegroundWindow(IntPtr hWnd);
				[DllImport("System.Windows.Forms.dll")]
				public static extern void SendKeys(string keys);
			}
"@
		$hwnd = [WindowHelper]::FindWindow($null, "%s")
		if ($hwnd -ne 0) {
			[WindowHelper]::SetForegroundWindow($hwnd)
			Start-Sleep -Milliseconds 500
			[WindowHelper]::SendKeys("^{%{r}}")
		} else {
			throw "未找到窗口: %s"
		}
	`, windowTitle, windowTitle)

	cmd := exec.Command("powershell", "-Command", script)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("无法启动针对窗口 '%s' 的录制: %v", windowTitle, err)
	}

	return nil
}

// StartWindowRegionRecording 启动指定窗口的指定区域录制（通过OBS）
func StartWindowRegionRecording(windowTitle string, x, y, width, height int) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	// 创建一个自定义场景配置文件来录制指定窗口的指定区域
	// 这需要OBS已安装并且配置了相应的场景
	sceneName := fmt.Sprintf("region_%s_%d_%d_%d_%d", windowTitle, x, y, width, height)
	
	// 使用OBS命令行参数启动录制特定场景
	cmd := exec.Command("obs64.exe", "--startrecording", "--scene", sceneName)
	
	err := cmd.Start()
	if err != nil {
		// 如果找不到 obs64.exe，尝试其他可能的路径
		cmd = exec.Command("C:\\Program Files\\obs-studio\\bin\\64bit\\obs64.exe", "--startrecording", "--scene", sceneName)
		err = cmd.Start()
		if err != nil {
			cmd = exec.Command("C:\\Program Files (x86)\\obs-studio\\bin\\32bit\\obs32.exe", "--startrecording", "--scene", sceneName)
			err = cmd.Start()
			if err != nil {
				return fmt.Errorf("无法启动针对窗口 '%s' 区域的录制: %v", windowTitle, err)
			}
		}
	}

	return nil
}

// StartFullScreenRecording 启动全屏录制（自动检测可用的录制方式）
func StartFullScreenRecording() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	// 首先尝试使用 OBS
	err := StartOBSRecording()
	if err != nil {
		// 如果 OBS 不可用，则尝试使用 Xbox Game Bar
		err = StartScreenRecording()
		if err != nil {
			return fmt.Errorf("无法启动任何屏幕录制程序: %v", err)
		}
	}

	return nil
}

// StopFullScreenRecording 停止全屏录制
func StopFullScreenRecording() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	// 尝试停止 OBS 录制
	err := StopOBSRecording()
	if err != nil {
		// 如果 OBS 停止失败，则尝试停止 Xbox Game Bar 录制
		err = StopScreenRecording()
		if err != nil {
			return fmt.Errorf("无法停止任何屏幕录制程序: %v", err)
		}
	}

	return nil
}