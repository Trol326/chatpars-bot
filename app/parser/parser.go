package parser

import (
	"chatparser-bot/tools"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/sync/errgroup"
)

type ParsedMessages []*Message

type Message struct {
	Text    string   `json:"message"`
	Attachs []string `json:"attachments"`
}

const maxMessages int = 150000

// TODO add error tracing
func ParseMessagesInChannels(discord *discordgo.Session, pars *tools.Parsings, notifChannel string, channels []string) error {
	g := new(errgroup.Group)

	for _, channel := range channels {
		g.Go(func() error {
			text := fmt.Sprintf("Начинаем парсить %s", channel)
			discord.ChannelMessageSend(notifChannel, text)

			err := ParseMessagesInChannel(discord, pars, channel)
			if err != nil {
				return err
			}

			text = fmt.Sprintf("Закончил парсить %s", channel)
			discord.ChannelMessageSend(notifChannel, text)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func ParseMessagesInChannel(discord *discordgo.Session, pars *tools.Parsings, channel string) error {
	beforeID := ""
	isLast := false
	isFirst := true
	channelID := ParseChannelID(channel)
	c, err := discord.Channel(channelID)
	if err != nil {
		return err
	}
	mCounter := 0
	pars.SetCounter(channel, mCounter)

	fileName := fmt.Sprintf("%s(%s)", c.Name, c.ID)

	for !isLast {
		var ms []*discordgo.Message
		ms, err = discord.ChannelMessages(channelID, 100, beforeID, "", "")
		if err != nil {
			pars.ChangeStatus(channel, tools.StatusError)
			return err
		}

		mCounter += len(ms)
		beforeID = ms[len(ms)-1].ID
		if len(ms) < 100 || mCounter >= maxMessages {
			isLast = true
		}

		var messages []*Message
		messages = make([]*Message, 0, len(ms))
		for _, m := range ms {
			message := discordMessageToMessage(m)
			if message != nil && message.Text != "" {
				messages = append(messages, message)
			}
		}

		err = saveIntoFile(messages, fileName, isFirst, isLast)
		if err != nil {
			pars.ChangeStatus(channel, tools.StatusError)
			return err
		}
		clear(ms)
		clear(messages)
		isFirst = false

		pars.SetCounter(channel, mCounter)
	}

	pars.ChangeStatus(channel, tools.StatusCompleted)

	return nil
}

func saveIntoFile(data ParsedMessages, fileName string, isFirst, isLast bool) error {
	val, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	temp, _ := strings.CutPrefix(string(val), "[")
	temp, _ = strings.CutSuffix(temp, "]")

	val = []byte(temp)

	relPath := path.Join("result")
	err = os.MkdirAll(relPath, os.ModePerm)
	if err != nil {
		return err
	}

	relPath = path.Join(relPath, fmt.Sprintf("%s.json", fileName))
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(absPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if isFirst {
		_, err = file.WriteString("[")
		if err != nil {
			return err
		}
	} else {
		_, err = file.WriteString(",\n")
		if err != nil {
			return err
		}
	}

	_, err = file.Write(val)
	if err != nil {
		return err
	}

	if isLast {
		_, err = file.WriteString("]")
		if err != nil {
			return err
		}
	}

	clear(val)

	return nil
}

func ParseChannelID(ch string) string {
	regex, err := regexp.Compile("<#([0-9]*)")
	if err != nil {
		return ""
	}
	result := regex.FindStringSubmatch(ch)
	return result[len(result)-1]
}

func discordMessageToMessage(m *discordgo.Message) *Message {
	if m == nil {
		return nil
	}

	result := Message{}
	result.Text = m.Content

	for _, a := range m.Attachments {
		result.Attachs = append(result.Attachs, a.URL)
	}

	return &result
}
