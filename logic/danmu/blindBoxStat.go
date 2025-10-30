package danmu

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"

	_ "github.com/glebarez/go-sqlite"
	"github.com/golang-module/carbon/v2"
	"github.com/xbclub/BilibiliDanmuRobot-Core/entity"
	"github.com/xbclub/BilibiliDanmuRobot-Core/logic"
	"github.com/xbclub/BilibiliDanmuRobot-Core/model"
	"github.com/xbclub/BilibiliDanmuRobot-Core/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

var errInfo string = "盲盒统计服务异常"

func SaveBlindBoxStat(g *entity.SendGiftText, svcCtx *svc.ServiceContext) {
	logx.Info(g.Data.BlindGift.OriginalGiftName)
	if g.Data.BlindGift.OriginalGiftName == "" {
		return
	}
	now := carbon.Now(carbon.Local)
	err := svcCtx.BlindBoxStatModel.Insert(context.Background(), nil, &model.BlindBoxStatBase{
		Uid:               int64(g.Data.UID),
		BlindBoxName:      g.Data.BlindGift.OriginalGiftName,
		Price:             int32(g.Data.Price),
		OriginalGiftPrice: int32(g.Data.BlindGift.OriginalGiftPrice),
		Cnt:               int32(g.Data.Num),
		Year:              int16(now.Year()),
		Month:             int16(now.Month()),
		Day:               int16(now.Day()),
	})
	if err != nil {
		logx.Alert("保存盲盒数据出错!!! " + err.Error())
	} else {
		logx.Info("盲盒数据保存成功!!! ")
	}
}

func DoBlindBoxStat(msg, uid, username string, svcCtx *svc.ServiceContext, reply ...*entity.DanmuMsgTextReplyInfo) {
	if !svcCtx.Config.BlindBoxStat {
		return
	}

	// 判断模式（今日 / 指定月份）
	mode := ""
	month := 0
	re := regexp.MustCompile(`^(?:今日盲盒|本日盲盒|今日盲盒盈亏)$`)
	if re.MatchString(msg) {
		mode = "today"
	} else {
		re := regexp.MustCompile(`(?P<month>^[0-9]+)月盲盒$`)
		match := re.FindStringSubmatch(msg)
		if len(match) != 2 {
			return
		}
		mode = "month"
		var err error
		month, err = strconv.Atoi(match[1])
		if err != nil || month < 1 || month > 12 {
			logic.PushToBulletSender(fmt.Sprintf("月份「%s」不正确!", match[1]), reply...)
			return
		}
	}

	// 解析用户 ID
	id, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		logx.Error(err)
		logic.PushToBulletSender(errInfo, reply...)
		return
	}

	// 获取当前时间
	now := carbon.Now(carbon.Local)
	var ret *model.Result

	// 查询数据
	if mode == "today" {
		if svcCtx.UserID == id {
			// 主播查询今日数据
			ret, err = svcCtx.BlindBoxStatModel.GetTotal(context.Background(), int16(now.Year()), int16(now.Month()), int16(now.Day()))
		} else {
			// 普通用户查询今日数据
			ret, err = svcCtx.BlindBoxStatModel.GetTotalOnePersion(context.Background(), id, int16(now.Year()), int16(now.Month()), int16(now.Day()))
		}
		if err != nil {
			logx.Alert("盲盒统计出错了! " + err.Error())
			logic.PushToBulletSender(errInfo, reply...)
			return
		}

		r := float64(ret.R) / 1000.0
		switch {
		case ret.R > 0:
			logic.PushToBulletSender(fmt.Sprintf("今天开%d个盲盒, 赚了%.2f元", ret.C, r), reply...)
		case ret.R == 0:
			logic.PushToBulletSender(fmt.Sprintf("今天共开%d个盲盒, 没亏没赚!", ret.C), reply...)
		default:
			logic.PushToBulletSender(fmt.Sprintf("今天共开%d个盲盒, 亏了%.2f元", ret.C, math.Abs(r)), reply...)
		}
		return
	}

	// month 模式
	{
		// 使用传入的 month（已验证）
		if svcCtx.UserID == id {
			// 主播查询当月数据
			ret, err = svcCtx.BlindBoxStatModel.GetTotal(context.Background(), int16(now.Year()), int16(month), 0)
		} else {
			// 普通用户查询当月数据
			ret, err = svcCtx.BlindBoxStatModel.GetTotalOnePersion(context.Background(), id, int16(now.Year()), int16(month), 0)
		}
		if err != nil {
			logx.Alert("盲盒统计出错了! " + err.Error())
			logic.PushToBulletSender(errInfo, reply...)
			return
		}

		r := float64(ret.R) / 1000.0
		monthLabel := fmt.Sprintf("%d", month)
		switch {
		case ret.R > 0:
			logic.PushToBulletSender(fmt.Sprintf("%s月共开%d个, 赚了%.2f元", monthLabel, ret.C, r), reply...)
		case ret.R == 0:
			logic.PushToBulletSender(fmt.Sprintf("%s月共开%d个, 没亏没赚!", monthLabel, ret.C), reply...)
		default:
			logic.PushToBulletSender(fmt.Sprintf("%s月共开%d个, 亏了%.2f元", monthLabel, ret.C, math.Abs(r)), reply...)
		}
	}
}
