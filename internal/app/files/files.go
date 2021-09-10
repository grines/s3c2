package files

import (
	"log"
	"os"
)

func CreateCommand(cmd string) string {
	f, e := os.CreateTemp("", "*.cmd")
	if e != nil {
		panic(e)
	}
	defer f.Close()

	f, err := os.OpenFile(f.Name(),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(cmd + "\n"); err != nil {
		log.Println(err)
	}

	return f.Name()
}
