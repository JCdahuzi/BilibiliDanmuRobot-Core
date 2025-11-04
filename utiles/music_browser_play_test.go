package utiles

import (
	"testing"
	"time"
)

// TestPlaySongByBrowser 测试通过浏览器播放歌曲的功能
// 注意：此测试会实际打开浏览器播放音乐，建议在测试环境中谨慎运行
func TestPlaySongByBrowser(t *testing.T) {
	// 使用知名歌曲进行测试，如《晴天》
	songName := "晴天"
	
	// 执行浏览器播放功能
	err := PlaySongByBrowser(songName, "")
	if err != nil {
		t.Errorf("播放歌曲失败: %v", err)
		return
	}
	
	// 如果没有错误，让测试等待一小段时间以便观察浏览器是否打开
	// 实际使用时不需要这个等待
	t.Log("测试等待5秒，观察浏览器是否成功打开并播放音乐...")
	time.Sleep(5 * time.Second)
	
	t.Log("浏览器播放测试完成")
}

// ExamplePlaySongByBrowser 示例：如何使用浏览器播放音乐
func ExamplePlaySongByBrowser() {
	// 播放指定的歌曲
	err := PlaySongByBrowser("小幸运", "")
	if err != nil {
		// 处理错误
		return
	}
	
	// 函数会自动打开默认浏览器并开始播放
	// ...
}