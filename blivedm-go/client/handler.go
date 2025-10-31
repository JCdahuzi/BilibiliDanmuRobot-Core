package client

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/xbclub/BilibiliDanmuRobot-Core/blivedm-go/message"
	"github.com/xbclub/BilibiliDanmuRobot-Core/blivedm-go/packet"
	"github.com/xbclub/BilibiliDanmuRobot-Core/blivedm-go/utils"
	log "github.com/zeromicro/go-zero/core/logx"
)

var (
	// 已知的命令类型列表
	knownCMD = []string{
		"INTERACT_WORD",                      // 用户进入房间
		"HOT_RANK_SETTLEMENT",                // 热门排名结算
		"DANMU_GIFT_LOTTERY_START",           // 弹幕礼物抽奖开始
		"WELCOME_GUARD",                      // 欢迎舰长
		"PK_PROCESS",                         // PK过程
		"PK_BATTLE_PRO_TYPE",                 // PK战斗专业类型
		"MATCH_TEAM_GIFT_RANK",               // 匹配团队礼物排名
		"PK_BATTLE_CRIT",                     // PK战斗暴击
		"LUCK_GIFT_AWARD_USER",               // 幸运礼物获奖用户
		"SCORE_CARD",                         // 积分卡
		"ONLINE_RANK_V2",                     // 在线排名V2
		"PK_BATTLE_SPECIAL_GIFT",             // PK战斗特殊礼物
		"SEND_TOP",                           // 发送顶部
		"SUPER_CHAT_MESSAGE_JPN",             // 日语超级聊天消息
		"ANIMATION",                          // 动画
		"GUARD_LOTTERY_START",                // 舰长抽奖开始
		"WEEK_STAR_CLOCK",                    // 周星时钟
		"WELCOME",                            // 欢迎
		"WIN_ACTIVITY",                       // 获奖活动
		"ROOM_KICKOUT",                       // 房间踢出
		"CHANGE_ROOM_INFO",                   // 更改房间信息
		"ROOM_SKIN_MSG",                      // 房间皮肤消息
		"ROOM_BLOCK_MSG",                     // 房间屏蔽消息
		"SUPER_CHAT_ENTRANCE",                // 超级聊天入口
		"PK_BATTLE_RANK_CHANGE",              // PK战斗排名变化
		"ROOM_LOCK",                          // 房间锁定
		"TV_END",                             // 电视结束
		"PK_PRE",                             // PK预热
		"ROOM_SILENT_OFF",                    // 房间静音关闭
		"SEND_GIFT",                          // 发送礼物
		"DANMU_MSG",                          // 弹幕消息
		"ANCHOR_LOT_START",                   // 主播抽奖开始
		"ROOM_BOX_USER",                      // 房间宝箱用户
		"ONLINE_RANK_TOP3",                   // 在线排名前3
		"WIDGET_BANNER",                      // 小部件横幅
		"PK_BATTLE_START",                    // PK战斗开始
		"ACTIVITY_MATCH_GIFT",                // 活动匹配礼物
		"PK_AGAIN",                           // 再次PK
		"PK_MATCH",                           // PK匹配
		"RAFFLE_START",                       // 抽奖开始
		"LIVE",                               // 开播
		"WISH_BOTTLE",                        // 许愿瓶
		"GUARD_ACHIEVEMENT_ROOM",             // 房间舰长成就
		"ONLINE_RANK_COUNT",                  // 在线排名计数
		"COMMON_NOTICE_DANMAKU",              // 普通告弹幕
		"LOL_ACTIVITY",                       // LOL活动
		"HOT_RANK_CHANGED",                   // 热门排名变化
		"ROOM_BLOCK_INTO",                    // 房间屏蔽进入
		"ROOM_LIMIT",                         // 房间限制
		"PANEL",                              // 面板
		"RAFFLE_END",                         // 抽奖结束
		"ENTRY_EFFECT",                       // 进入效果
		"STOP_LIVE_ROOM_LIST",                // 停止直播房间列表
		"TV_START",                           // 电视开始
		"WATCH_LPL_EXPIRED",                  // 观看LPL过期
		"PK_BATTLE_PRE",                      // PK战斗预热
		"USER_TOAST_MSG",                     // 用户Toast消息
		"BOX_ACTIVITY_START",                 // 宝箱活动开始
		"PK_MIC_END",                         // PK麦克风结束
		"LIVE_INTERACTIVE_GAME",              // 直播互动游戏
		"ROOM_BANNER",                        // 房间横幅
		"PK_BATTLE_GIFT",                     // PK战斗礼物
		"MESSAGEBOX_USER_GAIN_MEDAL",         // 消息盒用户获得勋章
		"LITTLE_TIPS",                        // 小贴士
		"HOUR_RANK_AWARDS",                   // 小时排名奖励
		"NOTICE_MSG",                         // 通知消息
		"ROOM_REAL_TIME_MESSAGE_UPDATE",      // 房间实时消息更新
		"ANCHOR_LOT_END",                     // 主播抽奖结束
		"PREPARING",                          // 准备中（下播）
		"GUARD_BUY",                          // 购买舰长
		"ROOM_CHANGE",                        // 房间变更
		"room_admin_entrance",                // 房间管理员入口
		"CHASE_FRAME_SWITCH",                 // 追逐帧切换
		"DANMU_GIFT_LOTTERY_AWARD",           // 弹幕礼物抽奖获奖
		"PK_BATTLE_VOTES_ADD",                // PK战斗投票增加
		"PK_BATTLE_END",                      // PK战斗结束
		"CUT_OFF",                            // 切断
		"PK_BATTLE_PROCESS",                  // PK战斗过程
		"PK_BATTLE_SETTLE_USER",              // PK战斗结算用户
		"ANCHOR_LOT_AWARD",                   // 主播抽奖获奖
		"WIN_ACTIVITY_USER",                  // 获奖活动用户
		"VOICE_JOIN_STATUS",                  // 语音加入状态
		"DANMU_GIFT_LOTTERY_END",             // 弹幕礼物抽奖结束
		"ROOM_RANK",                          // 房间排名
		"SUPER_CHAT_MESSAGE",                 // 超级聊天消息
		"ACTIVITY_BANNER_UPDATE_V2",          // 活动横幅更新V2
		"SPECIAL_GIFT",                       // 特殊礼物
		"ROOM_SILENT_ON",                     // 房间静音开启
		"WARNING",                            // 警告
		"ROOM_ADMINS",                        // 房间管理员
		"COMBO_SEND",                         // 连击发送
		"HOT_RANK_SETTLEMENT_V2",             // 热门排名结算V2
		"ANCHOR_LOT_CHECKSTATUS",             // 主播抽奖检查状态
		"HOT_RANK_CHANGED_V2",                // 热门排名变化V2
		"SUPER_CHAT_MESSAGE_DELETE",          // 超级聊天消息删除
		"PK_END",                             // PK结束
		"PK_SETTLE",                          // PK结算
		"ROOM_REFRESH",                       // 房间刷新
		"PK_START",                           // PK开始
		"COMBO_END",                          // 连击结束
		"PK_LOTTERY_START",                   // PK抽奖开始
		"GUARD_WINDOWS_OPEN",                 // 舰长窗口打开
		"REENTER_LIVE_ROOM",                  // 重新进入直播间
		"MESSAGEBOX_USER_MEDAL_CHANGE",       // 消息盒用户勋章变更
		"MESSAGEBOX_USER_MEDAL_COMPENSATION", // 消息盒用户勋章补偿
		"LITTLE_MESSAGE_BOX",                 // 小消息盒
		"PK_BATTLE_PRE_NEW",                  // PK战斗预热新
		"PK_BATTLE_START_NEW",                // PK战斗开始新
		"PK_BATTLE_PROCESS_NEW",              // PK战斗过程新
		"PK_BATTLE_FINAL_PROCESS",            // PK战斗最终过程
		"PK_BATTLE_SETTLE_V2",                // PK战斗结算V2
		"PK_BATTLE_SETTLE_NEW",               // PK战斗结算新
		"PK_BATTLE_PUNISH_END",               // PK战斗惩罚结束
		"PK_BATTLE_VIDEO_PUNISH_BEGIN",       // PK战斗视频惩罚开始
		"PK_BATTLE_VIDEO_PUNISH_END",         // PK战斗视频惩罚结束
		"ENTRY_EFFECT_MUST_RECEIVE",          // 必须接收的进入效果
		"SUPER_CHAT_AUDIT",                   // 超级聊天审核
		"VIDEO_CONNECTION_JOIN_START",        // 视频连接加入开始
		"VIDEO_CONNECTION_JOIN_END",          // 视频连接加入结束
		"VIDEO_CONNECTION_MSG",               // 视频连接消息
		"VTR_GIFT_LOTTERY",                   // VTR礼物抽奖
		"RED_POCKET_START",                   // 红包开始
		"FULL_SCREEN_SPECIAL_EFFECT",         // 全屏特殊效果
		"POPULARITY_RED_POCKET_START",        // 人气红包开始
		"POPULARITY_RED_POCKET_WINNER_LIST",  // 人气红包获奖者列表
		"USER_PANEL_RED_ALARM",               // 用户面板红色警报
		"SHOPPING_CART_SHOW",                 // 购物车显示
		"THERMAL_STORM_DANMU_BEGIN",          // 热力风暴弹幕开始
		"THERMAL_STORM_DANMU_UPDATE",         // 热力风暴弹幕更新
		"THERMAL_STORM_DANMU_CANCEL",         // 热力风暴弹幕取消
		"THERMAL_STORM_DANMU_OVER",           // 热力风暴弹幕结束
		"MILESTONE_UPDATE_EVENT",             // 里程碑更新事件
		"WEB_REPORT_CONTROL",                 // 网络报告控制
		"DANMU_TAG_CHANGE",                   // 弹幕标签变更
		"RANK_REM",                           // 排名REM
		"LIVE_PLAYER_LOG_RECYCLE",            // 直播播放器日志回收
		"LIVE_INTERNAL_ROOM_LOGIN",           // 直播内部房间登录
		"LIVE_OPEN_PLATFORM_GAME",            // 直播开放平台游戏
		"WATCHED_CHANGE",                     // 观看变化
		"DANMU_AGGREGATION",                  // 弹幕聚合
		"POPULARITY_RED_POCKET_NEW",          // 新人气红包
		"LIKE_INFO_V3_CLICK",                 // 点赞信息V3点击
		"POPULAR_RANK_CHANGED",               // 人气排名变化
		"DM_INTERACTION",                     // 弹幕互动
		"LIKE_INFO_V3_UPDATE",                // 点赞信息V3更新
		"HOT_ROOM_NOTIFY",                    // 热门房间通知
		"PLAY_TAG",                           // 播放标签
	}
	knownCMDMap map[string]int
	cmdReg      = regexp.MustCompile(`"cmd":"([^"]+)"`)
)

type eventHandlers struct {
	danmakuMessageHandlers []func(*message.Danmaku)
	superChatHandlers      []func(*message.SuperChat)
	giftHandlers           []func(*message.Gift)
	guardBuyHandlers       []func(*message.GuardBuy)
	liveStartHandlers      []func(start *message.LiveStart)
	liveStopHandlers       []func(start *message.LiveStop)
	userToastHandlers      []func(*message.UserToast)
}

type customEventHandlers map[string]func(s string)

func init() {
	knownCMDMap = make(map[string]int)
	for _, c := range knownCMD {
		knownCMDMap[c] = 0
	}
}

// RegisterCustomEventHandler 注册 自定义事件 的处理器
//
// 需要提供事件名，可参考 knownCMD
func (c *Client) RegisterCustomEventHandler(cmd string, handler func(s string)) {
	(*c.customEventHandlers)[cmd] = handler
}

// OnDanmaku 添加 弹幕事件 的处理器
func (c *Client) OnDanmaku(f func(*message.Danmaku)) {
	c.eventHandlers.danmakuMessageHandlers = append(c.eventHandlers.danmakuMessageHandlers, f)
}

// OnSuperChat 添加 醒目留言事件 的处理器
func (c *Client) OnSuperChat(f func(*message.SuperChat)) {
	c.eventHandlers.superChatHandlers = append(c.eventHandlers.superChatHandlers, f)
}

// OnGift 添加 礼物事件 的处理器
func (c *Client) OnGift(f func(gift *message.Gift)) {
	c.eventHandlers.giftHandlers = append(c.eventHandlers.giftHandlers, f)
}

// OnGuardBuy 添加 开通大航海事件 的处理器
func (c *Client) OnGuardBuy(f func(*message.GuardBuy)) {
	c.eventHandlers.guardBuyHandlers = append(c.eventHandlers.guardBuyHandlers, f)
}

// OnLiveStart 添加 开播事件 的处理器
func (c *Client) OnLiveStart(f func(start *message.LiveStart)) {
	c.eventHandlers.liveStartHandlers = append(c.eventHandlers.liveStartHandlers, f)
}

// OnLiveStop 添加 关播事件 的处理器
func (c *Client) OnLiveStop(f func(start *message.LiveStop)) {
	c.eventHandlers.liveStopHandlers = append(c.eventHandlers.liveStopHandlers, f)
}

// OnUserToast 添加 UserToast 的处理
// OnUserToast 添加 UserToast 的处理器
func (c *Client) OnUserToast(f func(*message.UserToast)) {
	c.eventHandlers.userToastHandlers = append(c.eventHandlers.userToastHandlers, f)
}

// Handle 处理一个包
func (c *Client) Handle(p packet.Packet) {
	switch p.Operation {
	case packet.Notification:
		cmd := parseCmd(p.Body)
		sb := utils.BytesToString(p.Body)
		// 新的弹幕 cmd 可能带参数
		if ind := strings.Index(cmd, ":"); ind >= 0 {
			cmd = cmd[:ind]
		}
		// 优先执行自定义 eventHandler ，会覆盖库内自带的 handler
		f, ok := (*c.customEventHandlers)[cmd]
		if ok {
			go cover(func() { f(sb) })
			return
		}
		switch cmd {
		// 弹幕
		case "DANMU_MSG":
			d := new(message.Danmaku)
			d.Parse(p.Body)
			for _, fn := range c.eventHandlers.danmakuMessageHandlers {
				go cover(func() { fn(d) })
			}
		// 醒目留言
		case "SUPER_CHAT_MESSAGE":
			s := new(message.SuperChat)
			s.Parse(p.Body)
			for _, fn := range c.eventHandlers.superChatHandlers {
				go cover(func() { fn(s) })
			}
		// 礼物
		case "SEND_GIFT":
			g := new(message.Gift)
			g.Parse(p.Body)
			for _, fn := range c.eventHandlers.giftHandlers {
				go cover(func() { fn(g) })
			}
		// 大航海
		case "GUARD_BUY":
			g := new(message.GuardBuy)
			g.Parse(p.Body)
			for _, fn := range c.eventHandlers.guardBuyHandlers {
				go cover(func() { fn(g) })
			}
		// 开播
		case "LIVE":
			l := new(message.LiveStart)
			l.Parse(p.Body)
			for _, fn := range c.eventHandlers.liveStartHandlers {
				go cover(func() { fn(l) })
			}
		//下播
		case "PREPARING":
			l := new(message.LiveStop)
			l.Parse(p.Body)
			for _, fn := range c.eventHandlers.liveStopHandlers {
				go cover(func() { fn(l) })
			}
		// 用户 toast
		case "USER_TOAST_MSG":
			u := new(message.UserToast)
			u.Parse(p.Body)
			for _, fn := range c.eventHandlers.userToastHandlers {
				go cover(func() { fn(u) })
			}
		default:
			if _, ok := knownCMDMap[cmd]; ok {
				return
			}
			log.Errorf("unknown cmd(%s), body: %s", cmd, p.Body)
		}
	case packet.HeartBeatResponse:
	case packet.RoomEnterResponse:
	default:
		log.Alert(fmt.Sprintf("protover: %v data: %v unknown protover", p.ProtocolVersion, p.Body))
	}
}

// parseCmd 获取 JSON 报文的 CMD
func parseCmd(d []byte) string {
	// {"cmd":"DANMU_MSG", ...
	str := utils.BytesToString(d)
	match := cmdReg.FindStringSubmatch(str)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func cover(f func()) {
	defer func() {
		if pan := recover(); pan != nil {
			log.Errorf("event error: %v\n%s", pan, debug.Stack())
		}
	}()
	f()
}
