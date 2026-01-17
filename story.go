package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)



func StoryHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	ackQuery(ctx, b, update)
	rawData := update.CallbackQuery.Data
	x := strings.Split(rawData, ":")
	storyId, _ := strconv.Atoi(x[1])
	msg, hasPrev, hasNext := getComments(ctx, storyId, 0)
	fmt.Println("sending comments", msg)
	params := &bot.SendMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		Text:      msg,
		ParseMode: models.ParseModeHTML,
	}
	if hasPrev || hasNext {
		params.ReplyMarkup = buildCommentKeyboard(storyId, 0, hasPrev, hasNext)
	}
	_, err := b.SendMessage(ctx, params)
	if err != nil {
		fmt.Printf("failed to send response: %v\n", err)
	}
}
