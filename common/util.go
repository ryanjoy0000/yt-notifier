package common

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func ReadConfig(configFilePath string) (kafka.ConfigMap, error){

    // map of key & config
    m := make(kafka.ConfigMap)
    
    // open config file 
    file, err := os.Open(configFilePath)
    if err!= nil {
        log.Panicln("Cannot open or locate config file: ", err)
        
    }
    defer file.Close()

    // read from config file
    scanner := bufio.NewScanner(file)
    for scanner.Scan(){
        line := strings.TrimSpace(scanner.Text())
        // skip comments & empty line
        if !strings.HasPrefix(line, "#") && len(line)!=0 {
        before, after, found := strings.Cut(line, "=")
        if found{
                parameter := strings.TrimSpace(before)
                value := strings.TrimSpace(after)
                m[parameter] = value;
            }
        }
    }
    err = scanner.Err()
    if err != nil {
        log.Panicln("Cannot read config file:", err)
    }
    log.Println("Config file acquired: ", configFilePath)
    return m, err
}

func LoadConfig() (kafka.ConfigMap, error){
    // load config file
    if len(os.Args) != 2 {
        fmt.Fprintf(os.Stderr, "Usage: %s <config-file-path>\n", os.Args[0])
        os.Exit(1)
    }
    configFilePath := os.Args[1]
    return ReadConfig(configFilePath)
}
