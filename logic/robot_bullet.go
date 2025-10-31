package logic

import (
	"context"
	"github.com/xbclub/BilibiliDanmuRobot-Core/entity"
	"github.com/xbclub/BilibiliDanmuRobot-Core/http"
	"github.com/xbclub/BilibiliDanmuRobot-Core/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"strings"
)

var robot *BulletRobot

type BulletRobot struct {
	bulletRobotChan chan entity.Bullet
}

func PushToBulletRobot(content string, reply ...*entity.DanmuMsgTextReplyInfo) {
	logx.Infof("PushToBulletRobot成功：%s", content)
	buttle := entity.Bullet{
		Msg:   content,
		Reply: reply,
	}
	robot.bulletRobotChan <- buttle
}

func StartBulletRobot(ctx context.Context, svcCtx *svc.ServiceContext) {
	robot = &BulletRobot{
		bulletRobotChan: make(chan entity.Bullet, 1000),
	}

	var content entity.Bullet

	for {
		select {
		case <-ctx.Done():
			goto END
		case content = <-robot.bulletRobotChan:
			handleRobotBullet(content, svcCtx)
		}
	}
END:
}

func handleRobotBullet(content entity.Bullet, svcCtx *svc.ServiceContext) {
	var err error
	var reply string
	if svcCtx.Config.RobotMode == "ChatGPT" {
		if reply, err = http.RequestChatgptRobot(content.Msg, svcCtx); err != nil {
			logx.Errorf("请求ChatGPT机器人失败：%v", err)
			PushToBulletSender("不好意思，机器人坏掉了...", content.Reply...)
			return
		}
	} else if svcCtx.Config.RobotMode == "Qingyunke" {
		if reply, err = http.RequestQingyunkeRobot(content.Msg); err != nil {
			logx.Errorf("请求Qingyunke机器人失败：%v", err)
			PushToBulletSender("不好意思，机器人坏掉了...", content.Reply...)
			return
		}
		bulltes := splitRobotReply(reply, svcCtx)
		for _, v := range bulltes {
			PushToBulletSender(v, content.Reply...)
		}
		return
	} else if svcCtx.Config.RobotMode == "DeepSeek" {
		if reply, err = http.RequestDeepSeekRobot(content.Msg, svcCtx); err != nil {
			logx.Errorf("请求DeepSeek机器人失败：%v", err)
			PushToBulletSender("不好意思，机器人坏掉了...", content.Reply...)
			return
		}
		// 处理异常信息
		reply = handleRobotReply(reply, svcCtx)
	} else {
		logx.Errorf("未知的机器人模式：%s", svcCtx.Config.RobotMode)
		PushToBulletSender("不好意思，机器人坏掉了...", content.Reply...)
		return
	}
	PushToBulletSender(reply, content.Reply...)
	logx.Infof("机器人回复：%s", reply)

}

func handleRobotReply(content string, svcCtx *svc.ServiceContext) string {
	// 处理异常信息
	return content
}

// 将机器人回复语句中的 {br} 进行分割
// b站弹幕一次只能发20个字符，需要切分
func splitRobotReply(content string, svcCtx *svc.ServiceContext) []string {
	//var res []string
	reply := strings.Split(content, "{br}")

	//for _, r := range reply {
	//	// 长度大于20再分割
	//	zh := []rune(r)
	//	if len(zh) > 20 {
	//		i := 0
	//		for i < len(zh) {
	//			if i+20 > len(zh) {
	//				res = append(res, string(zh[i:]))
	//			} else {
	//				res = append(res, string(zh[i:i+20]))
	//			}
	//			i += 20
	//		}
	//	} else {
	//		res = append(res, string(zh))
	//	}
	//}
	return reply
}
