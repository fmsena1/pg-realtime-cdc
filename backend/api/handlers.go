package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/fmsena1/pg-realtime-cdc/backend/db"
	"github.com/fmsena1/pg-realtime-cdc/backend/models"
)

func GetMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := db.GetAllMessages()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch messages")
		return
	}
	respondJSON(w, http.StatusOK, messages)
}

func CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "Content is required")
		return
	}

	message, err := db.CreateMessage(req.Content)
	if err != nil {
		log.Printf("Error creating message: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create message")
		return
	}

	respondJSON(w, http.StatusCreated, message)
}

func UpdateMessage(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	var req models.UpdateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "Content is required")
		return
	}

	message, err := db.UpdateMessage(id, req.Content)
	if err != nil {
		log.Printf("Error updating message: %v", err)
		if err.Error() == "message not found" {
			respondError(w, http.StatusNotFound, "Message not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update message")
		return
	}

	respondJSON(w, http.StatusOK, message)
}

func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	if err := db.DeleteMessage(id); err != nil {
		log.Printf("Error deleting message: %v", err)
		if err.Error() == "message not found" {
			respondError(w, http.StatusNotFound, "Message not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete message")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Message deleted successfully"})
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

