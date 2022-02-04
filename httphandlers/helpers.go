package httphandlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

func DumpJSON(jsonstruct interface{}) []byte {
	marshal, _ := json.Marshal(jsonstruct)
	return marshal
}

func WriteJSON(w http.ResponseWriter, jsonstruct interface{}, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Write(DumpJSON(jsonstruct))
}

func WriteForbiddenJWT(w http.ResponseWriter) {
	w.WriteHeader(403)
	w.Header().Set("Content-Type", "application/json")
	w.Write(DumpJSON(Response{Success: false, Data: "Forbidden"}))
}

func GetAuthorizationJWT(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	return strings.Split(h, " ")[1]
}
