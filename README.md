# 🔗 Link Manager

A robust, backend-first research link manager with AI-ready storage and a premium dashboard.

## 🚀 Getting Started

### 1. Requirements
Ensure you have **Docker** and **Docker Compose** installed.

### 2. Launch the Application
Run the following command in the root directory:
```bash
docker compose up --build -d
```
The system will automatically:
1. Initialize the PostgreSQL database with `pgvector`.
2. Run database migrations.
3. Seed the admin user and sample data.
4. Launch the Web UI and API.

### 3. Access
- **Web UI**: [http://localhost:5177](http://localhost:5177)
- **API Health**: [http://localhost:5177/api/v1/healthz](http://localhost:5177/api/v1/healthz)

**Default Credentials:**
- **Username**: `admin`
- **Password**: `admin`

---

## 🛠 Management Commands

| Action | Command |
| :--- | :--- |
| **Start Services** | `docker compose up -d` |
| **Stop Services** | `docker compose down` |
| **Rebuild & Start** | `docker compose up --build -d` |
| **View Logs** | `docker compose logs -f` |
| **Check Health** | `docker compose ps` |
| **Wipe Data** | `docker compose down -v` |

---

## 📂 Project Structure
- **/cmd/api**: Entry point for the Go backend.
- **/internal**: Core business logic, database handlers, and middleware.
- **/migrations**: SQL schema and seed data.
- **/web**: Frontend assets (HTML, CSS, JS) and Nginx configuration.
- **/docs/specs**: Detailed technical specifications.
