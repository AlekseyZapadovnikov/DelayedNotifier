package web

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/valid"
	"github.com/go-chi/chi/v5"
)

type postCreateNotifyRequest struct {
	msg      string    `json:"message"`
	date     time.Time `json:"dateTime"`
	SendChan string    `json:"sendChan"`
	from     string    `json:"from"`
	To       []string  `json:"to"`
}

func (s *Server) createNotify(w http.ResponseWriter, r *http.Request) {
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("culdn`t read data drom request body",
			"err", err,
			"method", r.Method,
			"path", r.URL.Path)
		http.Error(w, "Bad Request: failed to read request body", http.StatusBadRequest)
		return
	}

	var reqStruct postCreateNotifyRequest
	if err := json.Unmarshal(bodyData, reqStruct); err != nil {
		slog.Error("culdn`t unmurshal data drom request body", "err", err)
		http.Error(w, "Bad Request: failed to understand your request", http.StatusBadRequest)
		return
	}

	if reqStruct.msg == "" {
		slog.Error("empty msg field in createNotify",
			"method", r.Method,
			"path", r.URL.Path)
		http.Error(w, "empty msg field in createNotify", http.StatusBadRequest)
		return
	}

	if reqStruct.date.IsZero() || reqStruct.date.Before(time.Now()) {
		slog.Error("invalid date field in createNotify",
			"method", r.Method,
			"path", r.URL.Path,
			"date", reqStruct.date,
		)
		http.Error(w, "invalid date field in createNotify, date should be in future", http.StatusBadRequest)
		return
	}

	if !valid.ValidateEmailOrTg(reqStruct.from) {
		slog.Error("invalid from field",
			"method", r.Method,
			"path", r.URL.Path,
			"fromField", reqStruct.from,
		)
		http.Error(w, "invalid from field in createNotify, it should be valid email or tg username", http.StatusBadRequest)
		return
	}

	// вот тут, нормально ли, что я захардкодил?
	if reqStruct.SendChan != "tg" && reqStruct.SendChan != "mail" {
		slog.Error("invalid SendChan field in createNotify",
			"method", r.Method,
			"path", r.URL.Path,
			"SendChan", reqStruct.SendChan)
		http.Error(w, "invalid SendChan field in createNotify, it may be only one of (tg, mail)", http.StatusBadRequest)
		return
	}

	record := models.Record{
		Data:     []byte(reqStruct.msg),
		SendTime: reqStruct.date,
		RecStat:  models.RecordStatusWaiting,
		SendChan: reqStruct.SendChan,
		To:       reqStruct.To,
	}

	err = s.service.CreateNotify(context.Background(), record)
	if err != nil {
		slog.Error("couldn`t create Notify", "error", err.Error())
		http.Error(w, "couldn`t create Notify, try again", http.StatusInternalServerError)
		return
	}
}

func (s *Server) getNotifyStatByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	if strings.TrimSpace(idStr) == "" {
		slog.Error("empty id in getNotifyStatByID",
			"method", r.Method,
			"path", r.URL.Path,
			"id", idStr)
		http.Error(w, "got request with empty id", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Error("invalid id format in getNotifyStatByID",
			"method", r.Method,
			"path", r.URL.Path,
			"id", idStr)
		http.Error(w, "invalid id format, id should be a number", http.StatusBadRequest)
		return
	}

	err = s.service.GetNotifyStatByID(context.Background(), id)
	if err != nil {
		slog.Error("couldn`t get notify stat by ID",
			"error", err.Error(),
			"id", id)
		http.Error(w, "couldn`t get notify stat by ID", http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteNotifyByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	if strings.TrimSpace(idStr) == "" {
		slog.Error("empty id in deleteNotifyByID",
			"method", r.Method,
			"path", r.URL.Path,
			"id", idStr)
		http.Error(w, "got request with empty id", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Error("invalid id format in deleteNotifyByID",
			"method", r.Method,
			"path", r.URL.Path,
			"id", idStr)
		http.Error(w, "invalid id format, id should be a number", http.StatusBadRequest)
		return
	}

	err = s.service.DeleteNotifyByID(context.Background(), id)
	if err != nil {
		slog.Error("couldn`t delete notify by ID",
			"error", err.Error(),
			"id", id)
		http.Error(w, "couldn`t delete notify by ID", http.StatusInternalServerError)
		return
	}
}
