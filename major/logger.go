package major

import (
	"fmt"
	"log"
	"os"
	"time"
)

func InitLogger(name string) {
	file := "./log/"+ name + "/logger" + ".txt"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
	log.SetPrefix("[LogTool]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
	//log.SetFlags(log.LstdFlags | log.LUTC)
}

func Println(format string) {
	log.Printf(format)
	fmt.Println(time.Now().Format("2006-01-02T 15:04:05") + ": " + format)
}
