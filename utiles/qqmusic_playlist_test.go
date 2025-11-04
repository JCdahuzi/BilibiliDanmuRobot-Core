package utiles

import (
	"testing"
)

// TestLocalPlaylist 测试本地播放列表功能
func TestLocalPlaylist(t *testing.T) {
	// 创建测试歌曲
	testSong1 := SongSearchResult{
		ID:       "test_id_1",
		Mid:      "test_mid_1",
		Name:     "测试歌曲1",
		Singer:   "测试歌手1",
		Album:    "测试专辑1",
		Duration: 180,
	}

	testSong2 := SongSearchResult{
		ID:       "test_id_2",
		Mid:      "test_mid_2",
		Name:     "测试歌曲2",
		Singer:   "测试歌手2",
		Album:    "测试专辑2",
		Duration: 200,
	}

	// 测试1: 创建本地播放列表
	playlist := CreateLocalPlaylist("测试播放列表", "用于测试的播放列表")
	if playlist.ID == "" || playlist.Name != "测试播放列表" {
		t.Errorf("创建播放列表失败，期望名称为'测试播放列表'，实际为'%s'", playlist.Name)
	}

	// 测试2: 添加歌曲到播放列表
	err := AddSongToLocalPlaylist(playlist.ID, testSong1)
	if err != nil {
		t.Errorf("添加歌曲失败: %v", err)
	}

	// 验证歌曲已添加
	updatedPlaylist, _ := GetLocalPlaylist(playlist.ID)
	if len(updatedPlaylist.Songs) != 1 || updatedPlaylist.Songs[0].Mid != testSong1.Mid {
		t.Errorf("歌曲添加验证失败，期望1首歌曲，实际%v首", len(updatedPlaylist.Songs))
	}

	// 测试3: 尝试添加重复歌曲
	err = AddSongToLocalPlaylist(playlist.ID, testSong1)
	if err == nil || err.Error() != "歌曲已存在于播放列表中" {
		t.Errorf("添加重复歌曲验证失败，期望错误'歌曲已存在于播放列表中'，实际为'%v'", err)
	}

	// 测试4: 添加第二首歌曲
	err = AddSongToLocalPlaylist(playlist.ID, testSong2)
	if err != nil {
		t.Errorf("添加第二首歌曲失败: %v", err)
	}

	// 验证第二首歌曲已添加
	updatedPlaylist, _ = GetLocalPlaylist(playlist.ID)
	if len(updatedPlaylist.Songs) != 2 {
		t.Errorf("第二首歌曲添加验证失败，期望2首歌曲，实际%v首", len(updatedPlaylist.Songs))
	}

	// 测试5: 获取所有本地播放列表
	allPlaylists := GetAllLocalPlaylists()
	if len(allPlaylists) == 0 {
		t.Errorf("获取所有播放列表失败，期望至少1个播放列表")
	}

	// 测试6: 获取不存在的播放列表
	_, err = GetLocalPlaylist("non_existent_playlist")
	if err == nil || err.Error() != "本地播放列表不存在" {
		t.Errorf("获取不存在播放列表验证失败，期望错误'本地播放列表不存在'，实际为'%v'", err)
	}

	// 测试7: 测试QQ音乐API的模拟方法（不会实际执行，因为没有用户认证）
	api := NewQQMusicAPI()
	playlistID, err := api.CreatePlaylist("在线测试播放列表", "在线测试")
	if err != nil {
		t.Errorf("模拟创建在线播放列表失败: %v", err)
	}

	err = api.AddSongToPlaylist(playlistID, []string{testSong1.Mid, testSong2.Mid})
	if err != nil {
		t.Errorf("模拟添加歌曲到在线播放列表失败: %v", err)
	}
}