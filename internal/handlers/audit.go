package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// 1. Definimos nuestro contador global
var auditCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "api_audit_events_total",
		Help: "Cantidad total de eventos de auditoría registrados exitosamente",
	},
)

// 2. La función init() se ejecuta sola al arrancar el programa
func init() {
	// Registramos la métrica en el registro global de Prometheus
	prometheus.MustRegister(auditCounter)
}

// AuditEvent define la estructura de datos esperada.
// Las etiquetas entre comillas invertidas (`json:"..."`) le dicen al decodificador
// qué campo del JSON corresponde a qué propiedad del Struct.
type AuditEvent struct {
	Action    string    `json:"action"`
	User      string    `json:"user"`
	Timestamp time.Time `json:"timestamp"`
}

// CreateAuditLog procesa un evento de auditoría entrante
func CreateAuditLog(w http.ResponseWriter, r *http.Request) {
	var event AuditEvent

	// 1. Decodificar el JSON entrante
	// Pasamos el puntero (&event) para que la función Decode pueda modificar nuestra variable
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		// http.Error es una función de conveniencia que escribe el header y el mensaje en una sola línea
		http.Error(w, "JSON inválido o malformado", http.StatusBadRequest)
		return
	}

	// 2. Simulación de guardado (En el Proyecto 3 o 4 lo conectaremos a una BD real)
	fmt.Printf("[AUDIT LOG] Acción: %s | Usuario: %s | Hora: %s\n", event.Action, event.User, event.Timestamp.Format(time.RFC3339))

	// 3. Preparar la respuesta HTTP
	// Siempre debemos avisar al cliente qué tipo de contenido le estamos devolviendo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // HTTP 201 Created

	// Codificamos un mapa y evaluamos si hubo un error al enviarlo al cliente
	errEncode := json.NewEncoder(w).Encode(map[string]string{
		"status": "Evento de auditoría registrado exitosamente",
		"action": event.Action,
	})

	if errEncode != nil {
		fmt.Printf("[ERROR] No se pudo enviar la respuesta JSON al cliente: %v\n", errEncode)
	}

	// 4. Incrementar el contador de eventos de auditoría
	auditCounter.Inc()
}
