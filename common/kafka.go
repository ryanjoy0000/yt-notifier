package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	// "github.com/confluentinc/confluent-kafka-go/schemaregistry"
	// "github.com/confluentinc/confluent-kafka-go/schemaregistry/serde"
	// "github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/avro"
	// "github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/jsonschema"
)

var conf kafka.ConfigMap

func StartKafka(conf1 kafka.ConfigMap, schemaUrl string)(*kafka.Producer,error){
    conf = conf1
    // create new producer
    producerPtr, err := kafka.NewProducer(&conf)
    if err != nil {
        fmt.Printf("Cannot create the producer: %s", err)
        os.Exit(1)
    }

    // // setup schema registry client
    // schemaConf := schemaregistry.NewConfig(schemaUrl)
    // schemaClient, err := schemaregistry.NewClient(schemaConf)
    // if err != nil {
    //     fmt.Printf("Cannot create the schema registry client: %s", err)
    //     os.Exit(1)
    // }

    // // setup avro serializer
    // avroSerPtr, err := avro.NewGenericSerializer(schemaClient, serde.ValueSerde, avro.NewSerializerConfig())
    // if err != nil {
    //     fmt.Printf("Unable to create avro serializer: %s", err)
    //     os.Exit(1)
    // }
    //
    // // setup json serializer
    // jsonSerPtr, err := jsonschema.NewSerializer(schemaClient, serde.ValueSerde, jsonschema.NewSerializerConfig())
    // if err != nil {
    //     fmt.Printf("Unable to create json serializer: %s", err)
    //     os.Exit(1)
    // }

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

func SendDataKafka(producerPtr *kafka.Producer,topic string, key string, value any) error{

    v, _ := json.Marshal(value)
    log.Println("payload json marshalled: ", string(v))
    
    // a, err := avroSerPtr.Serialize(topic, value)
    // if err != nil {
    //     log.Println("unable to avro serialize payload:", err)
    //     // return err
    // }
    // log.Println("payload avro serialized: ", string(a))
    //
    // j, err := jsonSerPtr.Serialize(topic, value)
    // if err != nil {
    //     log.Println("unable to json serialize payload:", err)
    //     // return err
    // }
    // log.Println("payload json serialized: ", string(j))

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
    log.Println("~~~ Closing Kafka...")

    // wait for all messages to be delivered
    n := producerPtr.Flush(1 * 1000)
    log.Println("Any pending events still un-flushed:", n)
    producerPtr.Close()
}
