package main

import (
	"chatparser-bot/app/parser"
	"chatparser-bot/tools"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

type Client struct {
	Session        *discordgo.Session
	Log            *zerolog.Logger
	ActiveParsings *tools.Parsings
}

// TODO add database for parsing resume after reloading
func main() {
	absPath, err := filepath.Abs(".env")
	if err != nil {
		fmt.Printf("Error on path to .env creation, %s\n", err)
		return
	}
	if err := godotenv.Load(absPath); err != nil {
		fmt.Printf("Cannot read .env: %s\n", err)
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC822}
	log := zerolog.New(consoleWriter).With().Timestamp().Logger()

	key := os.Getenv("KEY")
	if key == "" {
		log.Err(fmt.Errorf("bot key not found"))
		return
	}

	data := fmt.Sprintf("Bot %s", key)
	s, err := discordgo.New(data)
	if err != nil {
		return
	}

	bot := Client{Session: s, Log: &log, ActiveParsings: tools.NewParsings()}

	bot.Session.AddHandler(bot.OnNewMessage)

	err = bot.Session.Open()
	if err != nil {
		log.Err(err).Msg("[bot.Start]Error on session open")
		return
	}
	defer bot.Session.Close()

	bot.Log.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	bot.Log.Info().Msg("Bot is shutting down...")
}

func (c Client) OnNewMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == discord.State.User.ID {
		return
	}

	prefix := "pr!"

	switch {
	case strings.HasPrefix(message.Content, prefix+"ping"):
		_, err := discord.ChannelMessageSend(message.ChannelID, "pong")
		if err != nil {
			c.Log.Err(err).Msg("[OnNewMessage]error on channelMessageSend")
			return
		}
	case strings.HasPrefix(message.Content, prefix+"pars"):
		parsed, err := ParseCommand(prefix, message.Content)
		if err != nil {
			c.Log.Err(err).Msg("[OnNewMessage]error on message parsing")
			text := fmt.Sprintf("Ошибка при парсинге. Пример использования команды: `%spars #канал`", prefix)
			_, errM := discord.ChannelMessageSend(message.ChannelID, text)
			if errM != nil {
				c.Log.Err(errM).Msg("[OnNewMessage]error on channelMessageSend")
			}
			return
		}
		if len(parsed) <= 2 {
			c.Log.Err(err).Msg("[OnNewMessage]not enough arguments")
			text := fmt.Sprintf("Ошибка при парсинге. Пример использования команды:  `%spars #канал`", prefix)
			_, errM := discord.ChannelMessageSend(message.ChannelID, text)
			if errM != nil {
				c.Log.Err(errM).Msg("[OnNewMessage]error on channelMessageSend")
			}
			return
		}
		channels := strings.Split(parsed[2], " ")
		text := fmt.Sprintf("Запускаем парсинг каналов... Порядок таков: %s", WriteChannels(channels))
		_, errM := discord.ChannelMessageSend(message.ChannelID, text)
		if errM != nil {
			c.Log.Err(errM).Msg("[OnNewMessage]error on channelMessageSend")
		}

		err = parser.ParseMessagesInChannels(discord, c.ActiveParsings, message.ChannelID, channels)
		if err != nil {
			c.Log.Err(err).Msg("[ParseMessagesInChannel]error on parsStart")
			_, errM := discord.ChannelMessageSend(message.ChannelID, "Ошибка при запуске парсинга, проверьте логи")
			if errM != nil {
				c.Log.Err(errM).Msg("[OnNewMessage]error on channelMessageSend")
			}
			return
		}
		_, errM = discord.ChannelMessageSend(message.ChannelID, "О. Парсинг завершён")
		if errM != nil {
			c.Log.Err(errM).Msg("[OnNewMessage]error on channelMessageSend")
		}
	case strings.HasPrefix(message.Content, prefix+"status"):
		t := ""
		parsKeys := c.ActiveParsings.Keys()
		for _, key := range parsKeys {
			n, ok := c.ActiveParsings.Get(key)
			if ok {
				t = fmt.Sprintf("%s\n%s - %d / ??? (%s)", t, key, n.Counter, n.Status)
			}
		}

		if t == "" {
			t = "\nНет данных о парсингах"
		}

		text := fmt.Sprintf("Статус парсингов:%s", t)
		_, err := discord.ChannelMessageSend(message.ChannelID, text)
		if err != nil {
			c.Log.Err(err).Msg("[OnNewMessage]error on channelMessageSend")
			return
		}
	}
}

func ParseCommand(prefix, command string) ([]string, error) {
	p := regexp.QuoteMeta(prefix)
	exp := fmt.Sprintf("(%s[a-z]* )(.*)", p)
	regex, err := regexp.Compile(exp)
	if err != nil {
		return nil, err
	}
	return regex.FindStringSubmatch(command), nil
}

func WriteChannels(channels []string) string {
	result := ""
	for _, c := range channels {
		if result == "" {
			result = c
		} else {
			result = fmt.Sprintf("%s, %s", result, c)
		}
	}

	return result
}
