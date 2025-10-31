package http

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"
	"github.com/xbclub/BilibiliDanmuRobot-Core/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// 为每个房间维护一个对话历史记录
var roomConversationHistory = make(map[int][]openai.ChatCompletionMessage)
var historyMutex = sync.Mutex{}

// getHistoryFilePath 获取历史记录文件路径
func getHistoryFilePath(roomID int) string {
	// 确保目录存在
	dir := "./db/history"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	return filepath.Join(dir, "room_"+strconv.Itoa(roomID)+".json")
}

// isValidMessage 检查消息是否有效
func isValidMessage(message openai.ChatCompletionMessage) bool {
	return message.Role != "" && message.Content != ""
}

// filterValidMessages 过滤掉无效的消息
func filterValidMessages(messages []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	validMessages := make([]openai.ChatCompletionMessage, 0)
	for _, message := range messages {
		if isValidMessage(message) {
			validMessages = append(validMessages, message)
		}
	}
	return validMessages
}

// removeDuplicateMessages 去除重复的消息
func removeDuplicateMessages(messages []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	if len(messages) <= 1 {
		return messages
	}

	// 使用map来检测重复内容
	seen := make(map[string]bool)
	result := make([]openai.ChatCompletionMessage, 0)

	for _, message := range messages {
		key := message.Role + ":" + message.Content
		if !seen[key] {
			seen[key] = true
			result = append(result, message)
		}
	}

	return result
}

// loadHistoryFromFile 从文件加载历史记录
func loadHistoryFromFile(roomID int) []openai.ChatCompletionMessage {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	// 如果内存中已有记录，直接返回
	if history, exists := roomConversationHistory[roomID]; exists {
		return filterValidMessages(history)
	}

	// 从文件加载
	filePath := getHistoryFilePath(roomID)
	data, err := os.ReadFile(filePath)
	if err != nil {
		// 文件不存在或其他错误，返回空历史记录
		roomConversationHistory[roomID] = make([]openai.ChatCompletionMessage, 0)
		return roomConversationHistory[roomID]
	}

	var history []openai.ChatCompletionMessage
	if err := json.Unmarshal(data, &history); err != nil {
		// 解析失败，返回空历史记录
		roomConversationHistory[roomID] = make([]openai.ChatCompletionMessage, 0)
		return roomConversationHistory[roomID]
	}

	// 过滤掉无效的消息
	history = filterValidMessages(history)

	// 保存到内存并返回
	roomConversationHistory[roomID] = history
	return history
}

// saveHistoryToFile 将历史记录保存到文件
func saveHistoryToFile(roomID int, history []openai.ChatCompletionMessage, maxHistoryMessages int) {
	// 去除重复消息（只在保存时去重）
	history = removeDuplicateMessages(history)

	// 限制历史记录长度，防止文件过大
	if len(history) > maxHistoryMessages {
		history = history[len(history)-maxHistoryMessages:]
	}

	data, err := json.Marshal(history)
	if err != nil {
		logx.Errorf("序列化历史记录失败: %v", err)
		return
	}

	filePath := getHistoryFilePath(roomID)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		logx.Errorf("保存历史记录到文件失败: %v", err)
		return
	}

	// 同时更新内存中的记录
	historyMutex.Lock()
	roomConversationHistory[roomID] = history
	historyMutex.Unlock()
}

// buildMessagesWithHistory 构建包含历史对话的消息列表
func buildMessagesWithHistory(svcCtx *svc.ServiceContext) []openai.ChatCompletionMessage {
	roomID := svcCtx.Config.RoomId

	// 从文件加载历史记录
	history := loadHistoryFromFile(roomID)

	// 构建系统提示词
	systemPrompt := svcCtx.Config.DeepSeek.Prompt
	systemPrompt += "，说话简明扼要！30字以内！不要截断！" // 强制要求回复长度
	
	// 添加屏蔽词提示到系统提示词
	if len(svcCtx.Config.DeepSeek.BlockedWords) > 0 {
		systemPrompt += " 请注意避免使用以下屏蔽词: " + strings.Join(svcCtx.Config.DeepSeek.BlockedWords, ", ")
	}
	
	systemMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
	}

	// 构建消息列表，包括系统提示词和历史对话
	messages := make([]openai.ChatCompletionMessage, 0)
	messages = append(messages, systemMessage)
	messages = append(messages, history...)

	return messages
}

// updateConversationHistory 更新对话历史记录
func updateConversationHistory(roomID int, userMsg, aiReply string, maxHistoryMessages int) {
	// 获取当前房间的历史对话记录
	history := loadHistoryFromFile(roomID)

	// 将历史数据中的 user 替换成 assistant，确保角色正确
	for i, msg := range history {
		if msg.Role == openai.ChatMessageRoleUser {
			history[i].Role = openai.ChatMessageRoleAssistant
		}
	}

	// 添加用户消息到历史记录
	history = append(history, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userMsg,
	})

	// 添加AI回复到历史记录（仅当回复非空时）
	if aiReply != "" {
		history = append(history, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: aiReply,
		})
	}

	// 保存到文件
	saveHistoryToFile(roomID, history, maxHistoryMessages)
}

func RequestDeepSeekRobot(msg string, svcCtx *svc.ServiceContext) (string, error) {
	config := openai.DefaultConfig(svcCtx.Config.DeepSeek.APIToken)
	config.BaseURL = svcCtx.Config.DeepSeek.APIUrl

	client := openai.NewClientWithConfig(config)

	// 获取当前房间ID和最大历史记录数
	roomID := svcCtx.Config.RoomId
	maxHistoryMessages := svcCtx.Config.DeepSeek.MaxHistoryMessages

	// 添加用户消息到历史记录
	updateConversationHistory(roomID, msg, "", maxHistoryMessages)

	// 构建包含历史对话的消息列表
	messages := buildMessagesWithHistory(svcCtx)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:               svcCtx.Config.DeepSeek.Model,
			Messages:            messages,
			MaxTokens:           100, // 回复长度限制 粗略估算：1个中文汉字 ≈ 2-3个 tokens
			MaxCompletionTokens: 100, // 最大生成 tokens 数
			Temperature:         0.8, // 生成文本的随机程度，范围是0到1，值越高，生成的文本越随机
		},
	)
	if err != nil {
		// 发生错误时，从历史记录中移除刚才添加的用户消息
		history := loadHistoryFromFile(roomID)
		if len(history) > 0 {
			history = history[:len(history)-1]
			saveHistoryToFile(roomID, history, maxHistoryMessages)
		}
		return "", err
	}

	// 提取回复内容
	reply := ""
	if len(resp.Choices) > 0 {
		reply = resp.Choices[0].Message.Content
	}

	// 更新AI回复到历史记录中
	updateConversationHistory(roomID, msg, reply, maxHistoryMessages)

	return reply, nil
}