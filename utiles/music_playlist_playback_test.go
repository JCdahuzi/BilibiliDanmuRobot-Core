package utiles

import (
	"fmt"
	"testing"
	"time"
)

// TestPlaylistPlayback 测试播放列表自动播放功能
// 注意：此测试仅验证线程启动和控制功能，不会实际播放音乐
func TestPlaylistPlayback(t *testing.T) {
	// 首先确保停止可能正在运行的播放线程
	StopPlaylistPlayback()
	time.Sleep(1 * time.Second)
	
	// 创建测试播放列表并添加歌曲
	playlist := CreateLocalPlaylist("机器人播放列表", "测试用播放列表")
	
	// 添加测试歌曲
	testSongs := []SongSearchResult{
		{ID: "test1", Mid: "test1", Name: "测试歌曲1", Singer: "测试歌手", Album: "测试专辑", Duration: 10}, // 10秒的短歌曲
		{ID: "test2", Mid: "test2", Name: "测试歌曲2", Singer: "测试歌手", Album: "测试专辑", Duration: 10}, // 10秒的短歌曲
	}
	
	for _, song := range testSongs {
		err := AddSongToLocalPlaylist(playlist.ID, song)
		if err != nil {
			t.Errorf("添加测试歌曲失败: %v", err)
		}
	}
	
	// 验证播放列表包含2首歌曲
	if len(playlist.Songs) != 2 {
		t.Errorf("播放列表应该包含2首歌曲，实际有%d首", len(playlist.Songs))
	}
	
	t.Log("开始测试播放线程控制功能...")
	
	// 测试启动播放线程
	err := StartPlaylistPlayback(QQMusic)
	if err != nil {
		t.Errorf("启动播放线程失败: %v", err)
	}
	
	// 验证播放线程正在运行
	if !IsPlaylistPlaybackRunning() {
		t.Errorf("播放线程应该在运行，但实际没有运行")
	}
	
	t.Log("播放线程已启动")
	
	// 测试尝试重复启动播放线程
	err = StartPlaylistPlayback(QQMusic)
	if err == nil || err.Error() != "播放列表播放线程已经在运行" {
		t.Errorf("重复启动播放线程应该返回错误，但返回: %v", err)
	}
	
	// 等待一段时间（不要等到歌曲播放完成，因为这只是功能测试）
	t.Log("等待3秒...")
	time.Sleep(3 * time.Second)
	
	// 测试停止播放线程
	StopPlaylistPlayback()
	t.Log("请求停止播放线程")
	
	// 等待线程停止
	time.Sleep(1 * time.Second)
	
	// 验证播放线程已停止
	if IsPlaylistPlaybackRunning() {
		t.Errorf("播放线程应该已停止，但实际仍在运行")
	}
	
	t.Log("播放线程已停止")
	
	// 清理测试数据 - 移除所有歌曲
	for len(playlist.Songs) > 0 {
		RemoveSongFromLocalPlaylist(playlist.ID, 0)
	}
	
	t.Log("测试完成，播放列表已清空")
	fmt.Println("播放列表自动播放功能测试完成")
}

// 以下是演示如何使用播放列表自动播放功能的示例代码
func Example() {
	// 1. 启动播放线程
	if err := StartPlaylistPlayback(QQMusic); err != nil {
		fmt.Printf("启动播放线程失败: %v\n", err)
		return
	}
	
	fmt.Println("播放列表自动播放已启动")
	
	// 2. 添加歌曲到播放列表
	AddSongToPlaylist(QQMusic, "周杰伦 晴天")
	AddSongToPlaylist(QQMusic, "陈奕迅 十年")
	
	// 3. 程序继续执行其他任务...
	
	// 4. 当需要停止播放时
	// StopPlaylistPlayback()
}