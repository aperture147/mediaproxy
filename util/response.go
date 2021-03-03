package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func WriteJsonResponse(w http.ResponseWriter, status int, message string, data interface{}) {
	response := Response{
		Status:  status,
		Message: message,
		Data:    data,
	}

	jsonData, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonData)
}

func WriteOkResponse(w http.ResponseWriter, data interface{}) {
	WriteJsonResponse(w, http.StatusOK, "ok", data)
}

func WriteJsonErrorResponse(w http.ResponseWriter, status int, message string, err error) {
	WriteJsonResponse(w, status, fmt.Sprint(message), nil)
}

func WriteForbiddenResponse(w http.ResponseWriter, err error) {
	WriteJsonErrorResponse(w, http.StatusForbidden, "forbidden", err)
}

func WriteBadRequestResponse(w http.ResponseWriter, err error) {
	WriteJsonErrorResponse(w, http.StatusBadRequest, "bad request", err)
}

func WriteUnauthorizedResponse(w http.ResponseWriter, err error) {
	WriteJsonErrorResponse(w, http.StatusUnauthorized, "unauthorized", err)
}

func WriteServerErrorResponse(w http.ResponseWriter, err error) {
	WriteJsonErrorResponse(w, http.StatusInternalServerError, "internal router error", err)
}
