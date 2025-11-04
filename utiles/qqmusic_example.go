package utiles

import (
	"fmt"
	"log"
)

// QQMusicExample 展示如何使用QQ音乐API和本地播放列表功能
func QQMusicExample() {
	// 创建QQ音乐API客户端
	api := NewQQMusicAPI()

	// 示例1: 搜索歌曲并添加到本地播放列表
	fmt.Println("示例1: 搜索歌曲并添加到本地播放列表")
	
	// 搜索歌曲
	songs, err := api.SearchSong("周杰伦", 1, 5)
	if err != nil {
		log.Printf("搜索歌曲失败: %v\n", err)
		return
	}
	
	fmt.Printf("搜索到 %d 首歌曲:\n", len(songs))
	for i, song := range songs {
		fmt.Printf("%d. %s - %s [专辑: %s, 时长: %d秒]\n", 
			i+1, song.Name, song.Singer, song.Album, song.Duration)
	}
	
	// 创建本地播放列表
	if len(songs) > 0 {
		playlist := CreateLocalPlaylist("我的音乐收藏", "我喜欢的歌曲")
		fmt.Printf("创建了本地播放列表: %s (ID: %s)\n", playlist.Name, playlist.ID)
		
		// 添加歌曲到本地播放列表
		for i := 0; i < len(songs) && i < 3; i++ { // 添加前3首歌
			err := AddSongToLocalPlaylist(playlist.ID, songs[i])
			if err != nil {
				log.Printf("添加歌曲失败: %v\n", err)
			} else {
				fmt.Printf("已添加: %s - %s\n", songs[i].Name, songs[i].Singer)
			}
		}
		
		// 获取并显示播放列表信息
		updatedPlaylist, _ := GetLocalPlaylist(playlist.ID)
		fmt.Printf("播放列表 '%s' 现在有 %d 首歌曲\n\n", updatedPlaylist.Name, len(updatedPlaylist.Songs))
	}

	// 示例2: 模拟在线播放列表操作（需要实际登录认证才能使用）
	fmt.Println("示例2: 模拟在线播放列表操作")
	
	// 模拟创建在线播放列表（实际使用需要用户登录）
	playlistID, err := api.CreatePlaylist("在线音乐收藏", "在线保存的歌曲")
	if err != nil {
		log.Printf("创建在线播放列表失败: %v\n", err)
	} else {
		fmt.Printf("模拟创建了在线播放列表，ID: %s (注意: 实际使用需要用户登录认证)\n", playlistID)
		
		// 准备歌曲ID列表
		songMids := make([]string, 0)
		for i := 0; i < len(songs) && i < 2; i++ {
			songMids = append(songMids, songs[i].Mid)
		}
		
		// 模拟添加歌曲到在线播放列表
		err = api.AddSongToPlaylist(playlistID, songMids)
		if err != nil {
			log.Printf("添加歌曲到在线播放列表失败: %v\n", err)
		} else {
			fmt.Printf("模拟添加 %d 首歌曲到在线播放列表 (注意: 实际使用需要用户登录认证)\n", len(songMids))
		}
	}

	// 示例3: 管理本地播放列表
	fmt.Println("\n示例3: 管理本地播放列表")
	
	// 获取所有本地播放列表
	allPlaylists := GetAllLocalPlaylists()
	fmt.Printf("共有 %d 个本地播放列表\n", len(allPlaylists))
	
	for _, pl := range allPlaylists {
		fmt.Printf("- %s (创建于: %s, 更新于: %s, %d首歌曲)\n", 
			pl.Name, pl.CreateTime.Format("2006-01-02 15:04:05"), 
			pl.UpdateTime.Format("2006-01-02 15:04:05"), len(pl.Songs))
	}

	// 示例4: 尝试添加重复歌曲
	if len(songs) > 0 && len(allPlaylists) > 0 {
		fmt.Println("\n示例4: 尝试添加重复歌曲")
		
		// 尝试添加已存在的歌曲
		err := AddSongToLocalPlaylist(allPlaylists[0].ID, songs[0])
		if err != nil {
			fmt.Printf("预期的错误: %v\n", err)
		}
	}
}