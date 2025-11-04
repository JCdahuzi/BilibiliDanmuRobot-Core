package utiles

import (
	"fmt"
	"testing"
)

// TestQQMusicAPISearch 测试搜索歌曲功能
func TestQQMusicAPISearch(t *testing.T) {
	// 创建API客户端
	api := NewQQMusicAPI()
	
	// 设置重试配置（可选）
	api.SetRetryConfig(3, 1000)
	
	// 搜索歌曲
	songs, err := api.SearchSong("周杰伦", 1, 10)
	if err != nil {
		t.Errorf("搜索歌曲失败: %v", err)
		return
	}
	
	// 打印搜索结果
	fmt.Println("搜索结果数量:", len(songs))
	for i, song := range songs {
		fmt.Printf("[%d] %s - %s (专辑: %s, 时长: %ds)\n", 
			i+1, song.Name, song.Singer, song.Album, song.Duration)
	}
	
	// 如果有搜索结果，测试获取歌曲详情
	if len(songs) > 0 {
		t.Run("GetSongDetail", func(t *testing.T) {
			detail, err := api.GetSongDetail(songs[0].Mid)
			if err != nil {
				t.Errorf("获取歌曲详情失败: %v", err)
				return
			}
			fmt.Printf("\n歌曲详情: %s - %s\n", detail.Name, detail.Singer)
		})
	}
}

// TestQQMusicAPIRecommend 测试获取推荐歌单功能
func TestQQMusicAPIRecommend(t *testing.T) {
	// 创建API客户端
	api := NewQQMusicAPI()
	
	// 获取推荐歌单
	playlists, err := api.GetRecommendPlaylists(1, 5)
	if err != nil {
		t.Errorf("获取推荐歌单失败: %v", err)
		return
	}
	
	// 打印推荐歌单
	fmt.Println("\n推荐歌单:")
	for i, pl := range playlists {
		fmt.Printf("[%d] %s (作者: %s, 歌曲数: %d)\n", 
			i+1, pl.Name, pl.Author, pl.Count)
	}
}

// TestQQMusicAPIHotSongs 测试获取热门歌曲功能
func TestQQMusicAPIHotSongs(t *testing.T) {
	// 创建API客户端
	api := NewQQMusicAPI()
	
	// 获取热门歌曲
	hotSongs, err := api.GetHotSongs(10)
	if err != nil {
		t.Errorf("获取热门歌曲失败: %v", err)
		return
	}
	
	// 打印热门歌曲
	fmt.Println("\n热门歌曲:")
	for i, song := range hotSongs {
		fmt.Printf("[%d] %s - %s\n", i+1, song.Name, song.Singer)
	}
}