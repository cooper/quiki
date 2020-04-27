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
		Condition: func(bot *hbot.Bot, mes *hbot.Message) bool {
			return mes.Command == "PRIVMSG" && strings.HasPrefix(mes.Content, "quiki")
		},
		Action: func(irc *hbot.Bot, mes *hbot.Message) bool {
			line := strings.TrimLeft(strings.TrimPrefix(mes.Content, "quiki"), " ,:")

			// markdown if starts with "md: "
			md := false
			withoutMd := strings.TrimPrefix(line, "md: ")
			if withoutMd != line {
				line = withoutMd
				md = true
			}

			// create page
			page := wikifier.NewPageSource(strings.Replace(line, "_NL_", "\n", -1))
			if md {
				page.Markdown = true
			}
			var reply string

			// parse/generate html
			if err := page.Parse(); err != nil {
				reply = err.Error()
			} else {
				reply = string(page.HTML())
			}

			// warnings
			for _, warning := range page.Warnings {
				reply += "\nWarning " + warning.Pos.String() + " " + warning.Message
			}

			// css
			if css := page.CSS(); css != "" {
				reply += "\n\nCSS:\n" + css
			}

			// keywords
			if kw := page.Keywords(); len(kw) != 0 {
				reply += "\n\nKeywords: " + strings.Join(kw, ", ")
			}

			// reply
			for _, line := range strings.Split(reply, "\n") {
				if line == "" {
					line = " "
				}
				irc.Send("PRIVMSG " + mes.To + " :" + line)
			}

			return false
		},
	})

	mybot.Run()
}
