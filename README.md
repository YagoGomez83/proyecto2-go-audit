# 🛡️ Proyecto 2: Microservicio de Auditoría Segura

> **Serie: Ruta Senior DevSecOps con Go** — Proyecto 2 de N

Un microservicio de alto rendimiento para la ingesta de logs de auditoría, construido con seguridad de grado bancario (JWT), observabilidad nativa con Prometheus/Grafana y despliegue en contenedores de superficie de ataque mínima.

---

## 📋 Tabla de Contenidos

- [Stack Técnico](#-stack-técnico)
- [Arquitectura](#-arquitectura)
- [Endpoints de la API](#-endpoints-de-la-api)
- [Seguridad Implementada](#-seguridad-implementada)
- [Observabilidad](#-observabilidad)
- [Puesta en Marcha](#-puesta-en-marcha)
- [Variables de Entorno](#-variables-de-entorno)
- [Flujo de CI/CD](#-flujo-de-cicd)
- [Conceptos Clave Aplicados](#-conceptos-clave-aplicados)

---

## 🧱 Stack Técnico

| Capa | Tecnología | Versión | Rol |
|---|---|---|---|
| Lenguaje | Go | 1.23 | Motor principal — tipado estático, alto rendimiento |
| Router | go-chi/chi | v5.2.5 | Enrutamiento ligero y middleware-friendly |
| Seguridad | golang-jwt/jwt | v5.3.1 | Validación de tokens con firma HMAC-SHA256 |
| Observabilidad | prometheus/client_golang | v1.23.2 | Métricas de telemetría y runtime |
| Contenedores | Docker Multi-stage + `scratch` | — | Binarios estáticos sin SO subyacente |
| Orquestación | Docker Compose | — | App + Prometheus + Grafana |
| CI/CD | GitHub Actions + `gosec` | — | Pipeline con análisis de seguridad estático (SAST) |

---

## 🏗️ Arquitectura

Se aplica el **Standard Go Project Layout** para garantizar separación de responsabilidades:

```
proyecto2-go-audit/
├── cmd/
│   └── api/
│       └── main.go          # Punto de entrada: orquesta config y ciclo de vida del servidor
├── internal/
│   ├── config/
│   │   └── config.go        # Gestión de secretos via env vars (Patrón 12-Factor App)
│   ├── handlers/
│   │   ├── audit.go         # Lógica de negocio: ingesta de eventos de auditoría
│   │   └── health.go        # Healthcheck del servicio
│   └── middleware/
│       ├── jwt.go           # Capa de defensa: validación de JWT con verificación de algoritmo
│       └── auth.go          # Middleware de API Key (referencia histórica)
├── Dockerfile               # Build Multi-stage (builder: golang-alpine → runtime: scratch)
├── docker-compose.yml       # Stack completo: App + Prometheus + Grafana
└── prometheus.yml           # Configuración de scraping para el Pull Model
```

### Diagrama de Flujo de una Petición

```
Cliente HTTP
     │
     ▼
[chi Router]
     │
     ├─── GET  /health   ──────────────────────────────► HealthCheck Handler
     ├─── GET  /metrics  ──────────────────────────────► Prometheus Handler
     │
     └─── POST /audit
               │
               ▼
        [Middleware: RequireJWT]
          ┌─── Token ausente / inválido ──► 401 Unauthorized
          │
          └─── Token HMAC válido ─────────► CreateAuditLog Handler
                                                    │
                                                    ├── Decodifica JSON
                                                    ├── Imprime log estructurado
                                                    ├── Incrementa métricas (Prometheus Counter)
                                                    └── Responde 201 Created
```

---

## 🌐 Endpoints de la API

### `GET /health`
Verifica que el servicio está operativo. No requiere autenticación.

```bash
curl http://localhost:8081/health
```

**Respuesta `200 OK`:**
```json
{ "status": "ok" }
```

---

### `GET /metrics`
Expone métricas en formato Prometheus (Pull Model). Consumido por el scraper de Prometheus.

```bash
curl http://localhost:8081/metrics
```

Incluye métricas del runtime de Go (GC, memoria, goroutines) y la métrica custom `api_audit_events_total`.

---

### `POST /audit` 🔐 *Requiere JWT*

Registra un evento de auditoría. Requiere cabecera `Authorization: Bearer <token>`.

```bash
# 1. Generar un token JWT de prueba (HS256, secret: "mi-super-secreto-en-docker")
# Puedes usar jwt.io o cualquier cliente compatible.

# 2. Enviar el evento
curl -X POST http://localhost:8081/audit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TU_JWT_AQUI>" \
  -d '{
    "action": "USER_LOGIN",
    "user": "yago.gomez",
    "timestamp": "2026-04-29T10:00:00Z"
  }'
```

**Respuesta `201 Created`:**
```json
{
  "action": "USER_LOGIN",
  "status": "Evento de auditoría registrado exitosamente"
}
```

**Respuestas de error:**

| Código | Causa |
|---|---|
| `401` | Token ausente, con formato incorrecto o firma inválida |
| `401` | Intento de degradación de algoritmo (ej. `alg: none`) |
| `400` | Cuerpo JSON malformado |

---

## 🔐 Seguridad Implementada

### Validación Estricta de Algoritmo JWT

El vector de ataque más común en JWT es la degradación del algoritmo (`alg: none`). El middleware **verifica explícitamente** que el método de firma sea HMAC antes de validar la firma:

```go
// internal/middleware/jwt.go
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
    return nil, fmt.Errorf("método de firma inesperado o malicioso: %v", token.Header["alg"])
}
```

### Protección Anti-Slowloris

El servidor HTTP tiene timeouts estrictos configurados para evitar el agotamiento de conexiones por ataques Slowloris:

```go
// cmd/api/main.go
srv := &http.Server{
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
```

### Gestión de Secretos (12-Factor App)

Ningún secreto está hardcodeado. El programa **falla en el arranque** si `JWT_SECRET` no está definido como variable de entorno — un *fail-fast* intencional:

```go
// internal/config/config.go
if jwtSecret == "" {
    log.Fatal("JWT_SECRET no está definido")
}
```

### Imagen `scratch` — Superficie de Ataque Cero

La imagen de producción no contiene shell, gestor de paquetes ni librerías del sistema operativo. Solo el binario compilado estáticamente:

```dockerfile
FROM scratch
COPY --from=builder /app/audit-api /audit-api
```

---

## 📊 Observabilidad

### Stack de Monitorización

| Servicio | Puerto | Función |
|---|---|---|
| API | `8081` | Aplicación principal + `/metrics` |
| Prometheus | `9090` | Scraping y almacenamiento de métricas |
| Grafana | `3000` | Visualización y dashboards |

### Métrica Custom: `api_audit_events_total`

Un contador Prometheus registra cada evento de auditoría procesado exitosamente. Permite detectar picos de actividad o anomalías desde Grafana.

```
# HELP api_audit_events_total Cantidad total de eventos de auditoría registrados exitosamente
# TYPE api_audit_events_total counter
api_audit_events_total 42
```

### Acceso a Grafana

1. Navega a `http://localhost:3000` (usuario: `admin`, contraseña: `admin`)
2. Añade Prometheus como Data Source: `http://prometheus:9090`
3. Crea un panel con la query: `rate(api_audit_events_total[1m])`

---

## 🚀 Puesta en Marcha

### Prerrequisitos

- [Docker](https://www.docker.com/get-started) + Docker Compose
- [Go 1.23+](https://go.dev/dl/) (solo para desarrollo local)

### Con Docker Compose (Recomendado)

```bash
# Clona el repositorio
git clone https://github.com/YagoGomez83/proyecto2-go-audit.git
cd proyecto2-go-audit

# Levanta el stack completo (API + Prometheus + Grafana)
docker-compose up --build

# Para reconstruir sin caché
docker-compose build --no-cache && docker-compose up
```

### Ejecución Local (Desarrollo)

```bash
# Define la variable de entorno requerida
$env:JWT_SECRET = "mi-secreto-local"   # PowerShell
# export JWT_SECRET="mi-secreto-local"  # Bash

# Ejecuta directamente
go run ./cmd/api/main.go
```

---

## ⚙️ Variables de Entorno

| Variable | Requerida | Default | Descripción |
|---|---|---|---|
| `JWT_SECRET` | ✅ Sí | — | Clave para firmar y verificar tokens JWT. El servicio no arranca sin ella. |
| `PORT` | No | `8081` | Puerto en el que escucha el servidor HTTP. |

---

## 🔄 Flujo de CI/CD

El pipeline de GitHub Actions ejecuta análisis de seguridad estático (SAST) con `gosec` en cada push:

- **`gosec`** escanea el código fuente en busca de vulnerabilidades conocidas (timeouts faltantes, credenciales hardcodeadas, etc.) antes de cualquier despliegue.
- La compilación usa `CGO_ENABLED=0` y `-ldflags="-w -s"` para producir binarios estáticos, sin información de debug y resistentes a ingeniería inversa.

---

## 🧠 Conceptos Clave Aplicados

### DevSecOps — Security Shift Left

| Práctica | Implementación |
|---|---|
| SAST | `gosec` en el pipeline de CI analiza el código antes del build |
| Fail-Fast | El servicio no arranca sin secretos obligatorios (`log.Fatal`) |
| Defense in Depth | JWT en la capa de middleware, no en el handler |
| Least Privilege | Imagen `scratch`: sin shell, sin binarios del SO |

### Observabilidad — Pull Model vs Push Model

Prometheus usa un **Pull Model**: el servidor de métricas extrae datos del endpoint `/metrics` a intervalos regulares. Esto desacopla la aplicación del sistema de monitoreo y reduce la carga sobre la API.

### Inyección de Dependencias con Closures

Los middlewares reciben sus dependencias (como `JWT_SECRET`) mediante clausuras, sin romper la firma estándar `func(http.Handler) http.Handler` que espera chi:

```go
// El middleware "envuelve" el secreto en su scope de closure
r.With(middleware.RequireJWT(cfg.JWTSecret)).Post("/audit", handlers.CreateAuditLog)
```

### Docker Multi-stage Build

```
Etapa 1: golang:1.23-alpine  (~300 MB)
  └── go build -ldflags="-w -s" → binario estático (~7 MB)
                                          │
Etapa 2: scratch              (0 MB)     │
  └── COPY --from=builder ◄──────────────┘
  └── Imagen final: ~7 MB  ✅
```

---

## 📁 Proyectos de la Serie

| # | Proyecto | Descripción | Estado |
|---|---|---|---|
| 1 | API REST básica en Go | Fundamentos del lenguaje y routing | ✅ Completado |
| **2** | **Microservicio de Auditoría** | **JWT, Prometheus, Docker scratch** | **✅ Completado** |
| 3 | Persistencia con PostgreSQL | Integración de base de datos | 🔜 Próximo |
| 4 | Despliegue en Kubernetes | Orquestación y escalado | 🔜 Próximo |

---

<div align="center">
  <sub>Construido con enfoque DevSecOps • Go 1.23 • Seguridad primero</sub>
</div>
