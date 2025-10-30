package http

import (
	"bytes"
	"context"

	openai "github.com/sashabaranov/go-openai"
	"github.com/xbclub/BilibiliDanmuRobot-Core/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

func RequestDeepSeekRobot(msg string, svcCtx *svc.ServiceContext) (string, error) {
	config := openai.DefaultConfig(svcCtx.Config.DeepSeek.APIToken)
	config.BaseURL = svcCtx.Config.DeepSeek.APIUrl

	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: svcCtx.Config.DeepSeek.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: svcCtx.Config.DeepSeek.Prompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: msg,
				},
			},
			MaxTokens: 75, // 回复长度限制 粗略估算：1个中文汉字 ≈ 2-3个 tokens 限制在30个字左右
		},
	)
	if err != nil {
		return "", err
	}

	logx.Infof("本次开销：%v tokens", resp.Usage.TotalTokens)
	msgs := ""
	for _, v := range resp.Choices {
		data := []byte(v.Message.Content)
		if bytes.HasPrefix(data, []byte{239, 188, 159}) {
			data = bytes.TrimPrefix(data, []byte{239, 188, 159})
		}
		data = bytes.ReplaceAll(data, []byte{10, 10}, []byte{})
		msgs += string(data)
	}
	return msgs, nil
}
