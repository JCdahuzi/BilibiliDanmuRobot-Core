package utiles

import (
	"testing"
)

// TestAddSongToPlaylist 测试添加歌曲到播放列表功能
// 注意：此测试在没有实际QQ音乐客户端运行时可能会失败，但应该能够正常搜索歌曲并添加到本地播放列表
func TestAddSongToPlaylist(t *testing.T) {
	// 测试参数
	songName := "周杰伦 晴天"
	playerType := QQMusic
	
	// 执行测试
	err := AddSongToPlaylist(playerType, songName)
	if err != nil {
		// 检查错误是否是因为QQ音乐客户端未运行
		// 如果是本地播放列表添加成功但客户端添加失败，这是可以接受的
		if err.Error() == "QQ音乐未运行或无法连接: QQ音乐未运行" {
			t.Log("警告：QQ音乐客户端未运行，但测试应继续验证本地播放列表功能")
		} else {
			t.Errorf("添加歌曲到播放列表失败: %v", err)
		}
	}
	
	// 验证本地播放列表是否成功创建并添加了歌曲
	allPlaylists := GetAllLocalPlaylists()
	foundPlaylist := false
	for _, playlist := range allPlaylists {
		if playlist.Name == "机器人播放列表" {
			foundPlaylist = true
			t.Logf("找到机器人播放列表，包含 %d 首歌曲", len(playlist.Songs))
			if len(playlist.Songs) > 0 {
				t.Logf("播放列表中的第一首歌曲: %s - %s", playlist.Songs[0].Name, playlist.Songs[0].Singer)
			} else {
				t.Errorf("机器人播放列表创建成功，但未添加任何歌曲")
			}
			break
		}
	}
	
	if !foundPlaylist {
		t.Errorf("未找到机器人播放列表")
	}
	
	// 测试搜索功能
	t.Log("测试QQ音乐API搜索功能")
	api := NewQQMusicAPI()
	songs, err := api.SearchSong("周杰伦", 1, 3)
	if err != nil {
		t.Errorf("搜索歌曲失败: %v", err)
	} else if len(songs) == 0 {
		t.Errorf("未找到搜索结果")
	} else {
		t.Logf("搜索成功，找到 %d 首歌曲", len(songs))
		for i, song := range songs {
			t.Logf("  %d. %s - %s (专辑: %s)", i+1, song.Name, song.Singer, song.Album)
		}
	}
}