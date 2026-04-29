package main

import (
	"fmt"
	"net/http"

	"github.com/YagoGomez83/proyecto2-go-audit/internal/config"
	"github.com/YagoGomez83/proyecto2-go-audit/internal/handlers"
	"github.com/YagoGomez83/proyecto2-go-audit/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	// 1. Cargamos la configuración (Si falta JWT_SECRET, el programa muere aquí)
	cfg := config.LoadConfig()
	// 1. Inicializamos el enrutador
	r := chi.NewRouter()

	// 2. Agregamos Middlewares base (Seguridad y Observabilidad)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// 3. Definimos nuestro primer endpoint: Healthcheck
	r.Get("/health", handlers.HealthCheck)

	// Esta línea expone todas las métricas de la app y de Go (memoria, CPU, garbage collector)
	r.Handle("/metrics", promhttp.Handler())

	r.With(middleware.RequireJWT(cfg.JWTSecret)).Post("/audit", handlers.CreateAuditLog)

	// Usamos el puerto de la configuración
	puerto := ":" + cfg.Port
	fmt.Printf("Iniciando servidor de auditoría en el puerto %s...\n", puerto)

	// ListenAndServe bloquea el hilo principal y mantiene el servicio vivo
	err := http.ListenAndServe(puerto, r)
	if err != nil {
		fmt.Printf("Error fatal al iniciar el servidor: %v\n", err)
	}
}
