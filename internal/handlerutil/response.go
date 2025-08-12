package handlerutil

import (
	"encoding/json"
	"net/http"
)

func JsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
