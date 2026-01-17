package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const PageSize int = 5

var state = make(map[int64]int)

// Send any text message to the bot after the bot has been started

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		panic("token not provided")
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(DefaultHandler),
		// bot.WithWebhookSecretToken(token),
		bot.WithCallbackQueryDataHandler("page:", bot.MatchTypePrefix, PageHandler),
		bot.WithCallbackQueryDataHandler("story:", bot.MatchTypePrefix, StoryHandler),
		bot.WithCallbackQueryDataHandler("comments:", bot.MatchTypePrefix, CommentsHandler),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		fmt.Println("panic in creating bot")
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/top", bot.MatchTypeExact, TopStoriesHandler)

	if os.Getenv("RENDER") == "true" {
		url := os.Getenv("RENDER_EXTERNAL_URL")
		_, err := b.SetWebhook(ctx, &bot.SetWebhookParams{
			URL: url,
		})
		if err != nil {
			panic("failed to set webhook url")
		}
		defer b.DeleteWebhook(ctx, &bot.DeleteWebhookParams{
			DropPendingUpdates: true,
		})
		go b.StartWebhook(ctx)
		http.ListenAndServe(":10000", b.WebhookHandler())
	} else {
		b.Start(ctx)
	}
}

func buildInlineKeyboard(stories []int, start, currPage int) *models.InlineKeyboardMarkup {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: fmt.Sprintf("%d", start+1), CallbackData: "story:" + strconv.Itoa(stories[0])},
				{Text: fmt.Sprintf("%d", start+2), CallbackData: "story:" + strconv.Itoa(stories[1])},
				{Text: fmt.Sprintf("%d", start+3), CallbackData: "story:" + strconv.Itoa(stories[2])},
				{Text: fmt.Sprintf("%d", start+4), CallbackData: "story:" + strconv.Itoa(stories[3])},
				{Text: fmt.Sprintf("%d", start+5), CallbackData: "story:" + strconv.Itoa(stories[4])},
			}, {
				{Text: "prev", CallbackData: "page:" + strconv.Itoa(currPage-1)},
				{Text: "next", CallbackData: "page:" + strconv.Itoa(currPage+1)},
			},
		},
	}
	return kb
}

func ackQuery(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})
}

func TopStoriesHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	res, ids := getStories(ctx, 0, PageSize)
	kb := buildInlineKeyboard(ids, 0, 0)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        res,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}



func DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("here")
	fmt.Println(update.Message)
	fmt.Println(update.CallbackQuery)
	if update.Message != nil {
		fmt.Println("got message")
		fmt.Println(update.Message)
		return
	}
	if update.CallbackQuery != nil {
		fmt.Println("got callbackquery")
		fmt.Println(update.CallbackQuery.Data)
		return
	}
}
