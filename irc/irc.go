package main

import (
	"os"
	"strings"

	"github.com/cooper/quiki/wikifier"
	hbot "github.com/whyrusleeping/hellabot"
)

func main() {
	mybot, err := hbot.NewBot(os.Args[1], "quiki")
	mybot.Channels = []string{os.Args[2]}
	if err != nil {
		panic(err)
	}

	mybot.AddTrigger(hbot.Trigger{
		func(bot *hbot.Bot, mes *hbot.Message) bool {
			return mes.Command == "PRIVMSG" && strings.HasPrefix(mes.Content, "quiki")
		},
		func(irc *hbot.Bot, mes *hbot.Message) bool {
			line := strings.TrimLeft(strings.TrimPrefix(mes.Content, "quiki"), " ,:")
			page := wikifier.NewPageSource(line)

			var reply string

			if err := page.Parse(); err != nil {
				reply = err.Error()
			} else {
				reply = string(page.HTML())
			}

			for _, line := range strings.Split(reply, "\n") {
				irc.Send("PRIVMSG " + mes.To + " :" + line)
			}

			return false
		},
	})

	mybot.Run() // Blocks until exit
}
