package handlerutil

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func GetRequestParamUUID(r *http.Request) (uuid.UUID, error) {
	idParam := r.URL.Query().Get("id")
	return uuid.Parse(idParam)
}

func GetRequestParamBool(r *http.Request, param string) (bool, error) {
	value := r.URL.Query().Get(param)
	if value == "" {
		return false, nil // Return false if the parameter is not provided
	}
	return strconv.ParseBool(value)
}

func ParseJsonBody(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	return decoder.Decode(v)
}
