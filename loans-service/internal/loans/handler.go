package loans
import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
service *Service
}

type ErrorResponse struct {
Message string `json:"message"`
}

func NewHandler(service *Service) *Handler {
return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
router.Post("/loans", h.registerLoan)
router.Post("/loans/{id}/return", h.returnLoan)
router.Get("/loans/users/{userId}/active", h.findActiveByUser)
router.Get("/loans/history", h.findHistory)
}

func (h *Handler) registerLoan(w http.ResponseWriter, r *http.Request) {
var request RegisterLoanRequest

if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
writeError(w, http.StatusBadRequest, "Invalid request body")
return
}

request.UserID = strings.TrimSpace(request.UserID)
request.BookID = strings.TrimSpace(request.BookID)

if request.UserID == "" || request.BookID == "" {
writeError(w, http.StatusBadRequest, "userId and bookId are required")
return
}

ctx, cancel := withTimeout(r)
defer cancel()

loan, err := h.service.RegisterLoan(ctx, request)
if err != nil {
if errors.Is(err, ErrActiveLoanExists) {
writeError(w, http.StatusConflict, "User already has an active loan for this book")
return
}

writeError(w, http.StatusBadGateway, err.Error())
return
}

writeJSON(w, http.StatusCreated, loan)
}

func (h *Handler) returnLoan(w http.ResponseWriter, r *http.Request) {
loanID := chi.URLParam(r, "id")

if strings.TrimSpace(loanID) == "" {
writeError(w, http.StatusBadRequest, "loan id is required")
return
}

ctx, cancel := withTimeout(r)
defer cancel()

loan, err := h.service.ReturnLoan(ctx, loanID)
if err != nil {
if errors.Is(err, ErrLoanNotFound) {
writeError(w, http.StatusNotFound, "Active loan not found")
return
}

writeError(w, http.StatusBadGateway, err.Error())
return
}

writeJSON(w, http.StatusOK, loan)
}

func (h *Handler) findActiveByUser(w http.ResponseWriter, r *http.Request) {
userID := chi.URLParam(r, "userId")

if strings.TrimSpace(userID) == "" {
writeError(w, http.StatusBadRequest, "user id is required")
return
}

ctx, cancel := withTimeout(r)
defer cancel()

loans, err := h.service.FindActiveByUser(ctx, userID)
if err != nil {
writeError(w, http.StatusInternalServerError, err.Error())
return
}

writeJSON(w, http.StatusOK, loans)
}

func (h *Handler) findHistory(w http.ResponseWriter, r *http.Request) {
ctx, cancel := withTimeout(r)
defer cancel()

loans, err := h.service.FindHistory(ctx)
if err != nil {
writeError(w, http.StatusInternalServerError, err.Error())
return
}

writeJSON(w, http.StatusOK, loans)
}

func withTimeout(r *http.Request) (context.Context, context.CancelFunc) {
return context.WithTimeout(r.Context(), 5*time.Second)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(statusCode)
_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
writeJSON(w, statusCode, ErrorResponse{
Message: message,
})
}
