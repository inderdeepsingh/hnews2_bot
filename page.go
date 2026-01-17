package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/alexferrari88/gohn/pkg/gohn"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func getStories(ctx context.Context, start, end int) (string, []int) {
	var sb strings.Builder
	ids := make([]int, 0)
	hn, _ := gohn.NewClient(nil)
	topStoriesIds, _ := hn.Stories.GetTopIDs(ctx)
	if len(topStoriesIds) < 5 || topStoriesIds[0] == nil {
		panic("cannot retrieve top stories")
	}
	for j := start; j < end; j++ {
		story, err := hn.Items.Get(ctx, *topStoriesIds[j])
		if err != nil {
			panic("failed to retrieve story")
		}
		sb.WriteString(fmt.Sprintf("%d. <b>%s</b> \n%d <i>by %s </i> | %d comments\n\n", j+1, *story.Title, *story.Score, *story.By, *story.Descendants))
		ids = append(ids, *topStoriesIds[j])
	}
	res := sb.String()
	return res, ids
}

func PageHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	ackQuery(ctx, b, update)
	rawData := update.CallbackQuery.Data
	x := strings.Split(rawData, ":")
	currPage, _ := strconv.Atoi(x[1])
	start := currPage * PageSize
	end := start + PageSize
	res, ids := getStories(ctx, start, end)
	kb := buildInlineKeyboard(ids, currPage*PageSize, currPage)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		Text:        res,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}