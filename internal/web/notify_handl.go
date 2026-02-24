package web

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/valid"
)

type PostCreateNotifyRequest struct {
	Msg      string    `json:"message" validate:"required"`
	Date     time.Time `json:"dateTime" validate:"required,gt=now"`
	SendChan string    `json:"sendChan" validate:"required,oneof=tg mail"`
	From     string    `json:"from"`
	To       []string  `json:"to" validate:"required,min=1,to_field"`
}

type PostCreateNotifyResponse struct {
	ID       int64               `json:"id"`
	Message  string              `json:"message"`
	DateTime time.Time           `json:"dateTime"`
	Status   models.RecordStatus `json:"status"`
	SendChan string              `json:"sendChan"`
	To       []string            `json:"to"`
}

func (s *Server) createNotify(w http.ResponseWriter, r *http.Request) {
	var reqStruct PostCreateNotifyRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&reqStruct); err != nil {
		slog.Error("couldn`t unmarshal data from request body",
			"method", r.Method,
			"url", r.URL.String(),
			"err", err,
		)
		http.Error(w, "Bad Request: failed to understand your request", http.StatusBadRequest)
		return
	}

	if err := valid.Validate.Struct(reqStruct); err != nil {
		slog.Error("validation failed",
			"method", r.Method,
			"url", r.URL,
			"err", err,
		)
		msgForUser := valid.RecordValidationDescription(err) // получаем читаемое сообщение об ошибке для пользователя
		http.Error(w, fmt.Sprintf("Validation error: %s", msgForUser), http.StatusBadRequest)
		return
	}

	record := models.Record{
		Data:     []byte(reqStruct.Msg),
		SendTime: reqStruct.Date,
		RecStat:  models.RecordStatusWaiting,
		SendChan: reqStruct.SendChan,
		From:     strings.TrimSpace(reqStruct.From),
		To:       reqStruct.To,
	}
	if record.SendChan == models.SendChanMail && strings.TrimSpace(s.defaultFrom) != "" {
		record.From = s.defaultFrom
	}

	err := s.service.CreateNotify(r.Context(), &record)
	if err != nil {
		slog.Error("couldn`t create Notify",
			"method", r.Method,
			"url", r.URL,
			"error", err,
		)
		http.Error(w, "couldn`t create Notify, try again", http.StatusInternalServerError)
		return
	}

	resp := PostCreateNotifyResponse{
		ID:       record.Id,
		Message:  reqStruct.Msg,
		DateTime: record.SendTime,
		Status:   record.RecStat,
		SendChan: record.SendChan,
		To:       append([]string(nil), record.To...),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) getNotifyStatByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Error("invalid id format",
			"method", r.Method,
			"url", r.URL,
			"input value", idStr,
			"error", err,
		)
		http.Error(w, "invalid id format, id should be a number", http.StatusBadRequest)
		return
	}

	status, err := s.service.GetNotifyStatByID(r.Context(), id)
	if err != nil {
		slog.Error("couldn`t get notify stat by ID",
			"method", r.Method,
			"url", r.URL,
			"id", id,
			"error", err,
		)
		http.Error(w, "couldn`t get notify stat by ID", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(status))
}

func (s *Server) deleteNotifyByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Error("invalid id format",
			"method", r.Method,
			"url", r.URL,
			"input value", idStr,
			"error", err,
		)
		http.Error(w, "invalid id format, id should be a number", http.StatusBadRequest)
		return
	}

	err = s.service.DeleteNotifyByID(r.Context(), id)
	if err != nil {
		slog.Error("couldn`t delete notify by ID",
			"method", r.Method,
			"url", r.URL,
			"id", id,
			"error", err,
		)
		http.Error(w, "couldn`t delete notify by ID", http.StatusInternalServerError)
		return
	}
}
