package httputils

import (
	"encoding/json"
	"net/http"
)

func WriteJson(w http.ResponseWriter, data interface{}) {
	out, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Write(out)
}
