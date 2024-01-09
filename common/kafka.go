package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var conf kafka.ConfigMap

func StartKafka(conf1 kafka.ConfigMap)(*kafka.Producer, error){
    conf = conf1
    // create new producer
    producerPtr, err := kafka.NewProducer(&conf)
    if err != nil {
        fmt.Printf("Cannot create the producer: %s", err)
        os.Exit(1)
    }

    // Go routine to handle message delivery reports & other event types
    go handleMsgDelivery(producerPtr)

    return producerPtr, err
}

func handleMsgDelivery(producerPtr *kafka.Producer){
     // for e := range producerPtr.Events(){
     //        switch ev := e.(type){
     //            case *kafka.Message:
     //                if ev.TopicPartition.Error != nil {
     //                fmt.Printf("Failed to deliver message: %v \n", ev.TopicPartition)
     //             }else{
     //                fmt.Printf("Produced event to topic %s: key = %-10s value = %s \n", 
     //                    *ev.TopicPartition.Topic,
     //                    string(ev.Key),
     //                    string(ev.Value))
     //            }
     //        }
     //    }
     //

    canContinue := true
 
    for canContinue{
        select {
            case v, ok := <-producerPtr.Events() :
                if ok {
                    kafkaMsg := v.(*kafka.Message)
                    key := string(kafkaMsg.Key)
                    val := string(kafkaMsg.Value)
                    topic := *kafkaMsg.TopicPartition.Topic
                    partition := kafkaMsg.TopicPartition.Partition
                    offset := kafkaMsg.TopicPartition.Offset

                log.Println(
                    "\n\n*** RECEIVED DELIVERY REPORT ***",
                    "\nTOPIC: ", topic,
                    "\nPARTITION-OFFSET: ", partition, "-", offset,
                    "\nKEY", key ,
                    "\nVALUE", val,
                    "\n\n",
                    )
                }else {
                    log.Println("~~~ Event channel closed...")
                    canContinue = false
                }
            default:
        }
    }

}

func SendDataKafka(producerPtr *kafka.Producer, topic string, key string, value any) error{

    v, _ := json.Marshal(value)

    // create kafka message
    msg := &kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic: &topic,
            Partition: kafka.PartitionAny,
        },
        Key: []byte(key),
        Value: v,
    }

    // produce
    err := producerPtr.Produce(msg, nil)
    if err != nil {
        log.Panic("Unable to enqueue msg:", msg, err)
    }
    return err
}

func CloseKafka(producerPtr *kafka.Producer){
    log.Println("======= Closing Kafka ... ========")

    // wait for all messages to be delivered
    n := producerPtr.Flush(15 * 1000)
    log.Println("Pending events still un-flushed:", n)
    producerPtr.Close()
}
