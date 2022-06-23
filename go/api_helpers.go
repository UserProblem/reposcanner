package swagger

import (
	"encoding/json"
	"net/http"
	"time"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	w.Write(response)
}

func currentTimestamptz() string {
	return time.Now().Format("2006-01-02T15:04:05Z07:00")
}
