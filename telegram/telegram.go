package telegram

import (
	"context"
	"log"
	"net/url"
	"strconv"
	"strings"

	telegramApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const(
    
	// Menu texts
	firstMenu  = "<b>Menu 1</b>\n\nA menu with inline button."
	secondMenu = "<b>Menu 2</b>\n\nA menu with even more buttons."

	// Button texts
	nextButton     = "Next"
	backButton     = "Back"
	tutorialButton = "Tutorial"


    WELCOME_MSG = "Welcome! I will help you to keep track of your favourite YouTube videos. I can notify you in case there is a change in the likes / comments in the videos. I will need the YouTube Playlist URL to proceed. Just make sure the playlist is public."
    YOUTUBE_URL = "https://youtube.com/"
    ACK_1 = "Thank you. I will track this playlist and let you know if there are any changes in the likes / comments of the videos"
)

var (
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
    Bot *telegramApi.BotAPI
    PlaylistIDChan chan string
    ClientTelegramId chan string
    WaitForUserInput bool
}

func NewTelegramService(apiKey string) (*TelegramBotService, error) {
    bot, err := telegramApi.NewBotAPI(apiKey)
	if err != nil {
        log.Panic("unable to create telegram bot: ", err)
        return nil, err
	}

    t := &TelegramBotService{
        Bot: bot,
        PlaylistIDChan: make(chan string),
        ClientTelegramId: make(chan string),
        WaitForUserInput: false,
    }

    return t, nil
}

func (t *TelegramBotService) InitTelegramBot(canDebug bool, ctx context.Context){
   
    t.Bot.Debug = canDebug

    // get updates since the last offset
	u := telegramApi.NewUpdate(0)
	u.Timeout = 60

	// `updatesChan` receives telegram updates
	updatesChan := t.Bot.GetUpdatesChan(u)

	// Pass cancellable context to goroutine
	go t.receiveUpdates(ctx, updatesChan)

	log.Println("Bot is online... ")
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

    // grab client id
    clientId := strconv.FormatInt(message.Chat.ID, 10)
    log.Println("client telegram id: ", clientId)

	var err error
	if strings.HasPrefix(text, "/") {
		err = t.handleCommand(message.Chat.ID, text)
        if err != nil {
            log.Println("unable to handle command in telegram bot", err)
        }
	} else if strings.HasPrefix(text, YOUTUBE_URL) && t.WaitForUserInput{
        plId , _ := extractPlaylistId(text);
        t.PlaylistIDChan<- plId
        t.ClientTelegramId<- clientId 
        t.WaitForUserInput = false
        resp := telegramApi.NewMessage(message.Chat.ID, ACK_1)
        _, err = t.Bot.Send(resp)
        if err != nil {
            log.Println("unable to send msg using telegram bot", err)
        }
    }else if len(text) > 0 {
		msg := telegramApi.NewMessage(message.Chat.ID, text)
		// To preserve markdown, we attach entities (bold, italic..)
		msg.Entities = message.Entities
		_, err = t.Bot.Send(msg)
        if err != nil {
            log.Println("unable to send msg using telegram bot", err)
        }
	} else {
		// This is equivalent to forwarding, without the sender's name
		copyMsg := telegramApi.NewCopyMessage(message.Chat.ID, message.Chat.ID, message.MessageID)
		_, err = t.Bot.CopyMessage(copyMsg)
	}
}

// When we get a command, we react accordingly
func (t *TelegramBotService)handleCommand(chatId int64, command string) error {
	var err error

	switch command {
        case "/start":
            err = t.handleStartCommand(chatId)
            break
        // case "/menu":
            // err = t.sendMenu(chatId)
            // break
        }

	return err
}

// handle start command
func (t *TelegramBotService)handleStartCommand(chatId int64) error {
	msg := telegramApi.NewMessage(chatId, WELCOME_MSG)
	_, err := t.Bot.Send(msg)
    t.WaitForUserInput = true
	return err
}


// extract playlistID 
func extractPlaylistId(u string) (string, error) {
    id := ""
    urlPtr, err := url.Parse(u)
    if err!= nil{
        log.Println("unable to parse url string", err)
    }
    valMap, err := url.ParseQuery(urlPtr.RawQuery)
    id = valMap["list"][0]
    log.Println("extracted playlist id: ", id)
    return id, err
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
	t.Bot.Send(callbackCfg)

	// Replace menu text and keyboard
	msg := telegramApi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, markup)
	msg.ParseMode = telegramApi.ModeHTML
	t.Bot.Send(msg)
}

func (t *TelegramBotService)sendMenu(chatId int64) error {
	msg := telegramApi.NewMessage(chatId, firstMenu)
	msg.ParseMode = telegramApi.ModeHTML
	msg.ReplyMarkup = firstMenuMarkup
	_, err := t.Bot.Send(msg)
	return err
}
