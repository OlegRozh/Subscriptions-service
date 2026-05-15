package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/OlegRozh/subscriptions-service/internal/storage"
	"github.com/OlegRozh/subscriptions-service/models"
)

type Handler struct {
	repo   storage.SubscriptionRepository
	logger *slog.Logger
}

func NewHandler(repo storage.SubscriptionRepository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Create godoc
// @Summary Создать новую подписку
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body models.Subscription true "Данные подписки"
// @Success 201 {object} map[string]int64
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var sub models.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	if sub.ServiceName == "" || sub.Price <= 0 || sub.UserId == "" || sub.StartMonth.IsZero() {
		h.logger.Warn("invalid request body")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	id, err := h.repo.Create(r.Context(), &sub)
	if err != nil {
		h.logger.Error("failed to create subscription", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create subscription"})
		return
	}
	h.logger.Info("subscription created", "id", id, "service_name", sub.ServiceName)
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// Get godoc
// @Summary Получить подписку по ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "ID подписки"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	idPath := r.PathValue("id")
	id, err := strconv.ParseInt(idPath, 10, 64)
	if err != nil {
		h.logger.Warn("invalid subscription id", "id", idPath, "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid subscription id"})
		return
	}
	req, err := h.repo.Get(r.Context(), id)
	if err != nil {
		h.logger.Warn("failed to get subscription", "id", id, "error", err)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Subscription not found"})
		return
	}
	writeJSON(w, http.StatusOK, req)
}

// GetSum godoc
// @Summary Получить сумму подписок с фильтрацией
// @Tags subscriptions
// @Produce json
// @Param user_id query string true "ID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param start query string false "Начало периода (YYYY-MM-DD)"
// @Param end query string false "Конец периода (YYYY-MM-DD)"
// @Success 200 {object} map[string]int
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/sum [get]
func (h *Handler) GetSum(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.logger.Warn("missing user_id parameter")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_id is required"})
		return
	}
	serviceName := r.URL.Query().Get("service_name")
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	sum, err := h.repo.GetSum(r.Context(), userID, serviceName, start, end)
	if err != nil {
		h.logger.Error("failed to get subscription sum", "user_id", userID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"total_sum": sum})
}

// GetList godoc
// @Summary Получить список подписок
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя"
// @Param service_name query string false "Название сервиса"
// @Success 200 {object} map[string][]models.Subscription
// @Failure 500 {object} map[string]string
// @Router /subscriptions [get]
func (h *Handler) GetList(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	subs, err := h.repo.GetList(r.Context(), userID, serviceName)
	if err != nil {
		h.logger.Error("failed to list subscriptions", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"subscriptions": subs})
}

// Update godoc
// @Summary Обновить подписку
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "ID подписки"
// @Param request body models.Subscription true "Данные для обновления"
// @Success 204 {object} nil
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	idPath := r.PathValue("id")
	id, err := strconv.ParseInt(idPath, 10, 64)
	if err != nil {
		h.logger.Warn("invalid subscription id", "id", idPath, "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid subscription id"})
		return
	}
	var sub models.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		h.logger.Warn("failed to decode request body", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}
	sub.Id = id
	if sub.ServiceName == "" || sub.Price <= 0 || sub.UserId == "" || sub.StartMonth.IsZero() {
		h.logger.Warn("validation failed", "service_name", sub.ServiceName, "price", sub.Price, "user_id", sub.UserId, "start_month", sub.StartMonth)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "service_name, price, user_id, start_month are required"})
		return
	}
	if err := h.repo.Update(r.Context(), &sub); err != nil {
		h.logger.Error("failed to update subscription", "id", id, "error", err)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Subscription not found"})
		return
	}
	h.logger.Info("subscription updated", "id", id)
	writeJSON(w, http.StatusNoContent, nil)
}

// Delete godoc
// @Summary Удалить подписку
// @Tags subscriptions
// @Param id path int true "ID подписки"
// @Success 204 {object} nil
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("invalid subscription id", "id", idStr, "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid subscription id"})
		return
	}
	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to delete subscription", "id", id, "error", err)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Subscription not found"})
		return
	}
	h.logger.Info("subscription deleted", "id", id)
	writeJSON(w, http.StatusNoContent, nil)
}
