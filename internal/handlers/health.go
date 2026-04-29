package handlers

import (
	"net/http"
)

// HealthCheck evalúa si el microservicio está respondiendo correctamente.
// Se usa principalmente para los liveness/readiness probes de Kubernetes o Docker.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	// Asignamos el resultado y el error a '_' indicando que los ignoramos conscientemente
	_, _ = w.Write([]byte("OK - Servicio Activo y Refactorizado"))
}
