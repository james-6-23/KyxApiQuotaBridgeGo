package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5/middleware"

    "github.com/kyx-api-quota-bridge/backend/internal/domain"
    "github.com/kyx-api-quota-bridge/backend/internal/service"
)

type AuthHandler struct {
    services *service.Services
}

func NewAuthHandler(services *service.Services) *AuthHandler {
    return &AuthHandler{services: services}
}

func (h *AuthHandler) Bind(w http.ResponseWriter, r *http.Request) {
    session := h.sessionFromContext(r.Context())
    if session == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    var payload struct {
        Username string `json:"username"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    result, err := h.services.Users.Bind(r.Context(), *session, payload.Username)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    h.writeJSON(w, result)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    sessionID := h.cookieValue(r, "session_id")
    if sessionID != "" {
        _ = h.services.Sessions.Delete(r.Context(), sessionID)
    }
    http.SetCookie(w, &http.Cookie{Name: "session_id", Path: "/", MaxAge: -1})
    h.writeJSON(w, map[string]any{"success": true})
}

func (h *AuthHandler) UserSessionMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sessionID := h.cookieValue(r, "session_id")
        if sessionID == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        session, err := h.services.Sessions.Get(r.Context(), sessionID)
        if err != nil || session == nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), sessionContextKey{}, session)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func (h *AuthHandler) AdminSessionMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sessionID := h.cookieValue(r, "admin_session")
        if sessionID == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        session, err := h.services.Sessions.Get(r.Context(), sessionID)
        if err != nil || session == nil || !session.IsAdmin {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), sessionContextKey{}, session)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

type sessionContextKey struct{}

func (h *AuthHandler) sessionFromContext(ctx context.Context) *domain.Session {
    session, _ := ctx.Value(sessionContextKey{}).(*domain.Session)
    return session
}

func (h *AuthHandler) cookieValue(r *http.Request, name string) string {
    cookie, err := r.Cookie(name)
    if err != nil {
        return ""
    }
    return cookie.Value
}

func (h *AuthHandler) writeJSON(w http.ResponseWriter, v any) {
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(v); err != nil {
        middleware.GetLogEntry(r.Context()).Errorf("write json: %v", err)
    }
}
