package danmu

import (
	"strings"
	"github.com/xbclub/BilibiliDanmuRobot-Core/entity"
	"github.com/xbclub/BilibiliDanmuRobot-Core/logic"
	"github.com/xbclub/BilibiliDanmuRobot-Core/svc"
	"github.com/xbclub/BilibiliDanmuRobot-Core/utiles"
	"github.com/zeromicro/go-zero/core/logx"
)

// ProcessMusicRequest 处理点歌请求
func ProcessMusicRequest(msg string, svcCtx *svc.ServiceContext, reply ...*entity.DanmuMsgTextReplyInfo) {
	// 检查是否为点歌指令
	if strings.HasPrefix(msg, "点歌 ") {
		songName := strings.TrimSpace(strings.TrimPrefix(msg, "点歌 "))
		if songName != "" {
			// 增加输入长度验证，防止过长输入
			if len(songName) > 100 {
				logx.Errorf("点歌请求被拒绝，歌曲名过长: %s", songName)
				logic.PushToBulletSender("点歌失败：歌曲名过长！", reply...)
				return
			}
			
			// 调用QQ音乐添加歌曲到播放列表
			err := utiles.AddSongToPlaylist(utiles.QQMusic, songName)
			if err != nil {
				// 优化日志输出，提供更清晰的错误信息
				logx.Errorf("点歌失败: %v，歌曲名: %s", err, songName)
				logic.PushToBulletSender("点歌失败！", reply...)
			} else {
				logx.Infof("已添加歌曲到播放列表: %s", songName)
				logic.PushToBulletSender("已添加歌曲到播放列表:"+songName, reply...)
			}
		}
	}
}