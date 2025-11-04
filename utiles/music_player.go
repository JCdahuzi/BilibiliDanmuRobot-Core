package utiles

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// MusicPlayerType 音乐播放器类型
type MusicPlayerType string

const (
	KuGouMusic MusicPlayerType = "kugou"
	QQMusic    MusicPlayerType = "qqmusic"
)

// PlayerInfo 存储播放器相关信息
type playerInfo struct {
	ProcessNames []string
	WindowNames  []string
	InstallPaths []string
	Executable   string
}

// 播放器配置信息
var playerConfigs = map[MusicPlayerType]playerInfo{
	KuGouMusic: {
		ProcessNames: []string{"KuGou", "kugou"},
		WindowNames:  []string{"酷狗音乐", "KuGou"},
		InstallPaths: []string{
			"C:\\Program Files (x86)\\KuGou\\KuGou.exe",
			"C:\\Program Files\\KuGou\\KuGou.exe",
			"C:\\Users\\" + os.Getenv("USERNAME") + "\\AppData\\Local\\KuGou\\KuGou.exe",
		},
		Executable: "KuGou.exe",
	},
	QQMusic: {
		ProcessNames: []string{"QQMusic", "QQMusicAgent", "QQMusicMiniPlayer", "QQMusicService"},
		WindowNames:  []string{"QQ音乐", "QQMusic", "Tencent QQ Music", "QQ音乐-千万正版音乐海量无损曲库"},
		InstallPaths: []string{
			"C:\\Program Files (x86)\\Tencent\\QQMusic\\QQMusic.exe",
			"C:\\Program Files\\Tencent\\QQMusic\\QQMusic.exe",
			"C:\\Users\\" + os.Getenv("USERNAME") + "\\AppData\\Local\\Programs\\Tencent\\QQMusic\\QQMusic.exe",
		},
		Executable: "QQMusic.exe",
	},
}

// AddSongToPlaylist 将指定名称的歌曲添加到指定音乐播放器的播放列表
func AddSongToPlaylist(player MusicPlayerType, songName string) error {
	if runtime.GOOS != "windows" {
		logx.Errorf("此功能仅支持 Windows 系统")
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	return addSongToPlaylist(player, songName)
}

// SearchAndPlaySong 搜索并播放指定歌曲
func SearchAndPlaySong(player MusicPlayerType, songName string) error {
	if runtime.GOOS != "windows" {
		logx.Errorf("此功能仅支持 Windows 系统")
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	return searchAndPlaySong(player, songName)
}

// PlayMusic 播放音乐
func PlayMusic(player MusicPlayerType) error {
	if runtime.GOOS != "windows" {
		logx.Errorf("此功能仅支持 Windows 系统")
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	return sendMediaKey(" ")
}

// PauseMusic 暂停音乐
func PauseMusic(player MusicPlayerType) error {
	if runtime.GOOS != "windows" {
		logx.Errorf("此功能仅支持 Windows 系统")
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	return sendMediaKey(" ")
}

// NextSong 下一首歌曲
func NextSong(player MusicPlayerType) error {
	if runtime.GOOS != "windows" {
		logx.Errorf("此功能仅支持 Windows 系统")
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	return sendMediaKey("{RIGHT}")
}

// PreviousSong 上一首歌曲
func PreviousSong(player MusicPlayerType) error {
	if runtime.GOOS != "windows" {
		logx.Errorf("此功能仅支持 Windows 系统")
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}

	return sendMediaKey("{LEFT}")
}

// addSongToPlaylist 将指定名称的歌曲添加到播放列表
func addSongToPlaylist(player MusicPlayerType, songName string) error {
	// 检查音乐播放器是否正在运行
	err := checkMusicPlayerRunning(player)
	if err != nil {
		return fmt.Errorf("%s未运行或无法连接: %v", getPlayerName(player), err)
	}

	config, ok := playerConfigs[player]
	if !ok {
		return fmt.Errorf("不支持的音乐播放器类型: %s", player)
	}

	// 查找窗口标题
	var windowTitle string
	for _, title := range config.WindowNames {
		windowTitle = title
		break // 使用第一个标题作为示例
	}

	// 截取窗口截图
	imagePath, err := captureWindow(windowTitle)
	if err != nil {
		// 如果截图失败，回退到启发式方法
		return addSongToPlaylistFallback(player, songName)
	}
	defer os.Remove(imagePath) // 清理临时文件

	// 使用图像分析识别搜索框位置
	x, y, err := findSearchBoxWithImageAnalysis(imagePath)
	if err != nil {
		// 如果图像分析失败，回退到启发式方法
		return addSongToPlaylistFallback(player, songName)
	}

	// 构造PowerShell脚本来添加歌曲到播放列表
	escapedSongName := strings.ReplaceAll(songName, `"`, `""`)
	windowCheck := buildWindowCheckScript(config.WindowNames)

	script := fmt.Sprintf(`
		# 显式导入Windows Forms程序集
		Add-Type -AssemblyName System.Windows.Forms
		Add-Type -AssemblyName System.Drawing
		
		Add-Type @"
			using System;
			using System.Runtime.InteropServices;
			public class WindowHelper {
				[DllImport("user32.dll")]
				public static extern IntPtr FindWindow(string lpClassName, string lpWindowName);
				[DllImport("user32.dll")]
				public static extern bool SetForegroundWindow(IntPtr hWnd);
				[DllImport("user32.dll")]
				public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
				[DllImport("user32.dll")]
				public static extern bool SetWindowPos(IntPtr hWnd, IntPtr hWndInsertAfter, int X, int Y, int cx, int cy, uint uFlags);
				[DllImport("user32.dll")]
				public static extern bool GetWindowRect(IntPtr hWnd, out RECT lpRect);
			}
			public class User32 {
			[DllImport("user32.dll")]
			public static extern void mouse_event(uint dwFlags, uint dx, uint dy, uint dwData, int dwExtraInfo);
			}
			public struct RECT {
				public int Left;
				public int Top;
				public int Right;
				public int Bottom;
			}
"@
		
		# 设置输出编码为UTF-8
		[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
		
		# 设置播放器名称变量
		$playerName = "QQ音乐"
		Write-Output "步骤1: 开始查找$playerName窗口..."
		
		# 查找音乐播放器窗口（已注释掉）
		# %s
		# 直接设置hwnd为非零值以跳过窗口查找步骤
		$hwnd = 1
		Write-Output "步骤2: 跳过窗口查找，直接设置hwnd值为: $hwnd"
		
		if ($hwnd -ne 0) {
			Write-Output "步骤3: 开始窗口前置操作..."
			# 将窗口显示在最前面
			Write-Output "步骤3.1: 恢复窗口..."
			[WindowHelper]::ShowWindow($hwnd, 9)  # SW_RESTORE
			Write-Output "步骤3.2: 设置窗口置顶..."
			[WindowHelper]::SetWindowPos($hwnd, -1, 0, 0, 0, 0, 3)  # HWND_TOPMOST, SWP_NOMOVE | SWP_NOSIZE
			Write-Output "步骤3.3: 设置窗口为前台..."
			[WindowHelper]::SetForegroundWindow($hwnd)
			Write-Output "步骤3.4: 等待窗口激活..."
			Start-Sleep -Milliseconds 1000  # 等待窗口激活
			
			Write-Output "步骤4: 获取窗口位置和大小..."
			Write-Output "[播放功能] 步骤4: 获取窗口位置和大小..."
			# 获取窗口位置和大小
			$rect = New-Object RECT
			[WindowHelper]::GetWindowRect($hwnd, [ref]$rect)
			Write-Output "[播放功能] 步骤4.1: 窗口位置 - 左: $($rect.Left), 上: $($rect.Top), 右: $($rect.Right), 下: $($rect.Bottom)"
			Write-Output "步骤4.1: 窗口位置 - 左: $($rect.Left), 上: $($rect.Top), 右: $($rect.Right), 下: $($rect.Bottom)"
			
			Write-Output "步骤5: 计算搜索框位置..."
			Write-Output "[播放功能] 步骤5: 计算搜索框位置..."
			# 计算搜索框绝对位置
			$searchBoxX = $rect.Left + %d
			$searchBoxY = $rect.Top + %d
			Write-Output "[播放功能] 步骤5.1: 搜索框位置 - X: $searchBoxX, Y: $searchBoxY"
			Write-Output "步骤5.1: 搜索框位置 - X: $searchBoxX, Y: $searchBoxY"
			
			Write-Output "步骤6: 移动鼠标到搜索框并点击..."
			# 移动鼠标到搜索框位置并点击
			[System.Windows.Forms.Cursor]::Position = New-Object System.Drawing.Point($searchBoxX, $searchBoxY)
			Write-Output "步骤6.1: 鼠标已移动到搜索框位置"
			
			# 正确的鼠标点击方法：使用Windows API
			[User32]::mouse_event(0x0002, 0, 0, 0, 0)  # MOUSEEVENTF_LEFTDOWN
			[User32]::mouse_event(0x0004, 0, 0, 0, 0)  # MOUSEEVENTF_LEFTUP
			Write-Output "步骤6.2: 已执行鼠标左键点击"
			Start-Sleep -Milliseconds 500
			
			Write-Output "步骤7: 发送歌曲名称..."
			# 发送歌曲名称
			# 使用变量方式避免%字符被解释为Alt键
			$songName = "%songName%"
			Write-Output "步骤7.1: 发送的歌曲名称: $songName"
			[System.Windows.Forms.SendKeys]::SendWait($songName)
			Write-Output "步骤7.2: 歌曲名称发送完成"
			Start-Sleep -Milliseconds 500
			
			Write-Output "步骤8: 执行搜索..."
			# 回车执行搜索
			[System.Windows.Forms.SendKeys]::SendWait("{ENTER}")
			Write-Output "步骤8.1: 已发送回车键，等待搜索结果..."
			Start-Sleep -Milliseconds 1500
			
			Write-Output "步骤9: 选择并添加歌曲到播放列表..."
			# 选择第一个结果并添加到播放列表（Shift+Enter）
			[System.Windows.Forms.SendKeys]::SendWait("{DOWN}")
			Write-Output "步骤9.1: 已选择第一个结果"
			Start-Sleep -Milliseconds 200
			[System.Windows.Forms.SendKeys]::SendWait("+{ENTER}")
			Write-Output "步骤9.2: 已添加到播放列表"
		} else {
			Write-Output "错误步骤: 窗口查找失败，hwnd为0"
			throw "错误：未找到" + $playerName + "窗口。请确保" + $playerName + "已启动且窗口标题包含" + $playerName + "。" 
		}
	`, windowCheck, x, y, escapedSongName)

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logx.Errorf("添加歌曲到播放列表失败: %v, 输出: %s", err, string(output))
		return fmt.Errorf("添加歌曲到播放列表失败: %v, 输出: %s", err, string(output))
	}

	return nil
}

// addSongToPlaylistFallback 回退方案，使用启发式方法定位搜索框
func addSongToPlaylistFallback(player MusicPlayerType, songName string) error {
	config, ok := playerConfigs[player]
	if !ok {
		return fmt.Errorf("不支持的音乐播放器类型: %s", player)
	}

	// 构造PowerShell脚本来添加歌曲到播放列表
	escapedSongName := strings.ReplaceAll(songName, `"`, `""`)
	windowCheck := buildWindowCheckScript(config.WindowNames)

	script := fmt.Sprintf(`
			# 显式导入Windows Forms程序集
			Add-Type -AssemblyName System.Windows.Forms
			Add-Type -AssemblyName System.Drawing
			
			Add-Type @"
				using System;
				using System.Runtime.InteropServices;
				public class WindowHelper {
				[DllImport("user32.dll")]
				public static extern IntPtr FindWindow(string lpClassName, string lpWindowName);
				[DllImport("user32.dll")]
				public static extern bool SetForegroundWindow(IntPtr hWnd);
				[DllImport("user32.dll")]
				public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
				[DllImport("user32.dll")]
				public static extern bool SetWindowPos(IntPtr hWnd, IntPtr hWndInsertAfter, int X, int Y, int cx, int cy, uint uFlags);
				[DllImport("user32.dll")]
				public static extern bool GetWindowRect(IntPtr hWnd, out RECT lpRect);
			}
			public class User32 {
			[DllImport("user32.dll")]
			public static extern void mouse_event(uint dwFlags, uint dx, uint dy, uint dwData, int dwExtraInfo);
			}
			public struct RECT {
				public int Left;
				public int Top;
				public int Right;
				public int Bottom;
			}
"@
		
		# 设置输出编码为UTF-8
		[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
		
		# 设置播放器名称变量
		$playerName = "QQ音乐"
		Write-Output "[播放功能] 步骤1: 开始查找$playerName窗口..."
		
		# 查找音乐播放器窗口（已注释掉）
		# %s
		# 直接设置hwnd为非零值以跳过窗口查找步骤
		$hwnd = 1
		Write-Output "[播放功能] 步骤2: 跳过窗口查找，直接设置hwnd值为: $hwnd"
		
		if ($hwnd -ne 0) {
			Write-Output "[播放功能] 步骤3: 开始窗口前置操作..."
			# 将窗口显示在最前面
			Write-Output "[播放功能] 步骤3.1: 恢复窗口..."
			[WindowHelper]::ShowWindow($hwnd, 9)  # SW_RESTORE
			Write-Output "[播放功能] 步骤3.2: 设置窗口置顶..."
			[WindowHelper]::SetWindowPos($hwnd, -1, 0, 0, 0, 0, 3)  # HWND_TOPMOST, SWP_NOMOVE | SWP_NOSIZE
			Write-Output "[播放功能] 步骤3.3: 设置窗口为前台..."
			[WindowHelper]::SetForegroundWindow($hwnd)
			Write-Output "[播放功能] 步骤3.4: 等待窗口激活..."
			Start-Sleep -Milliseconds 1000  # 等待窗口激活
			
			# 获取窗口位置和大小
			$rect = New-Object RECT
			[WindowHelper]::GetWindowRect($hwnd, [ref]$rect)
			
			# 计算搜索框绝对位置（启发式方法）
			$searchBoxX = $rect.Left + 300
			$searchBoxY = $rect.Top + 30
			
			Write-Output "[播放功能] 步骤6: 移动鼠标到搜索框并点击..."
			# 移动鼠标到搜索框位置并点击
			[System.Windows.Forms.Cursor]::Position = New-Object System.Drawing.Point($searchBoxX, $searchBoxY)
			Write-Output "[播放功能] 步骤6.1: 鼠标已移动到搜索框位置"
			# 正确的鼠标点击方法：使用Windows API
			[User32]::mouse_event(0x0002, 0, 0, 0, 0)  # MOUSEEVENTF_LEFTDOWN
			[User32]::mouse_event(0x0004, 0, 0, 0, 0)  # MOUSEEVENTF_LEFTUP
			Write-Output "[播放功能] 步骤6.2: 已执行鼠标左键点击"
			Start-Sleep -Milliseconds 500
			
			Write-Output "[播放功能] 步骤7: 发送歌曲名称..."
			# 发送歌曲名称
			# 使用变量方式避免%字符被解释为Alt键
			$songName = "%songName%"
			Write-Output "[播放功能] 步骤7.1: 发送的歌曲名称: $songName"
			[System.Windows.Forms.SendKeys]::SendWait($songName)
			Write-Output "[播放功能] 步骤7.2: 歌曲名称发送完成"
			Start-Sleep -Milliseconds 500
			
			# 回车执行搜索
			[System.Windows.Forms.SendKeys]::SendWait("{ENTER}")
			Start-Sleep -Milliseconds 1500
			
			# 选择第一个结果并添加到播放列表（Shift+Enter）
			[System.Windows.Forms.SendKeys]::SendWait("{DOWN}")
			Start-Sleep -Milliseconds 200
			[System.Windows.Forms.SendKeys]::SendWait("+{ENTER}")
		} else {
			Write-Output "[播放功能] 错误步骤: 窗口查找失败，hwnd为0"
			throw "错误：未找到" + $playerName + "窗口。请确保" + $playerName + "已启动且窗口标题包含" + $playerName + "。" 
		}
	`, windowCheck, escapedSongName)

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logx.Errorf("添加歌曲到播放列表失败: %v, 输出: %s", err, string(output))
		return fmt.Errorf("添加歌曲到播放列表失败: %v, 输出: %s", err, string(output))
	}

	return nil
}

// searchAndPlaySong 搜索并播放指定歌曲
func searchAndPlaySong(player MusicPlayerType, songName string) error {
	err := checkMusicPlayerRunning(player)
	if err != nil {
		return err
	}

	config, ok := playerConfigs[player]
	if !ok {
		return fmt.Errorf("不支持的音乐播放器类型: %s", player)
	}

	// 查找窗口标题
	var windowTitle string
	for _, title := range config.WindowNames {
		windowTitle = title
		break // 使用第一个标题作为示例
	}

	// 截取窗口截图
	imagePath, err := captureWindow(windowTitle)
	if err != nil {
		logx.Errorf("截图失败: %v", err)
		return fmt.Errorf("截图失败: %v", err)
	}
	defer os.Remove(imagePath) // 清理临时文件

	// 使用图像分析识别搜索框位置
	x, y, err := findSearchBoxWithImageAnalysis(imagePath)
	if err != nil {
		logx.Errorf("图像分析搜索框失败: %v", err)
		return fmt.Errorf("图像分析搜索框失败: %v", err)
	}

	// 构造PowerShell脚本来搜索和播放歌曲
	escapedSongName := strings.ReplaceAll(songName, `"`, `""`)
	windowCheck := buildWindowCheckScript(config.WindowNames)

	script := fmt.Sprintf(`
			# 显式导入Windows Forms程序集
			Add-Type -AssemblyName System.Windows.Forms
			Add-Type -AssemblyName System.Drawing
			
			Add-Type @"
				using System;
				using System.Runtime.InteropServices;
				public class WindowHelper {
					[DllImport("user32.dll")]
					public static extern IntPtr FindWindow(string lpClassName, string lpWindowName);
					[DllImport("user32.dll")]
					public static extern bool SetForegroundWindow(IntPtr hWnd);
					[DllImport("user32.dll")]
					public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
					[DllImport("user32.dll")]
					public static extern bool SetWindowPos(IntPtr hWnd, IntPtr hWndInsertAfter, int X, int Y, int cx, int cy, uint uFlags);
					[DllImport("user32.dll")]
					public static extern bool GetWindowRect(IntPtr hWnd, out RECT lpRect);
				}
				public class User32 {
				[DllImport("user32.dll")]
				public static extern void mouse_event(uint dwFlags, uint dx, uint dy, uint dwData, int dwExtraInfo);
				}
				public struct RECT {
					public int Left;
					public int Top;
					public int Right;
					public int Bottom;
				}
			"@
		
		# 设置输出编码为UTF-8
		[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
		
		# 设置播放器名称变量
		$playerName = "QQ音乐"
		
		# 查找音乐播放器窗口（已注释掉）
		# %s
		# 直接设置hwnd为非零值以跳过窗口查找步骤
		$hwnd = 1
		Write-Output "跳过窗口查找，直接设置hwnd值为: $hwnd"
		
		if ($hwnd -ne 0) {
			# 将窗口显示在最前面
			[WindowHelper]::ShowWindow($hwnd, 9)  # SW_RESTORE
			[WindowHelper]::SetWindowPos($hwnd, -1, 0, 0, 0, 0, 3)  # HWND_TOPMOST, SWP_NOMOVE | SWP_NOSIZE
			[WindowHelper]::SetForegroundWindow($hwnd)
			Start-Sleep -Milliseconds 1000  # 等待窗口激活
			
			# 获取窗口位置和大小
			$rect = New-Object RECT
			[WindowHelper]::GetWindowRect($hwnd, [ref]$rect)
			
			# 计算搜索框绝对位置
			$searchBoxX = $rect.Left + %d
			$searchBoxY = $rect.Top + %d
			
			# 移动鼠标到搜索框位置并点击
			[System.Windows.Forms.Cursor]::Position = New-Object System.Drawing.Point($searchBoxX, $searchBoxY)
			# 正确的鼠标点击方法：使用Windows API
			[User32]::mouse_event(0x0002, 0, 0, 0, 0)  # MOUSEEVENTF_LEFTDOWN
			[User32]::mouse_event(0x0004, 0, 0, 0, 0)  # MOUSEEVENTF_LEFTUP
			Start-Sleep -Milliseconds 500
			
			# 发送歌曲名称
			# 使用变量方式避免%字符被解释为Alt键
			$songName = "%songName%"
			[System.Windows.Forms.SendKeys]::SendWait($songName)
			Start-Sleep -Milliseconds 500
			
			# 回车执行搜索
			[System.Windows.Forms.SendKeys]::SendWait("{ENTER}")
			Start-Sleep -Milliseconds 1500
			
			# 选择第一个结果并播放
			[System.Windows.Forms.SendKeys]::SendWait("{DOWN}")
			Start-Sleep -Milliseconds 200
			[System.Windows.Forms.SendKeys]::SendWait("{ENTER}")
		} else {
			throw "错误：未找到" + $playerName + "窗口。请确保" + $playerName + "已启动且窗口标题包含" + $playerName + "。" 
		}
	`, windowCheck, x, y, escapedSongName)

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logx.Errorf("搜索并播放歌曲失败: %v, 输出: %s", err, string(output))
		return fmt.Errorf("搜索并播放歌曲失败: %v, 输出: %s", err, string(output))
	}

	return nil
}

// sendMediaKey 发送媒体控制键
func sendMediaKey(key string) error {
	script := fmt.Sprintf(`
		Add-Type -AssemblyName System.Windows.Forms
		[System.Windows.Forms.SendKeys]::SendWait("%s")
	`, key)

	cmd := exec.Command("powershell", "-Command", script)
	err := cmd.Run()
	if err != nil {
		logx.Errorf("无法发送媒体控制键: %v", err)
		return fmt.Errorf("无法发送媒体控制键: %v", err)
	}

	return nil
}

// checkMusicPlayerRunning 检查音乐播放器是否正在运行
func checkMusicPlayerRunning(player MusicPlayerType) error {
	config, ok := playerConfigs[player]
	if !ok {
		return fmt.Errorf("不支持的音乐播放器类型: %s", player)
	}

	// 构建检查进程的脚本 - 修改为只要检测到任何一个进程存在就返回成功
	var processCheckScript string
	for i, processName := range config.ProcessNames {
		if i == 0 {
			processCheckScript = fmt.Sprintf(`$processFound = $false
		$process = Get-Process -Name "%s" -ErrorAction SilentlyContinue
		if ($process -ne $null) { $processFound = $true }`, processName)
		} else {
			processCheckScript += fmt.Sprintf(`
		if (-not $processFound) {
			$process = Get-Process -Name "%s" -ErrorAction SilentlyContinue
			if ($process -ne $null) { $processFound = $true }
		}`, processName)
		}
	}

	script := fmt.Sprintf(`
		# 设置输出编码为UTF-8
		[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
		
		%s
		if (-not $processFound) {
			throw "%s未运行"
		}
	`, processCheckScript, getPlayerName(player))

	cmd := exec.Command("powershell", "-Command", script)
	err := cmd.Run()
	if err != nil {
		// 尝试启动音乐播放器
		startErr := startMusicPlayer(player)
		if startErr != nil {
		logx.Errorf("%s未运行且无法启动: %v", getPlayerName(player), startErr)
		return fmt.Errorf("%s未运行且无法启动: %v", getPlayerName(player), startErr)
	}
		// 给播放器一些启动时间
		time.Sleep(3 * time.Second)
		// 再次检查进程是否启动成功
		cmd = exec.Command("powershell", "-Command", script)
		err = cmd.Run()
		if err != nil {
		logx.Errorf("%s启动后仍未检测到进程", getPlayerName(player))
		return fmt.Errorf("%s启动后仍未检测到进程", getPlayerName(player))
	}
	}

	return nil
}

// startMusicPlayer 启动音乐播放器
func startMusicPlayer(player MusicPlayerType) error {
	config, ok := playerConfigs[player]
	if !ok {
		return fmt.Errorf("不支持的音乐播放器类型: %s", player)
	}

	// 尝试多种可能的安装路径
	for _, path := range config.InstallPaths {
		// 检查文件是否存在
		if _, err := os.Stat(path); err == nil {
			cmd := exec.Command(path)
			err := cmd.Start()
			if err == nil {
				return nil
			}
		}
	}

	// 如果所有路径都失败，则尝试通过系统PATH查找
	cmd := exec.Command(config.Executable)
	err := cmd.Start()
	if err != nil {
		logx.Errorf("无法启动音乐播放器: %v", err)
		return fmt.Errorf("无法启动音乐播放器: %v", err)
	}

	return nil
}

// buildWindowCheckScript 构建窗口检查脚本
func buildWindowCheckScript(windowNames []string) string {
	if len(windowNames) == 0 {
		return "$hwnd = 0"
	}

	// 构建更灵活的窗口查找逻辑，支持部分匹配
	script := fmt.Sprintf(`Write-Output "[窗口查找] 尝试通过精确标题查找窗口..."
		$windowNamesToTry = @("` + strings.Join(windowNames, `", "`) + `")
 	 	Write-Output "[窗口查找] 尝试的窗口标题列表:  $ ($windowNamesToTry -join ', ')"
		$hwnd = [WindowHelper]::FindWindow($null, "%s")
		Write-Output "[窗口查找] 尝试标题 '%s'，hwnd = $hwnd"`, windowNames[0], windowNames[0])
	for i := 1; i < len(windowNames); i++ {
		script += fmt.Sprintf(`
		if ($hwnd -eq 0) {
			Write-Output "[窗口查找] 尝试标题 '%s'..."
			$hwnd = [WindowHelper]::FindWindow($null, "%s")
			Write-Output "[窗口查找] hwnd = $hwnd"
		}`, windowNames[i], windowNames[i])
	}

	// 添加通过枚举窗口来查找的备用方法
	script += `
		if ($hwnd -eq 0) {
			Write-Output "[窗口查找] 精确匹配失败，尝试通过枚举窗口和部分匹配查找..."
			# 如果通过精确标题找不到，尝试通过枚举窗口来查找
			Add-Type @"
				using System;
				using System.Text;
				using System.Runtime.InteropServices;
				public class WindowEnumerator {
					[DllImport("user32.dll")]
					public static extern bool EnumWindows(EnumWindowsProc lpEnumFunc, IntPtr lParam);
					[DllImport("user32.dll", CharSet = CharSet.Unicode)]
					public static extern int GetWindowText(IntPtr hWnd, StringBuilder strText, int maxCount);
					[DllImport("user32.dll")]
					public static extern int GetWindowTextLength(IntPtr hWnd);
				}
				public delegate bool EnumWindowsProc(IntPtr hWnd, IntPtr lParam);
"@
			
			$hwnds = @()
			$windowNames = @("` + strings.Join(windowNames, `", "`) + `")
			$allWindows = @()
			
			Write-Output "[窗口查找] 开始枚举所有窗口..."
			$enumWindowsProc = {
				param([IntPtr]$hWnd, [IntPtr]$lParam)
				
				$length = [WindowEnumerator]::GetWindowTextLength($hWnd)
				if ($length -gt 0) {
					$builder = New-Object System.Text.StringBuilder ($length + 1)
					[WindowEnumerator]::GetWindowText($hWnd, $builder, $builder.Capacity)
					$windowTitle = $builder.ToString()
					
					# 记录所有找到的窗口标题（用于调试）
					$allWindows += $windowTitle
					
					# 检查是否匹配目标窗口
					foreach ($windowName in $windowNames) {
						if ($windowTitle -like "*$windowName*") {
							Write-Output "[窗口查找] 找到匹配窗口: '$windowTitle' 匹配 '$windowName'"
							$hwnds += $hWnd
							break
						}
					}
				}
				return $true
			}
			
			[WindowEnumerator]::EnumWindows($enumWindowsProc, [IntPtr]::Zero)
			Write-Output "[窗口查找] 枚举完成，找到 $($allWindows.Count) 个窗口"
			# 输出部分窗口标题用于调试（避免输出过多）
			Write-Output "[窗口查找] 前10个窗口标题: $($allWindows | Select-Object -First 10)"
			
			if ($hwnds.Count -gt 0) {
				Write-Output "[窗口查找] 通过枚举找到 $($hwnds.Count) 个匹配窗口，使用第一个"
				$hwnd = $hwnds[0]
			} else {
				Write-Output "[窗口查找] 枚举窗口后仍未找到匹配窗口"
			}
		}`

	script += "\n\t\tWrite-Output '[窗口查找] 最终hwnd值: $hwnd'\n\t"
	return script
}

// getPlayerName 获取播放器名称
func getPlayerName(player MusicPlayerType) string {
	switch player {
	case KuGouMusic:
		return "酷狗音乐"
	case QQMusic:
		return "QQ音乐"
	default:
		return "未知播放器"
	}
}

// GetRunningPlayerNames 获取当前运行的音乐播放器名称列表
func GetRunningPlayerNames() []string {
	var runningPlayers []string

	// 检查各个播放器是否在运行
	for player, config := range playerConfigs {
		for _, processName := range config.ProcessNames {
			cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq "+processName+".exe")
			output, err := cmd.Output()
			if err == nil && strings.Contains(string(output), processName) {
				runningPlayers = append(runningPlayers, string(getPlayerName(player)))
				break
			}
		}
	}

	return runningPlayers
}

// captureWindow 截取指定窗口的屏幕截图
func captureWindow(windowTitle string) (string, error) {
	// 创建临时文件路径存储截图，使用绝对路径并替换特殊字符
	sanitizedTitle := strings.ReplaceAll(strings.ReplaceAll(windowTitle, "\\", "_"), "/", "_")
	tempFile := fmt.Sprintf("%s/temp_%s.png", os.TempDir(), sanitizedTitle)

	script := fmt.Sprintf(`
		# 添加DPI感知设置以处理高DPI显示器
		Add-Type -TypeDefinition @"
		using System;
		using System.Runtime.InteropServices;
		using System.Drawing;
		using System.Drawing.Imaging;

		public class WindowCapture {
			// DPI感知设置
			[DllImport("user32.dll")]
			public static extern bool SetProcessDPIAware();

			// 窗口操作函数
			[DllImport("user32.dll")]
			public static extern IntPtr FindWindow(string lpClassName, string lpWindowName);
			[DllImport("user32.dll")]
			public static extern bool GetWindowRect(IntPtr hWnd, out RECT lpRect);
			[DllImport("user32.dll")]
			public static extern bool PrintWindow(IntPtr hWnd, IntPtr hdcBlt, int nFlags);
			[DllImport("user32.dll")]
			public static extern IntPtr GetWindowDC(IntPtr hWnd);
			[DllImport("user32.dll")]
			public static extern int ReleaseDC(IntPtr hWnd, IntPtr hDC);
		}

		public struct RECT {
			public int Left;
			public int Top;
			public int Right;
			public int Bottom;
		}
"@

		# 设置DPI感知
		[WindowCapture]::SetProcessDPIAware()

		# 查找窗口
		$hwnd = [WindowCapture]::FindWindow($null, "%s")
		if ($hwnd -eq 0) {
			throw "未找到窗口: %s"
		}

		# 获取窗口位置和大小
		$rect = New-Object RECT
		if (-not [WindowCapture]::GetWindowRect($hwnd, [ref]$rect)) {
			throw "获取窗口大小失败"
		}

		$width = $rect.Right - $rect.Left
		$height = $rect.Bottom - $rect.Top

		# 确保窗口尺寸有效
		if ($width -le 0 -or $height -le 0) {
			throw "无效的窗口尺寸: ${width}x${height}"
		}

		try {
			# 创建位图
			$bitmap = New-Object System.Drawing.Bitmap($width, $height)
			$graphics = [System.Drawing.Graphics]::FromImage($bitmap)
			$hdc = $graphics.GetHdc()

			# 使用PrintWindow方法（更可靠）
			$result = [WindowCapture]::PrintWindow($hwnd, $hdc, 0)
			
			# 如果PrintWindow失败，尝试使用CopyFromScreen作为后备方案
			if (-not $result) {
				Write-Host "PrintWindow失败，尝试使用CopyFromScreen作为后备方案"
				$graphics.CopyFromScreen($rect.Left, $rect.Top, 0, 0, $bitmap.Size)
			}

			$graphics.ReleaseHdc($hdc)

			# 确保目录存在
			$dir = Split-Path -Parent "%s"
			if (-not (Test-Path $dir)) {
				New-Item -ItemType Directory -Force -Path $dir | Out-Null
			}

			# 保存图片
			$bitmap.Save("%s")
			Write-Host "截图成功保存到: %s"
		} catch {
			throw "截图过程中发生错误: $_"
		} finally {
			# 确保资源释放
			if ($graphics -ne $null) { $graphics.Dispose() }
			if ($bitmap -ne $null) { $bitmap.Dispose() }
		}
	`, windowTitle, windowTitle, tempFile, tempFile, tempFile)

	cmd := exec.Command("powershell", "-Command", script)
	err := cmd.Run()
	if err != nil {
		logx.Errorf("截图失败: %v", err)
		return "", fmt.Errorf("截图失败: %v", err)
	}

	return tempFile, nil
}

// findSearchBoxWithImageAnalysis 使用图像分析识别搜索框位置
// 使用OCR技术实现搜索框定位
func findSearchBoxWithImageAnalysis(imagePath string) (int, int, error) {
	// 打开图像文件
	file, err := os.Open(imagePath)
	if err != nil {
		logx.Errorf("无法打开图像文件: %v", err)
		return 0, 0, fmt.Errorf("无法打开图像文件: %v", err)
	}
	defer file.Close()

	// 解码图像
	img, _, err := image.Decode(file)
	if err != nil {
		logx.Errorf("无法解码图像: %v", err)
		return 0, 0, fmt.Errorf("无法解码图像: %v", err)
	}

	// 获取图像边界
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y
	
	// 创建一个基于OCR的搜索框定位函数
	ocrSearchBoxX,	 ocrSearchBoxY, err := locateSearchBoxWithOCR(imagePath, width, height)
	if err == nil && ocrSearchBoxX > 0 && ocrSearchBoxY > 0 {
		// OCR成功定位到搜索框
		logx.Info("使用OCR技术成功定位搜索框")
		return ocrSearchBoxX, ocrSearchBoxY, nil
	}
	
	// OCR定位失败，回退到传统的图像分析方法
	logx.Info("OCR定位失败，回退到传统图像分析方法")
	
	// QQ音乐搜索框特征优化：
	// 1. 搜索框位于顶部	导航栏，通常在25-60像素高度范围内
	// 2. 宽度大约占屏幕的1/3到1/2
	// 3. 具有特定的浅灰色背景
	// 4. 左侧有搜索图标

	searchMinY := 20
	searchMaxY := 70
	searchMinX := width / 4
	searchMaxX := 3 * width / 4

	var bestX, bestY int = -1, -1
	maxScore := 0

	// 在更精确的区域内扫描可能的搜索框位置
	for y := searchMinY; y < searchMaxY; y += 3 {
		for x := searchMinX; x < searchMaxX; x += 5 {
			// 检查当前位置是否可能是搜索框的一部分
			score := analyzeQQMusicSearchBox(img, x, y, bounds, width, height)
			if score > maxScore {
				maxScore = score
				bestX = x
				bestY = y
			}
		}
	}

	// 如果找到可能的搜索框
	if bestX != -1 && bestY != -1 {
		// 确保点击位置在搜索框的中心区域
		adjustedX := bestX
		adjustedY := bestY
		return adjustedX, adjustedY, nil
	}

	// 根据QQ音乐界面布局，使用更精确的默认位置
	// 搜索框通常位于顶部导航栏中间位置
	defaultY := 38 // 更符合QQ音乐搜索框的垂直位置
	defaultX := - 70 // 调整为更居中的位置，不再偏左
	return defaultX, defaultY, nil
}

// analyzeQQMusicSearchBox 针对QQ音乐界面优化的搜索框分析函数
func analyzeQQMusicSearchBox(img image.Image, x, y int, bounds image.Rectangle, width, height int) int {
	// 检查边界
	if x < width/5 || y < 15 || x > 4*width/5 || y > 80 {
		return 0
	}

	score := 0

	// 获取中心点颜色
	centerColor := img.At(x, y)

	// 检查是否为QQ音乐搜索框特有的浅灰色背景
	r, g, b, _ := centerColor.RGBA()
	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)

	// QQ音乐搜索框背景色特征检测
	// 通常是浅灰色，但不是纯白
	luminance := 0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8)
	if luminance >= 180 && luminance <= 230 {
		// 检查是否为灰色（RGB值相近）
		maxRGB := max(max(r8, g8), b8)
		minRGB := min(min(r8, g8), b8)
		if maxRGB-minRGB <= 20 {
			score += 50 // 高分数，因为这是QQ音乐搜索框的主要特征
		}
	}

	// 检查搜索框的矩形形状特征
	// 检查水平方向的一致性（搜索框通常很宽）
	horizontalConsistent := true
	for i := -50; i <= 50; i++ {
		if x+i >= 0 && x+i < bounds.Max.X {
			neighborColor := img.At(x+i, y)
			if isColorDifferent(centerColor, neighborColor) {
				horizontalConsistent = false
				break
			}
		}
	}

	if horizontalConsistent {
		score += 30
	}

	// 检查垂直方向的一致性（搜索框高度固定）
	verticalConsistent := true
	for i := -5; i <= 5; i++ {
		if y+i >= 0 && y+i < bounds.Max.Y {
			neighborColor := img.At(x, y+i)
			if isColorDifferent(centerColor, neighborColor) {
				verticalConsistent = false
				break
			}
		}
	}

	if verticalConsistent {
		score += 20
	}

	// 检查上方是否为导航栏蓝色区域
	if y-20 >= 0 {
		topColor := img.At(x, y-20)
		topR, topG, topB, _ := topColor.RGBA()
		topR8 := uint8(topR >> 8)
		topG8 := uint8(topG >> 8)
		topB8 := uint8(topB >> 8)
		
		// QQ音乐顶部导航栏通常是蓝色调
		totalBrightness := int(topR8) + int(topG8) + int(topB8)
		if topB8 > topR8+30 && topB8 > topG8+30 && totalBrightness > 400 {
			score += 15
		}
	}

	// 检查左侧是否有搜索图标（暗灰色）
	if x-30 >= 0 {
		leftColor := img.At(x-30, y)
		leftR, leftG, leftB, _ := leftColor.RGBA()
		leftR8 := uint8(leftR >> 8)
		leftG8 := uint8(leftG >> 8)
		leftB8 := uint8(leftB >> 8)
		
		// 搜索图标通常是暗灰色
		if leftR8 >= 100 		&& leftR8 <= 150 && leftG8 >= 100 && leftG8 <= 150 && leftB8 >= 100 && leftB8 <= 150 {
			score += 15
		}
	}
	
	// 新增：检查右侧是否有搜索按钮（通常是蓝色），这有助于更准确地	定位搜索框中心
	if x+150 < bounds.Max.X { // 搜索框右侧大约150像素的位置可能有搜索按钮
		rightColor := img.At(x+150, y)
		rightR, rightG, rightB, _ := rightColor.RGBA()
		rightR8 := uint8(rightR >> 8)
		rightG8 := uint8(rightG >> 8)
		rightB8 := uint8(rightB >> 8)
		
		// 搜索按钮通常是蓝色或带有蓝色特征
		if rightR8 >= rightG8 && rightB8 > rightR8 && rightB8 > rightG8 && rightB8 > 150 {
			score += 25 // 给右侧搜索按钮更高的权重，帮助定位搜索框中心
		}
	}
	
	// 新增：位置评分优化 - 更倾向于选择搜索框的中心位置而	不是左侧
	// 计算当前位置与窗口中心的距离
	windowCenterX := width / 2
	distanceToCenter := abs(x - windowCenterX)
	// 距离中心越近，额外加分越多
	if distanceToCenter < 100 {
		score += (100 - distanceToCenter) / 10 // 最多加10分
	}


	return score
}

// 辅助函数：获取三个值中的最大值
func max(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

// 辅助函数：获取三个值中的最小值
func min(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}

// analyzeSearchBoxArea 分析指定区域是否可能是搜索框（保留原函数兼容性）
func analyzeSearchBoxArea(img image.Image, x, y int, bounds image.Rectangle) int {
	// 为了兼容性，调用新的优化函数
	width := bounds.Max.X
	height := bounds.Max.Y
	return analyzeQQMusicSearchBox(img, x, y, bounds, width, height)
}

// isLightColor 判断颜色是否为浅色
func isLightColor(r, g, b uint8) bool {
	// 计算亮度: (0.299*R + 0.587*G + 0.114*B)
	luminance := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	return luminance > 200 // 阈值可根据需要调整
}

// isColorDifferent 判断两个颜色是否不同
func isColorDifferent(c1, c2 color.Color) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	diff := abs(int(r1>>8)-int(r2>>8)) +
		abs(int(g1>>8)-int(g2>>8)) +
		abs(int(b1>>8)-int(b2>>8))

	return diff > 50 // 阈值可根据需要调整
}

// isColorVeryDifferent 判断两个颜色是否明显不同
func isColorVeryDifferent(c1, c2 color.Color) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	diff := abs(int(r1>>8)-int(r2>>8)) +
		abs(int(g1>>8)-int(g2>>8)) +
		abs(int(b1>>8)-int(b2>>8))

	return diff > 100 // 高阈值
}

// abs 计算绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// locateSearchBoxWithOCR 使用OCR技术定位搜索框
// 注意：此函数需要安装Tesseract OCR和gosseract库
func locateSearchBoxWithOCR(imagePath string, width, height int) (int, int, error) {
	// 方法1: 使用exec调用tesseract命令行工具（不需要额外Go依赖）
	// 这里采用命令行方式以避免引入新的依赖
	
	// 创建临时文件保存OCR结果
	resultFile	, err := os.CreateTemp("", "ocr_result_")
	if err != nil {
		logx.Errorf("创建临时文件失败: %v", err)
		return 0, 0, fmt.Errorf("创建临时文件失败: %v", err)
	}
	resultPath := resultFile.Name()
	resultFile.Close()
	defer os.Remove(resultPath)
	
	// 调用tesseract命令行工具，提取坐标信息
		cmd := exec.Command("tesseract", imagePath, resultPath[:len(resultPath)-4], "makebox")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logx.Errorf("Tesseract OCR执行失败: %v, 输出: %s", err, string(output))
		return 0, 0, fmt.Errorf("Tesseract OCR执行失败: %v, 输出: %s", err, string(output))
	}
	
	// 读取tesseract生成的box文件
	boxData, err := os.ReadFile(resultPath + ".box")
	if err != nil {
		logx.Errorf("读取OCR结果文件失败: %v", err)
		return 0, 0, fmt.Errorf("读取OCR结果文件失败: %v", err)
	}
	
	// 解析box文件，寻找搜索相关文字
	lineLines := strings.Split(string(boxData), "\n")
	searchKeywords := []string{"搜索", "search", "查找", "find", "搜索框", "搜索栏"}
	
	var searchBoxX, searchBoxY int
	bestMatchScore := 0
	
	for _, line := range lineLines {
		if line == "" {
			continue
		}
		
		// 解析box文件格式: 字符 x y		 width height
		parts := strings.Fields(line)
		if len(parts) < 6 {
			continue
		}
		
		char := parts[0]
				// 检查是否包含搜索关键词
		for _, keyword := range searchKeywords {
			if strings.Contains(strings.ToLower(keyword), strings.ToLower(char)) {
				// 尝试解析坐标
				x, errX := strconv.Atoi(parts[1])
				y, errY := strconv.Atoi(parts[2])
				boxWidth, errW := strconv.Atoi(parts[3])
				boxHeight, errH := strconv.Atoi(parts[4])
				
				if errX == nil && errY == nil && errW == nil && errH == nil {
					// 计算搜索框中心位置（在识别到的文字下方或旁边）
					// 假设搜索框在搜索文字右侧或下方
					candidateX := x + boxWidth + 20 // 右侧20像素
					candidateY := y + boxHeight/2
					
					// 验证坐标是否在有效范围内
					if candidateX > 0 && candidateX < width && candidateY > 0 && candidateY < height {
						// 计算匹配分数（根据关键词匹配长度和位置合理性）
						score := len(keyword)
						// 优先选择顶部区域的匹配
						if candidateY < height/4 {
							score += 10
						}
						// 优先选择中间区域的匹配
						if candidateX > width/4 && candidateX < 3*width/4 {
							score += 5
						}
						
												if score > bestMatchScore {
							bestMatchScore = score
							searchBoxX = candidateX
							searchBoxY = candidateY
						}
					}
				}
			}
		}
	}
	
	// 	如果找到了搜索框位置
	if bestMatchScore > 0 {
		return searchBoxX, searchBoxY, nil
	}
	
	logx.Errorf("OCR未能识别到搜索框")
	return 0, 0, fmt.Errorf("OCR未能识别到搜索框")
}
