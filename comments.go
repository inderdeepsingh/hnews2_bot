package main

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/alexferrari88/gohn/pkg/gohn"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func buildCommentKeyboard(storyId int, start int, hasPrev, hasNext bool) *models.InlineKeyboardMarkup {
	var buttons []models.InlineKeyboardButton
	if hasPrev {
		buttons = append(buttons, models.InlineKeyboardButton{
			Text:         "prev",
			CallbackData: fmt.Sprintf("comments:%d:%d", storyId, start-5),
		})
	}
	if hasNext {
		buttons = append(buttons, models.InlineKeyboardButton{
			Text:         "next",
			CallbackData: fmt.Sprintf("comments:%d:%d", storyId, start+5),
		})
	}
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{buttons},
	}
}

func getComments(ctx context.Context, storyId int, start int) (string, bool, bool) {
	hn, _ := gohn.NewClient(nil)
	story, err := hn.Items.Get(ctx, storyId)
	if err != nil {
		panic("failed to retrieve story")
	}

	totalComments := 0
	if story.Kids != nil {
		totalComments = len(*story.Kids)
	}

	end := start + 5
	if end > totalComments {
		end = totalComments
	}

	var sb strings.Builder
	for i := start; i < end; i++ {
		comment, err := hn.Items.Get(ctx, (*story.Kids)[i])
		if err != nil {
			panic("failed to fetch comment")
		}
		finalString := strings.ReplaceAll(*comment.Text, "<p>", "\n\n")
		finalString = html.UnescapeString(finalString)
		sb.WriteString(fmt.Sprintf("<b>%s</b>\n%s\n\n---\n\n", *comment.By, finalString))
	}

	hasPrev := start > 0
	hasNext := end < totalComments
	return sb.String(), hasPrev, hasNext
}

func CommentsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	ackQuery(ctx, b, update)
	rawData := update.CallbackQuery.Data
	x := strings.Split(rawData, ":")
	storyId, _ := strconv.Atoi(x[1])
	start, _ := strconv.Atoi(x[2])
	msg, hasPrev, hasNext := getComments(ctx, storyId, start)
	fmt.Println("sending comments", msg)
	params := &bot.SendMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		Text:      msg,
		ParseMode: models.ParseModeHTML,
	}
	if hasPrev || hasNext {
		params.ReplyMarkup = buildCommentKeyboard(storyId, start, hasPrev, hasNext)
	}
	_, err := b.SendMessage(ctx, params)
	if err != nil {
		fmt.Printf("failed to send response: %v\n", err)
	}
}
