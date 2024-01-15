package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ryanjoy0000/yt-notifier/common"
	"github.com/ryanjoy0000/yt-notifier/telegram"
	"github.com/ryanjoy0000/yt-notifier/yt"
)

func main(){
    fmt.Println("App starting ...")

    conf, err := common.ReadConfig(".conf")
    handleErr(err, "Unable to read / locate config file", true)

    kafkaProperties, err := common.ReadConfig("kafka.properties") 
    handleErr(err, "Unable to read / locate kafka config file", true)
    
    schemaUrl :="kafka.schema.properties" 

    telegramSvc, err := telegram.NewTelegramService(conf["TELEGRAM_API_KEY"].(string)) 
    handleErr(err, "unable to create telegram service", true)
    log.Println("telegram service created...")

    ctx := context.Background()
    telegramSvc.InitTelegramBot(false, ctx)
    handleErr(err, "unable to create telegram bot", true)

    playListID := <-telegramSvc.PlaylistIDChan
    clientTelegramId := <- telegramSvc.ClientTelegramId
    log.Println("playlist id received via channel", playListID)


        // TODO: Replace this functionality with a scheduler
    counter := 2
    waitTime := 30 * time.Second
    for i := 0; i < counter; i++ {        
        producerPtr, err := common.StartKafka(kafkaProperties, schemaUrl)
        handleErr(err, "Unable to start kafka", true)
        
        ytSvc := yt.NewYTDataService(&conf, producerPtr, playListID, clientTelegramId)
        _, err = ytSvc.GetYTData(ctx)
        handleErr(err, "Unable to fetch youtube data", false)

        if i < counter - 1 {
            log.Println("MOCK SCHEDULER: Sleeping for ", waitTime, " secs...")
            time.Sleep(waitTime)
        }
    }

    log.Println("App exiting...")
}

func handleErr(err error, msg string, shouldExit bool){
    if err != nil {
        if shouldExit {
            log.Fatalln(msg)
        }else {
            log.Println(msg)
        }
    }
}
