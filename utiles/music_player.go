package utiles

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
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

// PlaySongByBrowser 通过浏览器直接播放指定歌曲
func PlaySongByBrowser(songName string, artist string) error {
	// 根据是否提供歌手信息构建搜索关键词
	searchKeyword := songName
	if artist != "" {
		// 当有多个歌手时，将逗号替换为%20
		if strings.Contains(artist, ", ") || strings.Contains(artist, "， ") {
			artist = strings.ReplaceAll(artist, ", ", "%20")
			artist = strings.ReplaceAll(artist, "， ", "%20")
		}
		searchKeyword = fmt.Sprintf("%s%%20%s", songName, artist)
	}

	// 使用QQ音乐网页版的搜索结果页面
	logx.Infof("打开QQ音乐网页版搜索结果页面: %s", searchKeyword)
	// 构建QQ音乐网页版搜索URL
	webSearchURL := fmt.Sprintf("https://y.qq.com/n/ryqq/search?w=%s", searchKeyword)
	return openURLInBrowser(webSearchURL, searchKeyword)
}

// openURLInBrowser 在浏览器中打开指定的URL
func openURLInBrowser(url, searchKeyword string) error {
	// 通过浏览器打开链接
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	case "linux":
		// 尝试使用常见的浏览器命令
		browsers := []string{"xdg-open", "google-chrome", "firefox", "opera", "safari"}
		for _, browser := range browsers {
			if _, err := exec.LookPath(browser); err == nil {
				cmd = exec.Command(browser, url)
				break
			}
		}
		// 如果没有找到浏览器，使用默认的xdg-open
		if cmd == nil {
			cmd = exec.Command("xdg-open", url)
		}
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	// 执行命令打开浏览器
	if err := cmd.Run(); err != nil {
		logx.Errorf("打开浏览器失败: %v", err)
		return fmt.Errorf("打开浏览器失败: %v", err)
	}

	logx.Infof("已在浏览器中打开歌曲: %s", searchKeyword)
	return nil
}

// AddSongToPlaylist 将指定名称的歌曲添加到指定音乐播放器的播放列表
func AddSongToPlaylist(player MusicPlayerType, songName string) error {
	// 对于QQ音乐，使用API方式添加歌曲到播放列表
	if player == QQMusic {
		// 使用QQ音乐API搜索歌曲
		api := NewQQMusicAPI()
		logx.Infof("正在搜索歌曲: %s", songName)

		// 搜索歌曲
		songs, err := api.SearchSong(songName, 1, 5)
		if err != nil {
			logx.Errorf("搜索歌曲失败: %v", err)
			return fmt.Errorf("搜索歌曲失败: %v", err)
		}

		if len(songs) == 0 {
			logx.Errorf("未找到歌曲: %s", songName)
			return fmt.Errorf("未找到歌曲: %s", songName)
		}

		// 选择搜索结果中的第一首歌曲
		targetSong := songs[0]
		logx.Infof("找到歌曲: %s - %s (专辑: %s)", targetSong.Name, targetSong.Singer, targetSong.Album)

		// 使用本地播放列表功能（如果需要在线播放列表，需要用户认证）
		// 首先检查是否有名为"机器人播放列表"的本地播放列表
		var playlist *LocalPlaylist
		allPlaylists := GetAllLocalPlaylists()
		playlistExists := false

		for _, pl := range allPlaylists {
			if pl.Name == "机器人播放列表" {
				playlist = pl
				playlistExists = true
				break
			}
		}

		// 如果不存在，则创建一个新的播放列表
		if !playlistExists {
			logx.Infof("创建新的本地播放列表: 机器人播放列表")
			playlist = CreateLocalPlaylist("机器人播放列表", "由机器人自动添加的歌曲集合")
		}

		// 将歌曲添加到本地播放列表
		err = AddSongToLocalPlaylist(playlist.ID, targetSong)
		if err != nil {
			logx.Errorf("添加歌曲到本地播放列表失败: %v", err)
			return fmt.Errorf("添加歌曲到本地播放列表失败: %v", err)
		}

		logx.Infof("成功添加歌曲到本地播放列表 '%s'", playlist.Name)

		// 开启播放列表播放线程（如果未开启）
		if !playlistPlaybackRunning {
			logx.Infof("开启播放列表播放线程")
			err := StartPlaylistPlayback(QQMusic)
			if err != nil {
				logx.Errorf("开启播放列表播放线程失败: %v", err)
				return fmt.Errorf("开启播放列表播放线程失败: %v", err)
			}
		}

		return nil
	}

	// 对于其他播放器类型（如酷狗音乐），继续使用原来的方法
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

// 用于控制播放列表播放线程的变量
var (
	playlistPlaybackRunning  bool
	playlistPlaybackMutex    sync.Mutex
	playlistPlaybackStopChan chan struct{}
)

// StartPlaylistPlayback 启动本地播放列表播放线程
// 该线程会检查本地播放列表（默认使用"机器人播放列表"），播放第一首音乐，
// 播放完毕后从列表中移除，然后播放下一首，直到列表为空
func StartPlaylistPlayback(playerType MusicPlayerType) error {
	playlistPlaybackMutex.Lock()
	defer playlistPlaybackMutex.Unlock()

	// 检查是否已经在运行
	if playlistPlaybackRunning {
		return fmt.Errorf("播放列表播放线程已经在运行")
	}

	// 创建停止通道
	playlistPlaybackStopChan = make(chan struct{})
	playlistPlaybackRunning = true

	logx.Infof("启动本地播放列表播放线程，播放器类型: %s", playerType)

	// 启动播放线程
	go func() {
		for {
			select {
			case <-playlistPlaybackStopChan:
				logx.Infof("播放列表播放线程已停止")
				playlistPlaybackMutex.Lock()
				playlistPlaybackRunning = false
				playlistPlaybackMutex.Unlock()
				return
			default:
				// 检查并播放播放列表中的歌曲
				playNextSongFromPlaylist(playerType)

				// 如果播放列表为空，等待一段时间后再次检查
				time.Sleep(5 * time.Second)
			}
		}
	}()

	return nil
}

// StopPlaylistPlayback 停止播放列表播放线程
func StopPlaylistPlayback() {
	playlistPlaybackMutex.Lock()
	defer playlistPlaybackMutex.Unlock()

	if playlistPlaybackRunning && playlistPlaybackStopChan != nil {
		close(playlistPlaybackStopChan)
		logx.Infof("请求停止播放列表播放线程")
	}
}

// IsPlaylistPlaybackRunning 检查播放列表播放线程是否正在运行
func IsPlaylistPlaybackRunning() bool {
	playlistPlaybackMutex.Lock()
	defer playlistPlaybackMutex.Unlock()
	return playlistPlaybackRunning
}

// RemoveSongFromLocalPlaylist 从本地播放列表中移除指定索引的歌曲
func RemoveSongFromLocalPlaylist(playlistID string, index int) error {
	playlist, exists := localPlaylists[playlistID]
	if !exists {
		return fmt.Errorf("本地播放列表不存在")
	}

	if index < 0 || index >= len(playlist.Songs) {
		return fmt.Errorf("歌曲索引超出范围")
	}

	// 移除歌曲
	playlist.Songs = append(playlist.Songs[:index], playlist.Songs[index+1:]...)
	playlist.UpdateTime = time.Now()
	return nil
}

// playNextSongFromPlaylist 播放本地播放列表中的下一首歌曲
func playNextSongFromPlaylist(_ MusicPlayerType) {
	// 查找"机器人播放列表"
	var playlist *LocalPlaylist
	allPlaylists := GetAllLocalPlaylists()

	for _, pl := range allPlaylists {
		if pl.Name == "机器人播放列表" {
			playlist = pl
			break
		}
	}

	// 如果没有找到播放列表或播放列表为空，直接返回
	if playlist == nil || len(playlist.Songs) == 0 {
		return
	}

	// 获取第一首歌曲
	currentSong := playlist.Songs[0]
	songName := fmt.Sprintf("%s-%s", currentSong.Name, currentSong.Singer)
	logx.Infof("开始播放歌曲: %s，预计时长: %d秒", songName, currentSong.Duration)

	// 检查是否收到停止信号
	select {
	case <-playlistPlaybackStopChan:
		logx.Infof("播放线程停止，取消歌曲播放")
		return
	default:
		// 继续播放
	}

	// 使用浏览器播放歌曲
	err := PlaySongByBrowser(currentSong.Name, currentSong.Singer)
	if err != nil {
		logx.Errorf("播放歌曲失败: %v", err)
		return
	}

	// 等待歌曲播放完成
	// 使用歌曲实际时长加5秒的缓冲时间
	playDuration := time.Duration(currentSong.Duration+5) * time.Second
	logx.Infof("等待歌曲播放完成，预计等待时间: %v", playDuration)

	// 分阶段等待，以便能够及时响应停止信号
	totalWait := 0 * time.Second
	checkInterval := 1 * time.Second

	for totalWait < playDuration {
		// 检查是否收到停止信号
		select {
		case <-playlistPlaybackStopChan:
			logx.Infof("播放线程停止，取消歌曲播放等待")
			return
		case <-time.After(checkInterval):
			totalWait += checkInterval
			// 最后一次等待可能需要调整，避免等待过长
			if totalWait+checkInterval > playDuration {
				checkInterval = playDuration - totalWait
			}
		}
	}

	// 再次检查停止信号
	select {
	case <-playlistPlaybackStopChan:
		logx.Infof("播放线程停止，取消歌曲播放等待")
		return
	default:
		// 歌曲播放完成，从播放列表中移除
		logx.Infof("歌曲播放完成，从播放列表中移除: %s", songName)
		RemoveSongFromLocalPlaylist(playlist.ID, 0)
	}
}

// addSongToPlaylist 将指定名称的歌曲添加到播放列表
func addSongToPlaylist(player MusicPlayerType, songName string) error {
	if runtime.GOOS != "windows" {
		logx.Errorf("此功能仅支持 Windows 系统")
		return fmt.Errorf("此功能仅支持 Windows 系统")
	}
	// 检查音乐播放器是否正在运行
	err := checkMusicPlayerRunning(player)
	if err != nil {
		return fmt.Errorf("%s未运行或无法连接: %v", getPlayerName(player), err)
	}

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
		$playerName = "%s"
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
			// 获取窗口位置和大小
			$rect = New-Object RECT
			[WindowHelper]::GetWindowRect($hwnd, [ref]$rect)
			Write-Output "步骤4.1: 窗口位置 - 左: $($rect.Left), 上: $($rect.Top), 右: $($rect.Right), 下: $($rect.Bottom)"
			
			Write-Output "步骤5: 计算搜索框位置..."
			// 计算搜索框绝对位置
			$searchBoxX = $rect.Left + %d
			$searchBoxY = $rect.Top + %d
			Write-Output "步骤5.1: 搜索框位置 - X: $searchBoxX, Y: $searchBoxY"
			
			Write-Output "步骤6: 移动鼠标到搜索框并点击..."
			// 移动鼠标到搜索框位置并点击
			[System.Windows.Forms.Cursor]::Position = New-Object System.Drawing.Point($searchBoxX, $searchBoxY)
			Write-Output "步骤6.1: 鼠标已移动到搜索框位置"
			
			// 正确的鼠标点击方法：使用Windows API
			[User32]::mouse_event(0x0002, 0, 0, 0, 0)  # MOUSEEVENTF_LEFTDOWN
			[User32]::mouse_event(0x0004, 0, 0, 0, 0)  # MOUSEEVENTF_LEFTUP
			Write-Output "步骤6.2: 已执行鼠标左键点击"
			Start-Sleep -Milliseconds 500
			
			Write-Output "步骤7: 发送歌曲名称..."
			// 发送歌曲名称
			// 使用变量方式避免字符被解释为Alt键
			$songName = "%s"
			Write-Output "步骤7.1: 发送的歌曲名称: $songName"
			[System.Windows.Forms.SendKeys]::SendWait($songName)
			Write-Output "步骤7.2: 歌曲名称发送完成"
			Start-Sleep -Milliseconds 500
			
			Write-Output "步骤8: 执行搜索..."
			// 回车执行搜索
			[System.Windows.Forms.SendKeys]::SendWait("{ENTER}")
			Write-Output "步骤8.1: 已发送回车键，等待搜索结果..."
			Start-Sleep -Milliseconds 1500
			
			Write-Output "步骤9: 选择并添加歌曲到播放列表..."
			// 选择第一个结果并添加到播放列表（Shift+Enter）
			[System.Windows.Forms.SendKeys]::SendWait("{DOWN}")
			Write-Output "步骤9.1: 已选择第一个结果"
			Start-Sleep -Milliseconds 200
			[System.Windows.Forms.SendKeys]::SendWait("+{ENTER}")
			Write-Output "步骤9.2: 已添加到播放列表"
		} else {
			Write-Output "错误步骤: 窗口查找失败，hwnd为0"
			throw "错误：未找到" + $playerName + "窗口。请确保" + $playerName + "已启动且窗口标题包含" + $playerName + "。" 
		}
	`, getPlayerName(player), windowCheck, 0, 0, escapedSongName)

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
			# 使用变量方式避免字符被解释为Alt键
			$songName = "%songName"
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
	`, windowCheck, 0, 0, escapedSongName)

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
		$windowNamesToTry = @("`+strings.Join(windowNames, `", "`)+`")
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
