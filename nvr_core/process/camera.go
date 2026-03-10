package process


type Camera struct {
    ID         int    `json:"id"`
    Url        string `json:"url"`
    WorkerID   int    `json:"worker_id"`
    Status     string `json:"status"`
    ChannelID  int    `json:"channel"`
}

func NewCamera(camID int, url string) *Camera {
    cam := &Camera{
        ID: camID,
        Url: url,
        WorkerID: -1,
        Status: "",
        ChannelID: -1,
    }
    return cam
}

