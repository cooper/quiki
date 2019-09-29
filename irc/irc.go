package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cooper/quiki/wiki"
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

			// html
			if err := page.Parse(); err != nil {
				reply = err.Error()
			} else {
				reply = string(page.HTML())
			}

			// css
			if css := page.CSS(); css != "" {
				reply += "\n\nCSS:\n" + css
			}

			for _, line := range strings.Split(reply, "\n") {
				irc.Send("PRIVMSG " + mes.To + " :" + line)
			}

			return false
		},
	})

	mybot.AddTrigger(hbot.Trigger{
		func(bot *hbot.Bot, mes *hbot.Message) bool {
			return mes.Command == "PRIVMSG" && strings.HasPrefix(mes.Content, ".confvar")
		},
		func(irc *hbot.Bot, mes *hbot.Message) bool {
			line := strings.TrimLeft(strings.TrimPrefix(mes.Content, ".confvar"), " ,:")

			var reply string
			page := wikifier.NewPage("../wikis/mywiki/wiki.conf")

			// parse
			if err := page.Parse(); err != nil {
				reply = err.Error()
			} else {
				val, err := page.Get(line)
				if err != nil {
					reply = err.Error()
				} else {
					reply = fmt.Sprintf("%+v", val)
				}
			}

			for _, line := range strings.Split(reply, "\n") {
				irc.Send("PRIVMSG " + mes.To + " :" + line)
			}

			return false
		},
	})

	mybot.AddTrigger(hbot.Trigger{
		func(bot *hbot.Bot, mes *hbot.Message) bool {
			return mes.Command == "PRIVMSG" && strings.HasPrefix(mes.Content, ".unique")
		},
		func(irc *hbot.Bot, mes *hbot.Message) bool {
			// line := strings.TrimLeft(strings.TrimPrefix(mes.Content, "quiku"), " ,:")

			f, err := wikifier.UniqueFilesInDir("../standalone/", []string{"page"}, false)
			reply := fmt.Sprintf("err(%v) %v", err, f)

			for _, line := range strings.Split(reply, "\n") {
				irc.Send("PRIVMSG " + mes.To + " :" + line)
			}

			return false
		},
	})

	mybot.AddTrigger(hbot.Trigger{
		func(bot *hbot.Bot, mes *hbot.Message) bool {
			return mes.Command == "PRIVMSG" && strings.HasPrefix(mes.Content, ".confvars")
		},
		func(irc *hbot.Bot, mes *hbot.Message) bool {
			w, err := wiki.NewWiki("../wikis/mywiki/wiki.conf", "")

			var reply string
			if err != nil {
				reply = err.Error()
			} else {
				reply = fmt.Sprintf("%+v", w.Opt)
			}

			for _, line := range strings.Split(reply, "\n") {
				irc.Send("PRIVMSG " + mes.To + " :" + line)
			}

			return false
		},
	})

	mybot.AddTrigger(hbot.Trigger{
		func(bot *hbot.Bot, mes *hbot.Message) bool {
			return mes.Command == "PRIVMSG" && strings.HasPrefix(mes.Content, ".pageinfo")
		},
		func(irc *hbot.Bot, mes *hbot.Message) bool {
			var reply string
			page := wikifier.NewPage("../wikis/mywiki/pages/wikifier.page")

			// parse
			if err := page.Parse(); err != nil {
				reply = err.Error()
			} else {
				val := page.Info()
				reply = fmt.Sprintf("%+v", val)
			}

			for _, line := range strings.Split(reply, "\n") {
				irc.Send("PRIVMSG " + mes.To + " :" + line)
			}

			return false
		},
	})

	mybot.Run() // Blocks until exit
}
