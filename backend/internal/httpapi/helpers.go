package httpapi

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"error": message, "code": code})
}

func writeOK(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": data})
}

func writeList(w http.ResponseWriter, data interface{}, total int) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": data,
		"meta": map[string]int{"total": total},
	})
}
