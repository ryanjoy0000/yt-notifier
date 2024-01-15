package yt

import (
	"context"
	// "encoding/json"
	// "fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	// "github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/avro"
	// "github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/jsonschema"
	"github.com/ryanjoy0000/yt-notifier/common"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

var err error

type YTDataService struct{
    producerPtr *kafka.Producer
    confPtr *kafka.ConfigMap
    playlistID string
    clientTelegramId string
}

func(y *YTDataService) GetYTData(ctx context.Context) ([]*YTData, error){
    nextPageToken := ""
    hasPageToken := true 
    result := []*YTData{}
    resultSection := []*YTData{}

    var pId string
    if y.playlistID != ""{
        pId = y.playlistID
    }else{
        pId = (*y.confPtr)["YT_PLAYLIST_ID"].(string)
    }

    for hasPageToken{
        nextPageToken, resultSection, err =  y.fetchYTPlaylist(context.Background(), (*y.confPtr)["YT_API_KEY"].(string), pId, nextPageToken)
        if nextPageToken != "" {
            hasPageToken = true
            // fmt.Println("=== Fetching next section of results... ===.")
        }else{
            hasPageToken = false
            // fmt.Println("=== END of results===")
        }
        result = append(result, resultSection...)
    }

    // close kafka
    common.CloseKafka(y.producerPtr)

    return result, err
}

func NewYTDataService(conf1 *kafka.ConfigMap, producerPtr *kafka.Producer, playlistID string, clientTelegramId string) *YTDataService{
    return &YTDataService{
        producerPtr: producerPtr,
        confPtr: conf1,
        playlistID: playlistID,
        clientTelegramId: clientTelegramId,
    }
}

func (y *YTDataService)fetchYTPlaylist(ctx context.Context, apiKey string, playlistID string, pageToken string) (string, []*YTData, error){
    var nextPageToken string
    
    // connect to Youtube API
    ytSvc, err := yt.NewService(ctx, option.WithAPIKey(apiKey))
    if err!= nil {
        log.Panicln("Unable to connect to Youtube API : ", err)
        return "", nil, err
    }
    
    // sections required from Youtube data
    sectionList := []string{"contentDetails"} //"snippet", "id" 
   
    log.Println("fetching playlist...")
    // request youtube api
    plRespPtr, err := ytSvc.PlaylistItems.List(sectionList).PlaylistId(playlistID).PageToken(pageToken).Do()
    if err!= nil {
        log.Panicln("Unable to fetch Youtube playlist details : ", err)
        return "", nil, err
    }

    // logging
    // bSlice, err := json.MarshalIndent(*plRespPtr, "", "\t")
    // log.Println("Playlist Response: ", string(bSlice))
    
    // get next page
    if plRespPtr.NextPageToken != ""{
        nextPageToken = plRespPtr.NextPageToken
    }

    // Fetch Video Info
    log.Println("fetching videos...")
    result, err := y.fetchVideos(plRespPtr, ytSvc)
    if err != nil {
        log.Println("Unable to fetch video")
        return "", nil, err
    }

    // logging
    // bSlice, err = json.MarshalIndent(result, "", "\t")
    // fmt.Println("video info: ", string(bSlice))

    return nextPageToken, result, err
}

func (y *YTDataService)fetchVideos(plRespPtr *yt.PlaylistItemListResponse, ytSvc *yt.Service)([]*YTData, error) {
    result := []*YTData{}
    var err error
    if plRespPtr.Items != nil {
            
        // extract video ids, looping through playlist
        for _, plItemPtr:= range plRespPtr.Items {
            // log.Println("Index: ", key)
            videoId :=plItemPtr.ContentDetails.VideoId
            sectionList := []string{"statistics", "snippet"}  
            vidListResp, err := ytSvc.Videos.List(sectionList).Id(videoId).Do()
            if err != nil {
                log.Println("Unable to fetch video details: ", err)
               
            }

            if vidListResp.Items != nil {
                for _, v := range vidListResp.Items {
                    videoPtr := &YTData{
                        VideoTitle: v.Snippet.Title,
                        VideoLikeCount: int(v.Statistics.LikeCount),
                        VideoCommentCount: int(v.Statistics.CommentCount),
                        ClientTelegramId: y.clientTelegramId,
                    }
                    result = append(result, videoPtr)


                    // SEND DATA USING KAFKA
                    err := common.SendDataKafka(y.producerPtr,(*y.confPtr)["KAFKA_TOPIC"].(string), v.Id, *videoPtr)
                    if err != nil {
                        log.Println("unable to send data using kafka", err)
                        return nil, err
                    }
                }    
            }
        }
    }
    return result, err
}
