package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/domain"
)

func ParseJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}
	return json.NewDecoder(r.Body).Decode(v)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteMessage(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"message": msg})
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}

func ParseUUID(uuid string) (domain.UUID, error) {
	if len(uuid) != 36 {
		return "", fmt.Errorf("invalid UUID length %v", len(uuid))
	}
	if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
		return "", fmt.Errorf("invalid UUID format")
	}
	if uuid[14] != '4' {
		return "", fmt.Errorf("invalid UUID version")
	}
	if uuid[19] != '8' && uuid[19] != '9' && uuid[19] != 'a' && uuid[19] != 'b' {
		return "", fmt.Errorf("invalid UUID variant")
	}
	return domain.UUID(uuid), nil
}
