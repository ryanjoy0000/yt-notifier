package yt

type YTData struct{
    VideoId string `json:"video_id"`
    VideoTitle string `json:"video_title"`
    VideoLikeCount int `json:"video_like_count"`
    VideoCommentCount int `json:"video_comment_count"`
    ClientTelegramId string `json:"client_telegram_id"`
}
