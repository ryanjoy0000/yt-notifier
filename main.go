package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ryanjoy0000/yt-notifier/common"
	"github.com/ryanjoy0000/yt-notifier/yt"
)

func main(){
    fmt.Println("App starting ...")

    conf, err := common.ReadConfig(".conf")
    handleErr(err, "Unable to read / locate config file", true)

    kafkaProperties, err := common.ReadConfig("kafka.properties") 
    handleErr(err, "Unable to read / locate kafka config file", true)
    
    schemaUrl :="kafka.schema.properties" 
    producerPtr, err := common.StartKafka(kafkaProperties, schemaUrl)
    handleErr(err, "Unable to start kafka", true)

    ytSvc := yt.NewYTDataService(&conf, producerPtr)
    _, err = ytSvc.GetYTData(context.Background())
    handleErr(err, "Unable to fetch youtube data", false)


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
