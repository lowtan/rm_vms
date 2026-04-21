package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nvr_core/utils"
)

type WebConfig struct {
    APIPort int    `json:"apiPort"`
    WSUrl   string `json:"wsUrl"`
}

func GenerateConfig(config *utils.Config) WebConfig {

    return WebConfig{
        APIPort: config.Server.Port,
    }

}

// HandleGenerateConfig takes your config struct and returns an HTTP handler.
func HandleGenerateConfig(config WebConfig) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        // Serialize the Go struct safely into a JSON byte array
        jsonData, err := json.Marshal(config)
        if err != nil {
            http.Error(w, "Failed to generate configuration", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/javascript")
        w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")

        jsPayload := fmt.Sprintf("(()=>{if(window) {window.____API_WEB_CONFIG____ = %s}})();", string(jsonData))
        w.Write([]byte(jsPayload))

    }
}