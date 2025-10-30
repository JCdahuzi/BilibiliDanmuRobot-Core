package http

import (
	"bytes"
	"context"
	"io"

	openai "github.com/sashabaranov/go-openai"
	"github.com/xbclub/BilibiliDanmuRobot-Core/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

func RequestDeepSeekRobot(msg string, svcCtx *svc.ServiceContext) (string, error) {
	config := openai.DefaultConfig(svcCtx.Config.DeepSeek.APIToken)
	config.BaseURL = svcCtx.Config.DeepSeek.APIUrl

	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletionStream(
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
			MaxTokens:   100, // 回复长度限制 粗略估算：1个中文汉字 ≈ 2-3个 tokens 限制在30个字左右
			Temperature: 0.8, // 生成文本的随机程度，范围是0到1，值越高，生成的文本越随机
			// Stream:      true, // 流式响应
		},
	)
	if err != nil {
		return "", err
	}
	logx.Infof("本次开销：%v tokens", resp.GetRateLimitHeaders().LimitTokens)
	msgs := ""
	defer resp.Close()
	for {
		streamResp, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logx.Error(err)
			return "", err
		}

		var content string
		if len(streamResp.Choices) > 0 {
			for i := range streamResp.Choices {
				if streamResp.Choices[i].Delta.Content != "" {
					content = streamResp.Choices[i].Delta.Content
					data := bytes.TrimSpace([]byte(content))
					// 避免重复推送
					if len(msgs) > 0 && string(data) == string(msgs[len(msgs)-1]) {
						continue
					}
					msgs += string(data)
				} else {
					continue
				}
			}
		}
	}
	return msgs, nil
}
