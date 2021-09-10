package completion

import (
	"io"
	"log"
	"strings"

	"github.com/chzyer/readline"
	"github.com/gookit/color"
)

func Start() {

	red := color.FgRed.Render
	blue := color.FgBlue.Render
	green := color.FgGreen.Render
	ascii := `AWS S3 C2 Framework`
	print(ascii + "\n")

	for {

		//Autocompletion configuration
		var completer = readline.NewPrefixCompleter(

			//Grab Credentials
			readline.PcItem("token",
				readline.PcItem("profile",
					readline.PcItemDynamic(listProfiles(),
						readline.PcItem("us-east-1"),
						readline.PcItem("us-east-2"),
						readline.PcItem("us-west-1"),
						readline.PcItem("us-west-2"),
					),
				),
				readline.PcItem("AssumeRaw",
					readline.PcItem("us-east-1"),
					readline.PcItem("us-east-2"),
					readline.PcItem("us-west-1"),
					readline.PcItem("us-west-2"),
				),
				readline.PcItem("GetSessionToken",
					readline.PcItem("us-east-1"),
					readline.PcItem("us-east-2"),
					readline.PcItem("us-west-1"),
					readline.PcItem("us-west-2"),
				),
			),
		)

		//readlines configuration
		l, err := readline.NewEx(&readline.Config{
			Prompt:          "\033[31m»\033[0m ",
			HistoryFile:     "/tmp/readline.tmp",
			AutoComplete:    completer,
			InterruptPrompt: "^C",
			EOFPrompt:       "exit",

			HistorySearchFold:   true,
			FuncFilterInputRune: filterInput,
		})
		if err != nil {
			panic(err)
		}
		defer l.Close()

		log.SetOutput(l.Stderr())
		if target == "" || connected == false {
			l.SetPrompt(red("Not Connected") + " <" + blue("") + "> ")
		} else {
			l.SetPrompt(green("Connected") + " <" + blue(target+"/"+region+"/"+bucket) + "> ")
		}
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		//Trimwhitespace and send to commands switch
		line = strings.TrimSpace(line)
		Commands(line)
	}
}

//Filter input from readline CtrlZ
func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}
