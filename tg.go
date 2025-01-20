package main

import (
	"log"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	s "github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func main() {
	bot, err := tg.NewBotAPI("токен")
	if err != nil {
		log.Println("Ошибка токена: ", err)
		return
	}

	u := tg.NewUpdate(60)
	u.Timeout = 0
	updates := bot.GetUpdatesChan(u)

	keyboard := tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton("Отправить случайный комплимент"),
			tg.NewKeyboardButton("Случайный пароль"),
		),
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton("Просто кнопка"),
			tg.NewKeyboardButton("Просто кнопка"),
		),
	)

	for update := range updates {
		if update.Message != nil {
			handleUpdate(update, bot, keyboard)
		}
	}
}

func handleUpdate(update tg.Update, bot *tg.BotAPI, keyboard tg.ReplyKeyboardMarkup) {
	switch update.Message.Text {
	case "Отправить случайный комплимент":
		handleAction(update, bot, parseCompliment)
	case "Случайный пароль":
		handleAction(update, bot, randomPassword)
	case "Просто кнопка":
		sendMessage("Это кнопка без действия.", update, bot)
	default:
		sendMessage("Нет такой команды.", update, bot)
	}

	msg := tg.NewMessage(update.Message.Chat.ID, "Что дальше?")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func handleAction(update tg.Update, bot *tg.BotAPI, actionFunc func() string) {
	press := sendMessage("Подождите...", update, bot)
	result := actionFunc()
	deleteMessage(update.Message.Chat.ID, press.MessageID, bot)
	sendMessage(result, update, bot)
}

func sendMessage(text string, update tg.Update, bot *tg.BotAPI) tg.Message {
	msg := tg.NewMessage(update.Message.Chat.ID, text)
	gotIt, _ := bot.Send(msg)
	return gotIt
}

func deleteMessage(chatID int64, messageID int, bot *tg.BotAPI) {
	delete := tg.NewDeleteMessage(chatID, messageID)
	bot.Request(delete)
}

func setupWebDriver() (s.WebDriver, func(), error) {
	service, err := s.NewChromeDriverService("./chromedriver", 4444)
	if err != nil {
		log.Println("Не удается запустить драйвер: ", err)
		return nil, nil, err
	}

	cleanup := func() {
		service.Stop()
	}

	caps := s.Capabilities{"browserName": "chrome"}
	caps.AddChrome(chrome.Capabilities{
		Args: []string{
			"window-size=1920x1080",
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-blink-features=AutomationControlled",
			"--disable-infobars",
		},
	})

	driver, err := s.NewRemote(caps, "")
	if err != nil {
		log.Println("Не удается подключиться к драйверу: ", err)
		cleanup()
		return nil, nil, err
	}

	return driver, cleanup, nil
}

func parseCompliment() string {
	driver, cleanup, err := setupWebDriver()
	if err != nil {
		return "Не удалось инициализировать драйвер"
	}
	defer cleanup()
	err = driver.Get("https://www.hse.ru/n/valentine/")
	if err != nil {
		log.Println("Не удалось открыть страницу")
		return "Не удалось открыть страницу"
	}
	defer driver.Close()

	button, err := driver.FindElement(s.ByCSSSelector, "button")
	if err != nil {
		log.Println("Не удалось найти кнопку: ", err)
		return "Ошибка загрузки данных"
	}

	if err := button.Click(); err != nil {
		log.Println("Не удалось нажать на кнопку: ", err)
		return "Ошибка загрузки данных"
	}

	p, err := driver.FindElement(s.ByCSSSelector, "p.promise")
	if err != nil {
		log.Println("Не удалось найти текст: ", err)
		return "Ошибка загрузки данных"
	}

	text, _ := p.Text()
	return text
}

func randomPassword() string {
	driver, cleanup, err := setupWebDriver()
	if err != nil {
		return "Не удалось инициализировать драйвер"
	}
	defer cleanup()
	err = driver.Get("https://randstuff.ru/password/")
	if err != nil {
		log.Println("Не удалось открыть страницу")
		return "Не удалось открыть страницу"
	}
	defer driver.Close()

	span, err := driver.FindElement(s.ByCSSSelector, "span.cur")
	if err != nil {
		log.Println("Не удалось найти текст: ", err)
		return "Ошибка загрузки данных"
	}

	password, _ := span.Text()
	return password
}