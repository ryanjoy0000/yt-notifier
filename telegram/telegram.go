package telegram

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"

	telegramApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	// Menu texts
	firstMenu  = "<b>Menu 1</b>\n\nA menu with inline button."
	secondMenu = "<b>Menu 2</b>\n\nA menu with even more buttons."

	// Button texts
	nextButton     = "Next"
	backButton     = "Back"
	tutorialButton = "Tutorial"

	// Keyboard layout for the first menu. One button, one row
	firstMenuMarkup = telegramApi.NewInlineKeyboardMarkup(
		telegramApi.NewInlineKeyboardRow(
			telegramApi.NewInlineKeyboardButtonData(nextButton, nextButton),
		),
	)

	// Keyboard layout for the second menu. Two buttons, one per row
	secondMenuMarkup = telegramApi.NewInlineKeyboardMarkup(
		telegramApi.NewInlineKeyboardRow(
			telegramApi.NewInlineKeyboardButtonData(backButton, backButton),
		),
		telegramApi.NewInlineKeyboardRow(
			telegramApi.NewInlineKeyboardButtonURL(tutorialButton, "https://core.telegram.org/bots/api"),
		),
	)
)

type TelegramBotService struct{
    bot *telegramApi.BotAPI 
}

func NewTelegramService(apiKey string) (*TelegramBotService, error) {
    bot, err := telegramApi.NewBotAPI(apiKey)
	if err != nil {
        log.Panic("unable to create telegram bot: ", err)
        return nil, err
	}

    t := &TelegramBotService{
        bot: bot,
    }

    return t, nil
}

func (t *TelegramBotService) InitTelegramBot(canDebug bool, ctx context.Context){
   
    t.bot.Debug = canDebug

    // get updates since the last offset
	u := telegramApi.NewUpdate(0)
	u.Timeout = 60

    ctx, cancel := context.WithCancel(ctx)

	// `updatesChan` receives telegram updates
	updatesChan := t.bot.GetUpdatesChan(u)

	// Pass cancellable context to goroutine
	go t.receiveUpdates(ctx, updatesChan)

	log.Println("Bot is online. You can chat with the bot... ")

    // waiting for a newline interruption
    bufio.NewReader(os.Stdin).ReadBytes('\n')
    cancel()
}


func (t *TelegramBotService) receiveUpdates(ctx context.Context, updates telegramApi.UpdatesChannel) {
    canContinue := true
    for canContinue{
		select {
		// stop when ctx is cancelled
		case <-ctx.Done():
            canContinue = false
			return
		// receive update from channel and then handle it
		case update := <-updates:
			t.handleUpdate(update)
		}
	}
}

func (t *TelegramBotService)handleUpdate(update telegramApi.Update) {
	switch {
        // Handle messages
        case update.Message != nil:
            t.handleMessage(update.Message)
            break

        // Handle button clicks
        case update.CallbackQuery != nil:
            t.handleButton(update.CallbackQuery)
            break
	}
}

func (t *TelegramBotService)handleMessage(message *telegramApi.Message) {
	user := message.From
	text := message.Text

	if user == nil {
		return
	}

	// Print to console
	log.Printf("%s wrote %s", user.FirstName, text)

	var err error
	if strings.HasPrefix(text, "/") {
		err = t.handleCommand(message.Chat.ID, text)
        if err != nil {
            log.Println("unable to handle command in telegram bot", err)
        }
	} else if len(text) > 0 {
		msg := telegramApi.NewMessage(message.Chat.ID, text)
		// To preserve markdown, we attach entities (bold, italic..)
		msg.Entities = message.Entities
		_, err = t.bot.Send(msg)
        if err != nil {
            log.Println("unable to send msg using telegram bot", err)
        }
	} else {
		// This is equivalent to forwarding, without the sender's name
		copyMsg := telegramApi.NewCopyMessage(message.Chat.ID, message.Chat.ID, message.MessageID)
		_, err = t.bot.CopyMessage(copyMsg)
	}
}

// When we get a command, we react accordingly
func (t *TelegramBotService)handleCommand(chatId int64, command string) error {
	var err error

	switch command {
        case "/menu":
            err = t.sendMenu(chatId)
            break
        }

	return err
}

func (t *TelegramBotService)handleButton(query *telegramApi.CallbackQuery) {
	var text string

	markup := telegramApi.NewInlineKeyboardMarkup()
	message := query.Message

	if query.Data == nextButton {
		text = secondMenu
		markup = secondMenuMarkup
	} else if query.Data == backButton {
		text = firstMenu
		markup = firstMenuMarkup
	}

	callbackCfg := telegramApi.NewCallback(query.ID, "")
	t.bot.Send(callbackCfg)

	// Replace menu text and keyboard
	msg := telegramApi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, markup)
	msg.ParseMode = telegramApi.ModeHTML
	t.bot.Send(msg)
}

func (t *TelegramBotService)sendMenu(chatId int64) error {
	msg := telegramApi.NewMessage(chatId, firstMenu)
	msg.ParseMode = telegramApi.ModeHTML
	msg.ReplyMarkup = firstMenuMarkup
	_, err := t.bot.Send(msg)
	return err
}
