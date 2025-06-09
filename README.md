# 📘 Ad Campaign Delivery Microservice

A high-performance, horizontally scalable microservice built in Go that delivers active ad campaigns to users based on dynamic targeting rules like `app`, `country`, and `OS`.

---

## 🚀 Features

- ⚡️ High-throughput REST API to match campaigns
- 🔁 Redis caching for campaigns and rules
- 📊 Prometheus metrics integration
- 🗂️ Master-Slave PostgreSQL replication
- 🔄 Round-robin slave read strategy with atomic access
- 🔌 Middleware for:
  - Input sanitization
  - Request validation
  - Prometheus instrumentation
- 🧪 Load tested using `wrk` — 4K+ RPS sustained

---

## 🧱 Data Models

### 🗂️ `Campaign`

```go
type Campaign struct {
    ID        uint64
    Name      string
    Image     string
    CTA       string
    Status    Status // ACTIVE / INACTIVE
    ...
}
```

### 🎯 `TargetingRule`

```go
type TargetingRule struct {
    CampaignID    uint64
    DimensionType string // "app", "country", "os"
    Include       bool   // true = include, false = exclude
    Value         string
}
```

---

## ⚙️ Architecture Overview

```
                  +-------------------------+
                  |     Load Balancer       |
                  +-----------+-------------+
                              |
                   +----------v-----------+
                   |   Delivery Service   |
                   |  (RESTful HTTP API)  |
                   +----------+-----------+
                              |
        +---------------------+----------------------+
        |                    Middleware              |
        |--------------------------------------------|
        | ✅ Validate Query Params (app, os, country) |
        | ✅ Prometheus Metrics                       |
        | ✅ App Existence via Redis/DB               |
        +---------------------+----------------------+
                              |
               +-------------v-------------+
               |   Delivery Rule Engine     |
               |----------------------------|
               | - Get campaign IDs by app  |
               | - Get rules by campaign    |
               | - Match OS & country       |
               +-------------+--------------+
                             |
     +-----------------------+------------------------+
     |                Service Layer (Go)              |
     |------------------------------------------------|
     | - Redis cache lookup                           |
     | - DB fallback + set cache                      |
     | - Filter by status (only ACTIVE)               |
     +----------------------+-------------------------+
                            |
           +----------------v----------------+
           |     PostgreSQL (Master-Slaves)  |
           | - Master for writes             |
           | - Slaves for reads (RR policy)  |
           +---------------------------------+
```

---

## ⚡️ Load Testing Results

```bash
Command:
wrk -t12 -c400 -d300s \
"http://host.docker.internal:8081/internal/api/v1/delivery?app=com.ludo.king&country=us&os=android"

Results:
- Total requests:      1,260,045
- Requests/second:     ~4216
- Avg latency:         ~94 ms
- Timeouts:            3446
- Errors:              0 (connect/read/write)
- Data transferred:    ~247 MB
```

> 💡 This system demonstrated excellent stability at high concurrency.

---

## 🔁 Scaling Strategy

### ✅ Vertical Scaling (Phase 1)
- Upgrade EC2 instance: CPU, memory, IOPS
- Benefit: Simpler to implement

### ✅ Horizontal Scaling (Phase 2)
- Add more nodes behind load balancer
- Stateless app = easy replication
- Use database sharding:
  - Example: User ID `1-1e8` → Asia DB shard
  - User ID `>1e8` → US/Europe shard

---

## 🧠 Optimization Techniques

| Area                        | Strategy                                                                                         |
| --------------------------- | ------------------------------------------------------------------------------------------------ |
| **DB Reads**                | Read from slaves in round-robin using `atomic.AddUint64`                                         |
| **Caching**                 | Redis with 1-day TTL for campaigns & rules                                                       |
| **Load Balancing**          | AWS ALB                                                                                          |
| **Monitoring**              | Prometheus + Grafana                                                                             |
| **Targeting Rule Indexing** | ✅ Unique Index on `(campaign_id, dimension_type, value)`<br>✅ Index on `(dimension_type, value, Include)` |


---

## 📦 Tech Stack

| Component       | Stack             |
|----------------|--------------------|
| Language        | Go (1.23)         |
| HTTP Server     | Gin                |
| Database        | PostgreSQL (master + slave) |
| Cache           | Redis              |
| Metrics         | Prometheus         |
| Load Testing    | `wrk`              |
| Infra (scaling) | EC2 / Docker       |

---

## 🧪 Local Setup

### Run Load Test
```bash
wrk -t12 -c400 -d300s \
"http://host.docker.internal:8081/internal/api/v1/delivery?app=com.ludo.king&country=us&os=android"
```

---
