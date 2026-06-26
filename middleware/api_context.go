// middleware/ApiContext.go

package middleware

import (
	"encoding/json"
	"net/http"

	"gusseynov/GO-Quiz/config" // Конфиг для проверки ролей боссов
	"gusseynov/GO-Quiz/sso"    // Наш пакет SSO
)

// Authorize — единая middleware: извлекает IP, проверяет сессию в SSO и собирает контекст страниц
func ApiContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		page := GetOrCreatePageCtx(r.Context())
		page.IP = GetClientIP(r)

		ssoUser, err := sso.Client.CheckSession(r.Context(), page.IP)
		if err != nil {
			slog.Warn("SSO session check failed", "ip", page.IP, "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "unauthorized",
			})
			return
		}
		page.Lang = ssoUser.Lang
		page.FIO = ssoUser.FIO
		page.Post = ssoUser.Post
		page.DepName = ssoUser.DepName
		page.LoginName = ssoUser.LoginName
		page.IsAdmin = config.IsSuperAdmin(ssoUser.FIO)

		ctx := SavePageCtx(r.Context(), page)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
