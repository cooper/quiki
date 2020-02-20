package main

import (
	"encoding/json"
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

	mybot.AddTrigger(hbot.Trigger{
		func(bot *hbot.Bot, mes *hbot.Message) bool {
			return mes.Command == "PRIVMSG" && strings.HasPrefix(mes.Content, ".displaypage")
		},
		func(irc *hbot.Bot, mes *hbot.Message) bool {
			var reply string
			w, err := wiki.NewWiki("../wikis/mywiki/wiki.conf", "")

			if err != nil {
				reply = err.Error()
			} else {
				j, _ := json.Marshal(w.DisplayPage("wikifier.page"))
				reply = fmt.Sprintf("%+v", string(j))
			}

			for _, line := range strings.Split(reply, "\n") {
				irc.Send("PRIVMSG " + mes.To + " :" + line)
			}

			return false
		},
	})

	mybot.AddTrigger(hbot.Trigger{
		func(bot *hbot.Bot, mes *hbot.Message) bool {
			return mes.Command == "PRIVMSG" && strings.HasPrefix(mes.Content, ".displayimage")
		},
		func(irc *hbot.Bot, mes *hbot.Message) bool {
			line := strings.TrimLeft(strings.TrimPrefix(mes.Content, ".displayimage"), " ,:")

			var reply string
			w, err := wiki.NewWiki("../wikis/mywiki/wiki.conf", "")

			if err != nil {
				reply = err.Error()
			} else {
				res := w.DisplayImage(line)
				reply = fmt.Sprintf("%+v", res)
			}

			for _, line := range strings.Split(reply, "\n") {
				irc.Send("PRIVMSG " + mes.To + " :" + line)
			}

			return false
		},
	})

	mybot.AddTrigger(hbot.Trigger{
		func(bot *hbot.Bot, mes *hbot.Message) bool {
			return mes.Command == "PRIVMSG" && strings.HasPrefix(mes.Content, ".names ")
		},
		func(irc *hbot.Bot, mes *hbot.Message) bool {
			pageName := strings.TrimPrefix(mes.Content, ".names ")
			w, _ := wiki.NewWiki("../wikis/mywiki/wiki.conf", "")
			page := w.FindPage(pageName)
			reply := fmt.Sprintf("%+v", struct {
				Name, NameNE, RelName, RelNameNE, Path, RelPath string
			}{
				Name:      page.Name(),
				NameNE:    page.NameNE(),
				RelName:   page.RelName(),
				RelNameNE: page.RelNameNE(),
				Path:      page.Path(),
				RelPath:   page.RelPath(),
			})
			for _, line := range strings.Split(reply, "\n") {
				irc.Send("PRIVMSG " + mes.To + " :" + line)
			}
			return false
		},
	})

	mybot.Run() // Blocks until exit
}
