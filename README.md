# go-flight-tracker

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![GraphQL](https://img.shields.io/badge/GraphQL-gqlgen-E10098?style=for-the-badge&logo=graphql&logoColor=white)](https://gqlgen.com/)
[![Redis](https://img.shields.io/badge/Redis-go--redis%2Fv9-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-CKAD-326CE5?style=for-the-badge&logo=kubernetes&logoColor=white)](https://kubernetes.io/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](./LICENSE)

> High-performance concurrent event engine in Go that tracks aircraft from the OpenSky Network API and streams real-time updates to React/Flutter clients via GraphQL Subscriptions (WebSockets).
>
> Engine de eventos de alta performance e concorrência em Go que rastreia aeronaves da API do OpenSky Network e transmite atualizações em tempo real para clientes (React/Flutter) via GraphQL Subscriptions (WebSockets).

---

<details open>
<summary><strong>🇧🇷 Português</strong></summary>

## Visão Geral

O **go-flight-tracker** é uma engine backend cloud-native focada em **concorrência idiomática em Go**, **fan-out de eventos em tempo real** e **prontidão para clusters Kubernetes**. O sistema faz polling de alta frequência na [OpenSky Network](https://opensky-network.org/), mantém estado concorrente em memória, propaga eventos via **Redis Pub/Sub** e entrega atualizações aos clientes através de **GraphQL Subscriptions** sobre WebSockets.

### Objetivos de Design

| Princípio | Implementação |
|---|---|
| Zero frameworks HTTP externos | `net/http` nativo — sem Chi, Gin ou Echo |
| Segurança contra data races | `sync.RWMutex` + canais tipados + `-race` |
| Escalabilidade horizontal | Redis Pub/Sub desacopla produtores e consumidores |
| Operabilidade em K8s | Liveness/Readiness probes, HPA, ConfigMaps/Secrets |
| Código idiomático | Goroutines, context cancellation, interfaces estreitas |

---

## Arquitetura & Fluxo de Dados

```
┌──────────────────┐     HTTP Poll      ┌─────────────────────────────┐
│  OpenSky Network │◄───────────────────│  Poller (Goroutine+Ticker)  │
│   (REST / states)│                    └──────────────┬──────────────┘
└──────────────────┘                                   │
                                                       ▼
                                        ┌─────────────────────────────┐
                                        │  In-Memory State Manager    │
                                        │  map[icao24]*Aircraft       │
                                        │  sync.RWMutex (R/W safe)    │
                                        └──────────────┬──────────────┘
                                                       │ publish
                                                       ▼
                                        ┌─────────────────────────────┐
                                        │  Redis                      │
                                        │  ├─ Pub/Sub (flight:events) │
                                        │  └─ Cache   (flight:state)  │
                                        └──────────────┬──────────────┘
                                                       │ subscribe
                                                       ▼
┌──────────────────┐     WebSocket      ┌─────────────────────────────┐
│ React / Flutter  │◄───────────────────│  GraphQL Subscription Hub   │
│     Clients      │   (gqlgen + WS)    │  gorilla/websocket           │
└──────────────────┘                    └─────────────────────────────┘
```

### Pipeline de Eventos

```
Ticker ──► Fetch OpenSky ──► Diff/Merge State ──► Redis PUBLISH
                                                         │
                              Clients ◄── GraphQL Sub ◄── Redis SUBSCRIBE
```

1. **Poller** — `time.Ticker` + goroutine dedicada consulta a OpenSky em intervalo configurável.
2. **State Manager** — estado em memória protegido por `sync.RWMutex`; leituras concorrentes sem bloqueio desnecessário.
3. **Redis** — Pub/Sub para fan-out entre réplicas; Cache para snapshot rápido (cold start / queries).
4. **GraphQL Hub** — subscriptions gqlgen sobre WebSocket entregam deltas aos clientes.

---

## Destaques Técnicos

### Concorrência & Safety
- Goroutines coordenadas por `context.Context` (graceful shutdown).
- Estado compartilhado com `sync.RWMutex` — padrão readers/writers.
- Canais tipados para desacoplar produção e consumo de eventos.
- Build/test com `-race` para detectar data races em CI.

### Stack Idiomática
- **Go 1.25** + `net/http` padrão da linguagem.
- **gqlgen** — schema-first GraphQL com code generation.
- **gorilla/websocket** — transporte de subscriptions em tempo real.
- **go-redis/v9** — Pub/Sub e cache distribuído.

### Cloud-Native / CKAD
- **Liveness Probe** — processo vivo (`/healthz`).
- **Readiness Probe** — Redis alcançável e poller inicializado (`/readyz`).
- **HPA** — escala por CPU/memória sob carga de subscriptions.
- **ConfigMaps & Secrets** — intervalo de polling, endpoints e credenciais Redis.
- **Docker multi-stage** — imagem final mínima (distroless/scratch-friendly).

---

## Stack

| Camada | Tecnologia |
|---|---|
| Runtime | Go 1.25 |
| HTTP | `net/http` (stdlib) |
| GraphQL | [gqlgen](https://github.com/99designs/gqlgen) |
| WebSocket | [gorilla/websocket](https://github.com/gorilla/websocket) |
| Cache / PubSub | Redis + [go-redis/v9](https://github.com/redis/go-redis) |
| Fonte de dados | [OpenSky Network API](https://opensky-network.org/) |
| Clientes | React / Flutter |
| Infra | Docker, Docker Compose, Kubernetes |

---

## Pré-requisitos

- [Go 1.25+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/)
- (Opcional) `kubectl` + cluster local (kind/minikube) para manifests K8s

---

## Execução Local (Docker Compose)

```bash
# Clone
git clone https://github.com/mateusfmfm/go-flight-tracker.git
cd go-flight-tracker

# Sobe API + Redis
docker compose up --build

# GraphQL Playground
open http://localhost:8080/
```

Variáveis de ambiente típicas:

| Variável | Descrição | Default |
|---|---|---|
| `PORT` | Porta HTTP do servidor | `8080` |
| `REDIS_ADDR` | Endereço do Redis | `redis:6379` |
| `POLL_INTERVAL` | Intervalo do poller OpenSky | `10s` |
| `OPENSKY_URL` | Endpoint base da OpenSky | OpenSky public API |

### Desenvolvimento sem Docker (API)

```bash
# Redis local (exemplo)
docker run -d --name redis -p 6379:6379 redis:7-alpine

export REDIS_ADDR=localhost:6379
export PORT=8080

go run ./server.go
```

### Geração GraphQL (gqlgen)

```bash
go tool gqlgen generate
```

---

## Health Checks (Kubernetes)

Endpoints previstos para probes:

| Path | Tipo | Critério |
|---|---|---|
| `GET /healthz` | Liveness | Processo respondendo |
| `GET /readyz` | Readiness | Redis OK + estado inicializado |

Exemplo de probes no Deployment:

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  initialDelaySeconds: 3
  periodSeconds: 5
```

---

## Estrutura do Projeto

```
go-flight-tracker/
├── server.go                 # Entrypoint (net/http + gqlgen)
├── graph/
│   ├── schema.graphqls       # Schema GraphQL (schema-first)
│   ├── schema.resolvers.go   # Resolvers (queries/mutations/subscriptions)
│   ├── resolver.go           # Dependências do Resolver
│   ├── generated.go          # Código gerado pelo gqlgen
│   └── model/                # Models gerados
├── gqlgen.yml                # Configuração do code generation
├── go.mod / go.sum
├── docker-compose.yml        # (planejado) API + Redis
├── Dockerfile                # (planejado) multi-stage build
└── k8s/                      # (planejado) Deployment, Service, HPA, ConfigMap
```

---

## Roadmap

- [x] Bootstrap gqlgen + `net/http`
- [x] Poller OpenSky com ticker e context
- [x] State manager concorrente (`sync.RWMutex`)
- [x] Integração Redis Pub/Sub + cache
- [x] GraphQL Subscriptions via WebSocket
- [x] Docker Compose + health endpoints
- [x] Manifests Kubernetes (probes, HPA, ConfigMaps)

---

## Autor

**Mateus Félix** — Engenheiro de Software com foco em Go, sistemas concorrentes e infraestrutura cloud-native (CKAD).

- GitHub: [@mateusfmfm](https://github.com/mateusfmfm)
- E-mail: [mateusfmfm@outlook.com](mailto:mateusfmfm@outlook.com)
- Repositório: [go-flight-tracker](https://github.com/mateusfmfm/go-flight-tracker)

---

## Licença

Distribuído sob a licença [MIT](./LICENSE).

</details>

---

<details open>
<summary><strong>🇺🇸 English</strong></summary>

## Overview

**go-flight-tracker** is a cloud-native backend engine focused on **idiomatic Go concurrency**, **real-time event fan-out**, and **Kubernetes cluster readiness**. It performs high-frequency polling against the [OpenSky Network](https://opensky-network.org/), keeps concurrent in-memory state, propagates events via **Redis Pub/Sub**, and delivers updates to clients through **GraphQL Subscriptions** over WebSockets.

### Design Goals

| Principle | Implementation |
|---|---|
| No external HTTP frameworks | Native `net/http` — no Chi, Gin, or Echo |
| Data-race safety | `sync.RWMutex` + typed channels + `-race` |
| Horizontal scalability | Redis Pub/Sub decouples producers and consumers |
| K8s operability | Liveness/Readiness probes, HPA, ConfigMaps/Secrets |
| Idiomatic code | Goroutines, context cancellation, narrow interfaces |

---

## Architecture & Data Flow

```
┌──────────────────┐     HTTP Poll      ┌─────────────────────────────┐
│  OpenSky Network │◄───────────────────│  Poller (Goroutine+Ticker)  │
│   (REST / states)│                    └──────────────┬──────────────┘
└──────────────────┘                                   │
                                                       ▼
                                        ┌─────────────────────────────┐
                                        │  In-Memory State Manager    │
                                        │  map[icao24]*Aircraft       │
                                        │  sync.RWMutex (R/W safe)    │
                                        └──────────────┬──────────────┘
                                                       │ publish
                                                       ▼
                                        ┌─────────────────────────────┐
                                        │  Redis                      │
                                        │  ├─ Pub/Sub (flight:events) │
                                        │  └─ Cache   (flight:state)  │
                                        └──────────────┬──────────────┘
                                                       │ subscribe
                                                       ▼
┌──────────────────┐     WebSocket      ┌─────────────────────────────┐
│ React / Flutter  │◄───────────────────│  GraphQL Subscription Hub   │
│     Clients      │   (gqlgen + WS)    │  gorilla/websocket           │
└──────────────────┘                    └─────────────────────────────┘
```

### Event Pipeline

```
Ticker ──► Fetch OpenSky ──► Diff/Merge State ──► Redis PUBLISH
                                                         │
                              Clients ◄── GraphQL Sub ◄── Redis SUBSCRIBE
```

1. **Poller** — Dedicated goroutine + `time.Ticker` hits OpenSky on a configurable interval.
2. **State Manager** — In-memory state guarded by `sync.RWMutex` for concurrent readers.
3. **Redis** — Pub/Sub for fan-out across replicas; Cache for fast snapshots (cold start / queries).
4. **GraphQL Hub** — gqlgen subscriptions over WebSocket push deltas to clients.

---

## Technical Highlights

### Concurrency & Safety
- Goroutines coordinated via `context.Context` (graceful shutdown).
- Shared state with `sync.RWMutex` — readers/writers pattern.
- Typed channels to decouple event production and consumption.
- Build/test with `-race` to catch data races in CI.

### Idiomatic Stack
- **Go 1.25** + standard-library `net/http`.
- **gqlgen** — schema-first GraphQL with code generation.
- **gorilla/websocket** — real-time subscription transport.
- **go-redis/v9** — distributed Pub/Sub and cache.

### Cloud-Native / CKAD
- **Liveness Probe** — process alive (`/healthz`).
- **Readiness Probe** — Redis reachable and poller initialized (`/readyz`).
- **HPA** — scale on CPU/memory under subscription load.
- **ConfigMaps & Secrets** — poll interval, endpoints, and Redis credentials.
- **Multi-stage Docker** — minimal final image (distroless/scratch-friendly).

---

## Stack

| Layer | Technology |
|---|---|
| Runtime | Go 1.25 |
| HTTP | `net/http` (stdlib) |
| GraphQL | [gqlgen](https://github.com/99designs/gqlgen) |
| WebSocket | [gorilla/websocket](https://github.com/gorilla/websocket) |
| Cache / PubSub | Redis + [go-redis/v9](https://github.com/redis/go-redis) |
| Data source | [OpenSky Network API](https://opensky-network.org/) |
| Clients | React / Flutter |
| Infra | Docker, Docker Compose, Kubernetes |

---

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/)
- (Optional) `kubectl` + local cluster (kind/minikube) for K8s manifests

---

## Local Run (Docker Compose)

```bash
# Clone
git clone https://github.com/mateusfmfm/go-flight-tracker.git
cd go-flight-tracker

# Start API + Redis
docker compose up --build

# GraphQL Playground
open http://localhost:8080/
```

Typical environment variables:

| Variable | Description | Default |
|---|---|---|
| `PORT` | HTTP server port | `8080` |
| `REDIS_ADDR` | Redis address | `redis:6379` |
| `POLL_INTERVAL` | OpenSky poller interval | `10s` |
| `OPENSKY_URL` | OpenSky base endpoint | OpenSky public API |

### Development without Docker (API)

```bash
# Local Redis (example)
docker run -d --name redis -p 6379:6379 redis:7-alpine

export REDIS_ADDR=localhost:6379
export PORT=8080

go run ./server.go
```

### GraphQL Code Generation (gqlgen)

```bash
go tool gqlgen generate
```

---

## Health Checks (Kubernetes)

Planned probe endpoints:

| Path | Type | Criteria |
|---|---|---|
| `GET /healthz` | Liveness | Process responding |
| `GET /readyz` | Readiness | Redis OK + state initialized |

Example Deployment probes:

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  initialDelaySeconds: 3
  periodSeconds: 5
```

---

## Project Structure

```
go-flight-tracker/
├── server.go                 # Entrypoint (net/http + gqlgen)
├── graph/
│   ├── schema.graphqls       # GraphQL schema (schema-first)
│   ├── schema.resolvers.go   # Resolvers (queries/mutations/subscriptions)
│   ├── resolver.go           # Resolver dependencies
│   ├── generated.go          # gqlgen-generated code
│   └── model/                # Generated models
├── gqlgen.yml                # Code generation config
├── go.mod / go.sum
├── docker-compose.yml        # (planned) API + Redis
├── Dockerfile                # (planned) multi-stage build
└── k8s/                      # (planned) Deployment, Service, HPA, ConfigMap
```

---

## Roadmap

- [x] gqlgen + `net/http` bootstrap
- [x] OpenSky poller with ticker and context
- [x] Concurrent state manager (`sync.RWMutex`)
- [x] Redis Pub/Sub + cache integration
- [x] GraphQL Subscriptions over WebSocket
- [x] Docker Compose + health endpoints
- [x] Kubernetes manifests (probes, HPA, ConfigMaps)

---

## Author

**Mateus Félix** — Software Engineer focused on Go, concurrent systems, and cloud-native infrastructure (CKAD).

- GitHub: [@mateusfmfm](https://github.com/mateusfmfm)
- Email: [mateusfmfm@outlook.com](mailto:mateusfmfm@outlook.com)
- Repository: [go-flight-tracker](https://github.com/mateusfmfm/go-flight-tracker)

---

## License

Distributed under the [MIT](./LICENSE) license.

</details>
