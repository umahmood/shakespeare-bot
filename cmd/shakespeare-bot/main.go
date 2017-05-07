package main

import (
	"fmt"
	"log"
	"os"
	"time"

	shakespearebot "github.com/umahmood/shakespeare-bot"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: shakespeare-bot <TOKEN>")
		return
	}
	token := os.Args[1]
	bot, err := shakespearebot.NewBot(token)
	if err != nil {
		log.Println(err)
		return
	}
	err = bot.Start()
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println("stopping")
		bot.Stop()
	}()

	err = bot.ListenAndRespond() // blocking call
	if err != nil {
		log.Println(err)
	}

	fmt.Println("starting bot again...")
	err = bot.Start()
	if err != nil {
		log.Println(err)
		return
	}

	err = bot.ListenAndRespond() // blocking call
	if err != nil {
		log.Println(err)
	}
}
