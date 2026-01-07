# Research Compute Queue API

A distributed job scheduling system designed for academic research computing workloads. Features priority-based fair-share scheduling, resource management, and concurrent job execution.

![Status](https://img.shields.io/badge/status-active-success.svg)
![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

## 🎯 Project Overview

This project is a production-grade REST API that simulates a job scheduling system for research computing clusters. It implements sophisticated scheduling algorithms including fair-share resource allocation, priority-based queuing, and automatic resource matching.

**Built as a learning project to demonstrate:**
- RESTful API design and implementation
- Distributed systems concepts (scheduling, resource management)
- Database design and SQL optimization
- JWT authentication and authorization
- Concurrent programming with goroutines
- Infrastructure software development

## ✨ Features

### Core Functionality
- **RESTful API** for job submission and management
- **JWT Authentication** with secure token generation and validation
- **Priority-based scheduling** with configurable job priorities
- **Fair-share algorithm** - ensures equitable resource distribution across research groups
- **Resource matching** - automatically matches jobs to workers with sufficient CPU, memory, and GPU
- **Concurrent execution** - runs multiple jobs simultaneously with configurable limits
- **Real-time monitoring** - track job status (pending → running → completed/failed)
- **User isolation** - users can only view and manage their own jobs
- **Usage tracking** - logs CPU hours for fair-share calculations

### Scheduling Algorithm
The scheduler uses a sophisticated multi-factor priority calculation:
```
final_priority = base_priority × fair_share_multiplier × wait_time_boost

Where:
- base_priority: User + group priority (1-10)
- fair_share_multiplier: quota / actual_usage (prevents resource hogging)
- wait_time_boost: 1 + (wait_minutes / 60 * 0.01) (prevents starvation)
```

**Example:**
- Group A used 90% of quota → fair_share = 1.11 (slight boost)
- Group B used 25% of quota → fair_share = 2.0 (high boost)
- Job waiting 10 hours → wait_boost = 1.10
- Result: Group B's jobs get scheduled first, older jobs gradually gain priority

## 🏗️ Architecture
```
┌─────────────┐      HTTP/REST       ┌──────────────┐
│   Client    │ ──────────────────>  │   API Server │
│  (curl,     │ <──────────────────  │   (Go/Gin)   │
│  Postman)   │      JSON            └──────┬───────┘
└─────────────┘                             │
                                            │
                    ┌───────────────────────┼──────────────────┐
                    ▼                       ▼                  ▼
             ┌─────────────┐        ┌─────────────┐     ┌──────────┐
             │ PostgreSQL  │        │  Scheduler  │     │   File   │
             │  Database   │        │  (Goroutine)│     │  Storage │
             │             │        │             │     │          │
             │ - Users     │        │ - Priority  │     │ - Logs   │
             │ - Jobs      │        │ - Matching  │     │ - Output │
             │ - Groups    │        │ - Fair-share│     └──────────┘
             │ - Workers   │        │ - Executor  │
             └─────────────┘        └─────────────┘
```

### Component Breakdown

**API Server (Go + Gin)**
- Handles HTTP requests and responses
- JWT authentication middleware
- Request validation and error handling
- Routes: `/auth`, `/jobs`, `/queue`, `/admin`

**Scheduler (Background Goroutine)**
- Runs every 30 seconds (configurable)
- Fetches pending jobs from database
- Calculates priorities using fair-share algorithm
- Matches jobs to available workers
- Starts job execution and tracks completion

**Database (PostgreSQL)**
- Stores users, groups, jobs, workers
- Tracks resource usage for fair-share
- ACID transactions for job state changes
- Indexed for fast queries

## 🚀 Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Language** | Go 1.21+ | High-performance, concurrent programming |
| **Web Framework** | Gin | Fast HTTP routing and middleware |
| **Database** | PostgreSQL 15 | Relational data storage with ACID guarantees |
| **Authentication** | JWT (golang-jwt) | Stateless authentication |
| **Password Hashing** | bcrypt | Secure password storage |
| **Containerization** | Docker | Database isolation and portability |
| **API Design** | REST | Standard HTTP methods and status codes |

## 📋 Prerequisites

- **Go 1.21 or higher** - [Install Go](https://go.dev/dl/)
- **PostgreSQL 15 or higher** - Via Docker (recommended) or local install
- **Docker** (optional but recommended) - [Install Docker](https://docs.docker.com/get-docker/)
- **Git** - For cloning the repository
- **curl** or **Postman** - For testing API endpoints

## 🛠️ Installation & Setup

### 1. Clone the Repository
```bash
git clone https://github.com/YOUR_USERNAME/research-compute-queue.git
cd research-compute-queue
```

### 2. Install Go Dependencies
```bash
go mod download
```

### 3. Set Up PostgreSQL Database

**Option A: Using Docker (Recommended)**
```bash
# Start PostgreSQL container
docker run --name research-queue-db \
  -e POSTGRES_PASSWORD=dev123 \
  -e POSTGRES_DB=research_queue \
  -p 5432:5432 \
  -d postgres:15

# Verify it's running
docker ps
```

**Option B: Local PostgreSQL**
```bash
# macOS with Homebrew
brew install postgresql@15
brew services start postgresql@15
createdb research_queue

# Ubuntu/Debian
sudo apt install postgresql-15
sudo systemctl start postgresql
sudo -u postgres createdb research_queue
```

### 4. Create Database Schema
```bash
# Using Docker
docker exec -i research-queue-db psql -U postgres -d research_queue < scripts/setup_db.sql

# Using local PostgreSQL
psql -U postgres -d research_queue -f scripts/setup_db.sql
```

You should see:
```
CREATE TABLE
CREATE TABLE
CREATE TABLE
...
INSERT 0 3
INSERT 0 3
```

### 5. Configure Environment Variables
```bash
# Copy example config
cp .env.example .env

# Edit .env with your values
# Make sure DATABASE_URL matches your setup
```

**`.env` file:**
```bash
DATABASE_URL=postgres://postgres:dev123@localhost:5432/research_queue?sslmode=disable
PORT=8080
ENVIRONMENT=development
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY_HOURS=24
SCHEDULER_INTERVAL_SECONDS=30
MAX_CONCURRENT_JOBS=10
LOG_DIRECTORY=./logs
OUTPUT_DIRECTORY=./output
```

### 6. Run the Server
```bash
go run cmd/server/main.go
```

**Expected output:**
```
========================================
Research Compute Queue API
Environment: development
========================================
✓ Database connection established
✓ JWT manager initialized
✓ Directories created
✓ Scheduler started (interval: 30s, max concurrent: 10)
✓ API server starting on port 8080
========================================
System is ready!
API: http://localhost:8080
Press Ctrl+C to stop
========================================
```

## 📖 API Documentation

### Base URL
```
http://localhost:8080
```

### Authentication

All `/api/jobs` endpoints require a valid JWT token in the `Authorization` header:
```
Authorization: Bearer <your_jwt_token>
```

---

### Health Check

**Check API Status**
```bash
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "message": "Research Compute Queue API is running",
  "version": "1.0.0"
}
```

---

### Authentication Endpoints

#### Register User
```bash
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123",
  "group_id": 1
}
```

**Response:**
```json
{
  "message": "User registered successfully",
  "user_id": 2
}
```

#### Login
```bash
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 2,
    "email": "user@example.com",
    "group_id": 1,
    "is_admin": false
  }
}
```

---

### Job Endpoints

#### Submit Job
```bash
POST /api/jobs
Authorization: Bearer <token>
Content-Type: application/json

{
  "script": "python train_model.py --epochs 100",
  "cpu_cores": 8,
  "memory_gb": 32,
  "gpu_count": 1,
  "estimated_hours": 4.5,
  "priority": 3
}
```

**Response:**
```json
{
  "message": "Job submitted successfully",
  "job_id": 1,
  "status": "pending"
}
```

#### Get Job Status
```bash
GET /api/jobs/{job_id}
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": 1,
  "user_id": 2,
  "group_id": 1,
  "script": "python train_model.py --epochs 100",
  "cpu_cores": 8,
  "memory_gb": 32,
  "gpu_count": 1,
  "status": "running",
  "priority": 3,
  "submitted_at": "2026-01-08T15:30:00Z",
  "started_at": "2026-01-08T15:30:30Z",
  "completed_at": null,
  "worker_id": 2
}
```

#### List Jobs
```bash
GET /api/jobs?status=running&limit=10
Authorization: Bearer <token>
```

**Query Parameters:**
- `status` (optional): Filter by status (`pending`, `running`, `completed`, `failed`, `cancelled`)
- `limit` (optional): Max number of results (default: 50)

**Response:**
```json
{
  "jobs": [
    {
      "id": 1,
      "status": "running",
      "script": "python train_model.py",
      "cpu_cores": 8,
      "submitted_at": "2026-01-08T15:30:00Z"
    }
  ],
  "count": 1
}
```

#### Cancel Job
```bash
DELETE /api/jobs/{job_id}
Authorization: Bearer <token>
```

**Response:**
```json
{
  "message": "Job cancelled successfully",
  "job_id": 1
}
```

---

## 🧪 Testing

### Quick Test Script

Save this as `test.sh`:
```bash
#!/bin/bash

API="http://localhost:8080"

echo "=== Testing Research Compute Queue API ==="

# 1. Health check
echo -e "\n1. Health Check:"
curl -s $API/health | jq

# 2. Register user
echo -e "\n2. Register User:"
curl -s -X POST $API/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","group_id":1}' | jq

# 3. Login and get token
echo -e "\n3. Login:"
TOKEN=$(curl -s -X POST $API/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.token')

echo "Token: ${TOKEN:0:50}..."

# 4. Submit job
echo -e "\n4. Submit Job:"
JOB_ID=$(curl -s -X POST $API/api/jobs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"script":"python test.py","cpu_cores":4,"memory_gb":16,"priority":3}' \
  | jq -r '.job_id')

echo "Created Job ID: $JOB_ID"

# 5. Get job status
echo -e "\n5. Get Job Status:"
curl -s $API/api/jobs/$JOB_ID \
  -H "Authorization: Bearer $TOKEN" | jq

# 6. List all jobs
echo -e "\n6. List All Jobs:"
curl -s "$API/api/jobs" \
  -H "Authorization: Bearer $TOKEN" | jq

echo -e "\n=== Test Complete ==="
```

**Run tests:**
```bash
chmod +x test.sh
./test.sh
```

### Manual Testing Examples

**1. Register and Login:**
```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@wisc.edu","password":"password123","group_id":1}'

# Login and save token
export TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@wisc.edu","password":"password123"}' \
  | grep -o '"token":"[^"]*' | cut -d'"' -f4)
```

**2. Submit and Monitor Jobs:**
```bash
# Submit job
curl -X POST http://localhost:8080/api/jobs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "script": "python train.py",
    "cpu_cores": 8,
    "memory_gb": 32,
    "gpu_count": 1,
    "priority": 5
  }'

# List all jobs
curl http://localhost:8080/api/jobs \
  -H "Authorization: Bearer $TOKEN"

# Get specific job
curl http://localhost:8080/api/jobs/1 \
  -H "Authorization: Bearer $TOKEN"

# Filter by status
curl "http://localhost:8080/api/jobs?status=running" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🗂️ Project Structure
```
research-compute-queue/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/            # HTTP request handlers
│   │   │   ├── auth.go         # Registration & login
│   │   │   ├── jobs.go         # Job management
│   │   │   └── health.go       # Health check
│   │   ├── middleware/          # HTTP middleware
│   │   │   ├── auth.go         # JWT validation
│   │   │   └── logging.go      # Request logging
│   │   └── router.go            # Route definitions
│   ├── auth/
│   │   └── jwt.go              # JWT token generation/validation
│   ├── models/                  # Data structures
│   │   ├── user.go             # User & Group models
│   │   └── job.go              # Job models
│   ├── database/                # Database operations
│   │   └── postgres.go         # PostgreSQL connection
│   ├── scheduler/               # Job scheduling logic
│   │   ├── scheduler.go        # Main scheduler loop
│   │   ├── priority.go         # Priority calculation
│   │   ├── matcher.go          # Resource matching
│   │   └── executor.go         # Job execution
│   └── config/
│       └── config.go            # Configuration loading
├── scripts/
│   └── setup_db.sql             # Database schema
├── .env                         # Environment variables (not committed)
├── .env.example                 # Example environment config
├── .gitignore                   # Git ignore rules
├── go.mod                       # Go dependencies
├── go.sum                       # Go dependency checksums
├── LICENSE                      # MIT License
└── README.md                    # This file
```

---

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `PORT` | API server port | `8080` |
| `ENVIRONMENT` | Environment mode (`development`, `production`) | `development` |
| `JWT_SECRET` | Secret key for JWT signing | Required |
| `JWT_EXPIRY_HOURS` | JWT token validity duration | `24` |
| `SCHEDULER_INTERVAL_SECONDS` | How often scheduler runs | `30` |
| `MAX_CONCURRENT_JOBS` | Max simultaneous jobs | `10` |
| `LOG_DIRECTORY` | Directory for job logs | `./logs` |
| `OUTPUT_DIRECTORY` | Directory for job outputs | `./output` |

---

## 🎯 Database Schema

### Key Tables

**users** - User accounts with authentication
```sql
- id: Primary key
- email: Unique email address
- password_hash: bcrypt hashed password
- group_id: Foreign key to groups
- is_admin: Admin flag
```

**groups** - Research groups with resource quotas
```sql
- id: Primary key
- name: Group name
- cpu_quota: Monthly CPU hour quota
- priority: Base group priority (1-10)
```

**jobs** - Compute jobs
```sql
- id: Primary key
- user_id, group_id: Foreign keys
- script: Command to execute
- cpu_cores, memory_gb, gpu_count: Resource requirements
- status: pending/running/completed/failed/cancelled
- priority: Job priority (1-10)
- submitted_at, started_at, completed_at: Timestamps
```

**workers** - Compute nodes
```sql
- id: Primary key
- hostname: Worker identifier
- cpu_cores, memory_gb, gpu_count: Available resources
- status: idle/busy/offline
```

**usage_logs** - Resource usage tracking for fair-share
```sql
- group_id: Foreign key to groups
- job_id: Foreign key to jobs
- cpu_hours_used: Calculated CPU hours
- logged_at: Timestamp
```

---

## 🚧 Roadmap & Future Enhancements

- [ ] **Job Dependencies** - DAG-based workflow execution
- [ ] **Queue Viewing Endpoints** - See pending jobs and estimated wait times
- [ ] **Admin Dashboard API** - System-wide statistics and management
- [ ] **WebSocket Support** - Real-time log streaming
- [ ] **Redis Integration** - Improved queue performance and caching
- [ ] **Multi-node Workers** - Actual distributed execution
- [ ] **Email Notifications** - Notify users on job completion
- [ ] **Web UI** - React frontend for visualization
- [ ] **S3 Integration** - Store outputs in cloud storage
- [ ] **Prometheus Metrics** - Export metrics for monitoring
- [ ] **Rate Limiting** - API request throttling
- [ ] **Audit Logging** - Track all API actions

---

## 🎓 Learning Outcomes

This project demonstrates proficiency in:

### Backend Development
- RESTful API design principles
- HTTP methods, status codes, and error handling
- Request validation and input sanitization
- Middleware patterns (authentication, logging)

### Database & SQL
- Relational database design and normalization
- Complex SQL queries with JOINs and aggregations
- Transactions and ACID properties
- Database indexing for performance

### Authentication & Security
- JWT token-based authentication
- Password hashing with bcrypt
- Authorization and access control
- Secure secret management

### Distributed Systems
- Job scheduling algorithms
- Resource allocation and matching
- Fair-share scheduling
- Concurrent programming with goroutines

### Infrastructure & DevOps
- Docker containerization
- Environment-based configuration
- Graceful shutdown handling
- Logging and monitoring

### Software Engineering
- Project organization and modularity
- Error handling patterns
- Testing strategies
- Version control with Git

---

## 🤝 Contributing

This is a portfolio/learning project, but feedback and suggestions are welcome!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 👤 Author

**Samik Kundu**

- 🎓 University of Wisconsin-Madison - Computer Science & Data Science
- 💼 Infrastructure Engineer Intern @ Ripple Labs (Summer 2025)
- 🔗 LinkedIn: [samik-kundu](https://linkedin.com/in/samik-kundu-862753338/)
- 📧 Email: skundu2448@gmail.com
- 🐙 GitHub: [@samik-k21](https://github.com/YOUR_USERNAME)

---

## 🙏 Acknowledgments

- **Inspiration:** Enterprise job schedulers like Slurm, PBS Pro, and Kubernetes
- **Learning Resources:** Go documentation, PostgreSQL docs, and various software engineering blogs
- **Purpose:** Built during winter break 2025 as a hands-on learning project to deepen understanding of APIs, distributed systems, and infrastructure software

---

## 📞 Support & Questions

If you're a recruiter or developer interested in this project:
- **Issues:** Open an issue on GitHub
- **Email:** skundu2448@gmail.com
- **LinkedIn:** Feel free to connect and message me

---

**⭐ If you find this project interesting, please consider starring it on GitHub!**
