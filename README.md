# Research Compute Queue API

A distributed job scheduling system designed for academic research computing workloads. Features priority-based fair-share scheduling, resource management, and job dependency handling.

## 🎯 Features

- **RESTful API** for job submission and management
- **Priority-based scheduling** with fair-share allocation across research groups
- **Resource matching** (CPU, memory, GPU) to available compute nodes
- **Job dependencies** for complex workflows (DAG execution)
- **User authentication** with JWT tokens
- **Usage tracking** and quota management
- **Real-time job monitoring** and log retrieval
- **Admin dashboard** for system analytics

## 🏗️ Architecture
```
┌─────────────┐      HTTP/REST       ┌──────────────┐
│   Client    │ ──────────────────> │   API Server │
│  (curl,     │ <────────────────── │   (Go/Gin)   │
│  Postman)   │      JSON            └──────┬───────┘
└─────────────┘                             │
                                            │
                    ┌───────────────────────┼──────────────┐
                    ▼                       ▼              ▼
             ┌─────────────┐        ┌─────────────┐  ┌──────────┐
             │ PostgreSQL  │        │  Scheduler  │  │   File   │
             │  Database   │        │  (Goroutine)│  │  Storage │
             │             │        │             │  │          │
             │ - Users     │        │ - Priority  │  │ - Logs   │
             │ - Jobs      │        │ - Matching  │  │ - Output │
             │ - Groups    │        │ - Fair-share│  └──────────┘
             └─────────────┘        └─────────────┘
```

## 🚀 Tech Stack

- **Language:** Go 1.21+
- **Web Framework:** Gin
- **Database:** PostgreSQL 15
- **Authentication:** JWT (golang-jwt)
- **API Style:** REST
- **Testing:** Go testing + testify

## 📋 Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Git

## 🛠️ Installation

1. **Clone the repository**
```bash
   git clone https://github.com/YOUR_USERNAME/research-compute-queue.git
   cd research-compute-queue
```

2. **Install dependencies**
```bash
   go mod download
```

3. **Set up PostgreSQL**
```bash
   # Using Docker (recommended)
   docker run --name research-queue-db \
     -e POSTGRES_PASSWORD=dev123 \
     -e POSTGRES_DB=research_queue \
     -p 5432:5432 \
     -d postgres:15
   
   # Create database schema
   psql -h localhost -U postgres -d research_queue -f scripts/setup_db.sql
```

4. **Configure environment variables**
```bash
   cp .env.example .env
   # Edit .env with your database credentials
```

5. **Run the server**
```bash
   go run cmd/server/main.go
```

   Server will start on `http://localhost:8080`

## 📖 API Documentation

### Authentication

#### Register User
```bash
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword",
  "group_id": 1
}
```

#### Login
```bash
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}

Response: { "token": "eyJhbGciOiJ..." }
```

### Job Management

#### Submit Job
```bash
POST /api/jobs
Authorization: Bearer <token>
Content-Type: application/json

{
  "script": "python train_model.py",
  "cpu_cores": 8,
  "memory_gb": 32,
  "gpu_count": 1,
  "estimated_hours": 4.5,
  "priority": 2
}
```

#### Get Job Status
```bash
GET /api/jobs/{job_id}
Authorization: Bearer <token>
```

#### List Jobs
```bash
GET /api/jobs?status=running&limit=10
Authorization: Bearer <token>
```

#### Cancel Job
```bash
DELETE /api/jobs/{job_id}
Authorization: Bearer <token>
```

### Queue & Monitoring

#### View Queue
```bash
GET /api/queue
Authorization: Bearer <token>
```

#### Get Job Logs
```bash
GET /api/jobs/{job_id}/logs
Authorization: Bearer <token>
```

## 🧪 Testing
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test ./internal/scheduler -v
```

## 🗂️ Project Structure
```
research-compute-queue/
├── cmd/server/          # Application entry point
├── internal/
│   ├── api/             # HTTP handlers and routing
│   ├── models/          # Data structures
│   ├── database/        # Database operations
│   ├── scheduler/       # Job scheduling logic
│   └── config/          # Configuration management
├── scripts/             # Database setup and utilities
├── tests/               # Test files
└── docs/                # Additional documentation
```

## 🔧 Configuration

Key configuration options in `.env`:

- `DATABASE_URL`: PostgreSQL connection string
- `PORT`: API server port (default: 8080)
- `JWT_SECRET`: Secret key for JWT signing
- `SCHEDULER_INTERVAL_SECONDS`: How often scheduler runs (default: 30)
- `MAX_CONCURRENT_JOBS`: Maximum parallel jobs (default: 10)

## 🎯 Scheduling Algorithm

The scheduler uses a **priority-based fair-share algorithm**:
```
final_priority = base_priority × fair_share_multiplier × wait_time_boost

Where:
- base_priority: User-defined (1-10)
- fair_share_multiplier: group_quota / actual_usage
- wait_time_boost: 1 + (wait_minutes / 60)
```

Jobs are scheduled in priority order, with backfilling for smaller jobs to maximize cluster utilization.

## 🚧 Roadmap

- [ ] Basic API server with authentication
- [ ] Job submission and status tracking
- [ ] Priority-based scheduling
- [ ] Fair-share algorithm
- [ ] Job dependencies (DAG)
- [ ] WebSocket for real-time logs
- [ ] Redis integration
- [ ] Multi-node worker support
- [ ] Web UI dashboard
- [ ] Email notifications
- [ ] Prometheus metrics

## 🤝 Contributing

This is a learning project, but suggestions and feedback are welcome! Feel free to open issues or submit PRs.

## 📝 License

MIT License - see LICENSE file for details

## 👤 Author

**Samik Kundu**
- GitHub: [@samik-k21](https://github.com/samik-k21)
- LinkedIn: [samik-kundu](https://linkedin.com/in/samik-kundu-862753338/)
- Email: skundu2448@gmail.com

## 🙏 Acknowledgments

Built as a portfolio project to learn about API development, distributed systems, and infrastructure software. Inspired by job schedulers like Slurm, PBS, and Kubernetes.

---

**Status:** 🚧 Work in Progress - actively being developed during winter break 2025