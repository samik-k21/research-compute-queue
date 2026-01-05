-- Drop existing tables if they exist (for clean setup)
DROP TABLE IF EXISTS usage_logs CASCADE;
DROP TABLE IF EXISTS job_dependencies CASCADE;
DROP TABLE IF EXISTS jobs CASCADE;
DROP TABLE IF EXISTS workers CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS groups CASCADE;

-- Research groups table
CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    cpu_quota INTEGER DEFAULT 100,  -- CPU hours per month
    priority INTEGER DEFAULT 1,      -- Higher = more important
    created_at TIMESTAMP DEFAULT NOW()
);

-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    group_id INTEGER REFERENCES groups(id),
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Jobs table
CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    group_id INTEGER REFERENCES groups(id) NOT NULL,
    
    -- Job specification
    script TEXT NOT NULL,
    cpu_cores INTEGER NOT NULL,
    memory_gb INTEGER NOT NULL,
    gpu_count INTEGER DEFAULT 0,
    estimated_hours DECIMAL,
    
    -- Status tracking
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending, running, completed, failed, cancelled
    priority INTEGER DEFAULT 1,
    
    -- Timing
    submitted_at TIMESTAMP DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Results
    exit_code INTEGER,
    output_path TEXT,
    error_message TEXT,
    
    -- Worker assignment
    worker_id INTEGER,
    
    CONSTRAINT valid_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled'))
);

-- Job dependencies (for DAG execution)
CREATE TABLE job_dependencies (
    job_id INTEGER REFERENCES jobs(id) ON DELETE CASCADE,
    depends_on_job_id INTEGER REFERENCES jobs(id) ON DELETE CASCADE,
    PRIMARY KEY (job_id, depends_on_job_id),
    CONSTRAINT no_self_dependency CHECK (job_id != depends_on_job_id)
);

-- Worker nodes
CREATE TABLE workers (
    id SERIAL PRIMARY KEY,
    hostname VARCHAR(255) NOT NULL UNIQUE,
    cpu_cores INTEGER NOT NULL,
    memory_gb INTEGER NOT NULL,
    gpu_count INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'idle',  -- idle, busy, offline
    last_heartbeat TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT valid_worker_status CHECK (status IN ('idle', 'busy', 'offline'))
);

-- Usage tracking (for fair-share calculations)
CREATE TABLE usage_logs (
    id SERIAL PRIMARY KEY,
    group_id INTEGER REFERENCES groups(id),
    job_id INTEGER REFERENCES jobs(id),
    cpu_hours_used DECIMAL NOT NULL,
    logged_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for common queries
CREATE INDEX idx_jobs_user_id ON jobs(user_id);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_group_id ON jobs(group_id);
CREATE INDEX idx_jobs_submitted_at ON jobs(submitted_at);
CREATE INDEX idx_usage_logs_group_id ON usage_logs(group_id);
CREATE INDEX idx_usage_logs_logged_at ON usage_logs(logged_at);

-- Insert some sample data for testing
INSERT INTO groups (name, cpu_quota, priority) VALUES 
    ('ML Research Lab', 500, 3),
    ('Systems Lab', 400, 2),
    ('Theory Group', 200, 1);

INSERT INTO workers (hostname, cpu_cores, memory_gb, gpu_count, status) VALUES
    ('compute-node-01', 32, 128, 2, 'idle'),
    ('compute-node-02', 16, 64, 1, 'idle'),
    ('compute-node-03', 64, 256, 4, 'idle');

-- Create a default admin user (password: admin123)
-- Password hash for 'admin123' using bcrypt
INSERT INTO users (email, password_hash, group_id, is_admin) VALUES
    ('admin@research.edu', '$2a$10$rN7qXqXqXqXqXqXqXqXqXuO7vKq9q9q9q9q9q9q9q9q9q9q9q9q9q', 1, TRUE);

COMMENT ON TABLE groups IS 'Research groups with resource quotas';
COMMENT ON TABLE users IS 'User accounts with authentication';
COMMENT ON TABLE jobs IS 'Submitted computing jobs';
COMMENT ON TABLE job_dependencies IS 'Job execution dependencies (DAG)';
COMMENT ON TABLE workers IS 'Available compute nodes';
COMMENT ON TABLE usage_logs IS 'Historical resource usage for fair-share';