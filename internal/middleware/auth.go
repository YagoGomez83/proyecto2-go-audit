package middleware

import (
	"net/http"
)

// RequireAPIKey intercepta la petición y valida un token de seguridad
func RequireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extraemos el header de seguridad de la petición entrante
		apiKey := r.Header.Get("X-API-Key")

		// 2. Validación (Hardcodeado por ahora. En producción lo leeremos de variables de entorno)
		const validKey = "super-secreto-devsecops-123"

		if apiKey != validKey {
			// Si no coincide, cortamos la ejecución aquí mismo y devolvemos 401 Unauthorized
			http.Error(w, "Acceso Denegado: API Key inválida o ausente", http.StatusUnauthorized)
			return
		}

		// 3. Si la llave es correcta, pasamos el control al siguiente handler (nuestro CreateAuditLog)
		next.ServeHTTP(w, r)
	})
}
