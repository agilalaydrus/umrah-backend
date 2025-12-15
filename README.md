# 🕋 UmrahConnect Backend API

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Fiber Framework](https://img.shields.io/badge/Fiber-v2-black?style=flat&logo=go)](https://gofiber.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)](https://redis.io/)

**UmrahConnect** is a robust, high-performance backend service designed to manage Umrah travel groups. It features real-time pilgrim tracking, secure group chats, commerce management, and attendance systems using a clean architecture approach.

---

## 🚀 Key Features

### 🛡️ Security & Auth
* **Role-Based Access Control (RBAC):** Strict separation between `ADMIN`, `MUTAWWIF` (Tour Leader), and `JAMAAH` (Pilgrim).
* **Single Session Policy:** Users are automatically logged out from old devices when logging into a new one (powered by Redis).
* **Secure Registration:** Privilege escalation protection (default role assignment).

### 📡 Real-Time Capabilities (WebSocket)
* **Live Pilgrim Tracking:** Real-time location streaming for group monitoring.
* **Group Chat:** Persistent chat history with real-time message broadcasting.

### 🛒 Commerce & Booking
* **Atomic Booking System:** Thread-safe quota management using atomic database updates to prevent overselling.
* **Order Management:** Product catalog, order creation, and payment proof verification.
* **Secure Uploads:** File validation for payment proofs (MIME type & size checks).

### 📋 Core Management
* **Group Management:** Join via unique codes.
* **Itinerary & Attendance:** QR-based attendance scanning for itinerary events.
* **Manasik Guide:** Digital content management for prayers and guides.

---

## 🛠️ Tech Stack

* **Language:** Go (Golang)
* **Framework:** Fiber v2 (Fast HTTP)
* **Database:** PostgreSQL (via GORM)
* **Caching/PubSub:** Redis
* **Authentication:** JWT (JSON Web Tokens) with Redis Session
* **Architecture:** Clean Architecture (Handler -> Service -> Repository)

---

## 📂 Project Structure

```bash
.
├── cmd/
│   └── api/
│       └── main.go       # Application Entry Point
├── internal/
│   ├── entity/           # Database Models & DTOs
│   ├── handler/          # HTTP & WebSocket Handlers
│   ├── middleware/       # JWT, RBAC, & Session Middleware
│   ├── repository/       # Database Logic (GORM)
│   └── service/          # Business Logic
├── pkg/
│   └── database/         # DB & Redis Connection Wrappers
├── uploads/              # Storage for payment proofs
├── .env                  # Environment Variables
├── docker-compose.yml    # Docker Setup
└── go.mod
````

-----

## ⚙️ Installation & Setup

Follow these steps to set up the project locally.

### 1\. Prerequisites

  * **Go** (version 1.20 or higher)
  * **Docker & Docker Compose** (Recommended for DB)
  * *Alternatively:* Manual installation of PostgreSQL & Redis

### 2\. Clone the Repository

```bash
git clone [https://github.com/your-username/umrah-connect-backend.git](https://github.com/your-username/umrah-connect-backend.git)
cd umrah-connect-backend
```

### 3\. Setup Environment Variables

Create a file named `.env` in the root directory and copy the following configuration:

```env
# APP SETTINGS
APP_PORT=3000
APP_ENV=development

# DATABASE
# Note: Use 'localhost' if running Go locally. Use 'postgres' if running Go inside Docker.
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=umrah_db

# REDIS
# Note: Use 'localhost' if running Go locally. Use 'redis' if running Go inside Docker.
REDIS_HOST=localhost
REDIS_PORT=6379

# SECURITY
JWT_SECRET=your_super_secret_key_change_this
```

### 4\. Run Infrastructure (Database & Redis)

Use Docker Compose to spin up the required services instantly:

```bash
docker-compose up -d
```

*This command will pull and start PostgreSQL and Redis containers in the background.*

### 5\. Run the Application

Install dependencies and start the server:

```bash
# Download Go modules
go mod tidy

# Run the server
go run cmd/api/main.go
```

The server will start at `http://localhost:3000`.

-----

## 🔌 API Endpoints Overview

### 🔓 Public

  * `POST /api/register` - Register new Jamaah
  * `POST /api/login` - Login & Get Token
  * `GET  /api/packages` - View Travel Packages

### 🔒 Protected (User/Jamaah)

  * `POST /api/groups/join` - Join a group via code
  * `GET  /api/groups/:id/members` - List group members
  * `GET  /api/orders/my` - View purchase history
  * `POST /api/attendance/scan` - Scan QR for attendance
  * **WebSocket:** `ws://localhost:3000/ws/tracking/:group_id?token=JWT`
  * **WebSocket:** `ws://localhost:3000/ws/chat/:group_id?token=JWT`

### 🛡️ Admin / Mutawwif Only

  * `POST /api/admin/packages` - Create new Travel Package
  * `POST /api/admin/products` - Create Commerce Product
  * `POST /api/admin/groups` - Create new Group
  * `PATCH /api/admin/orders/:id/verify` - Verify Payment Proof

-----

## 🤝 Contributing

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the Branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request

-----

## 📝 License

Distributed under the MIT License. See `LICENSE` for more information.
