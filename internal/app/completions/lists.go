package completion

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"regexp"
)

func listProfiles() func(string) []string {
	return func(line string) []string {
		rule := `\[(.*)\]`
		var profiles []string

		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}

		dat, err := ioutil.ReadFile(usr.HomeDir + "/.aws/credentials")
		if err != nil {
			fmt.Println(err)
		}

		r, _ := regexp.Compile(rule)
		if r.MatchString(string(dat)) {
			matches := r.FindAllStringSubmatch(string(dat), -1)
			for _, v := range matches {
				profiles = append(profiles, v[1])
			}
		}
		return profiles
	}
}
