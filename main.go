package main

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/alexferrari88/gohn/pkg/gohn"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const PageSize int = 5

var state = make(map[int64]int)

// Send any text message to the bot after the bot has been started

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(DefaultHandler),
		bot.WithCallbackQueryDataHandler("page:", bot.MatchTypePrefix, PageHandler),
		bot.WithCallbackQueryDataHandler("story:", bot.MatchTypePrefix, StoryHandler),
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		panic("token not provided")
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		fmt.Println("panic in creating bot")
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/top", bot.MatchTypeExact, TopStoriesHandler)
	go func() {
		http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello World")
		})
		http.ListenAndServe(":10000", nil)
	}()
	
	b.Start(ctx)
}


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

func getTopComments(ctx context.Context, storyId int) []string {
	hn, _ := gohn.NewClient(nil)
	story, err := hn.Items.Get(ctx, storyId)
	if err != nil {
		panic("failed to retrieve story")
	}

	limit := 5
	var sb strings.Builder
	var res []string
	if *story.Kids != nil && len(*story.Kids) < limit {
		limit = len(*story.Kids)
	}
	for i := 0; i < limit; i++ {
		sb.Reset()
		comment, err := hn.Items.Get(ctx, (*story.Kids)[i])
		if err != nil {
			panic("failed to fetch comment")
		}
		finalString := strings.ReplaceAll(*comment.Text, "<p>", "\n\n")
		finalString = html.UnescapeString(finalString)
		// finalString = strings.ReplaceAll(finalString, "<", "&lt;")
		// finalString = strings.ReplaceAll(finalString, ">", "&gt;")
		// finalString = strings.ReplaceAll(finalString, "&", "&amp;")
		sb.WriteString(fmt.Sprintf("<b>%s</b>\n %s\n\n", *comment.By, finalString))
		res = append(res, sb.String())
	}
	return res
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

func StoryHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	ackQuery(ctx, b, update)
	rawData := update.CallbackQuery.Data
	x := strings.Split(rawData, ":")
	storyId, _ := strconv.Atoi(x[1])
	// kb := buildInlineKeyboard([]int{1, 2, 3, 4, 5}, 10, 2)
	res := getTopComments(ctx, storyId)
	fmt.Println("sending comments", res)
	for _, msg := range res {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			Text:      msg,
			ParseMode: models.ParseModeHTML,
		})

		if err != nil {
			fmt.Printf("failed to send response: %v\n", err)
		}

	}

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
