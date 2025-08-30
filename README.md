# 🌊 Hydrocast  
![status](https://img.shields.io/badge/status-work%20in%20progress-yellow)
One-stop solution for **fish pond monitoring** and **water quality forecasting**.

---

## 📦 Dependencies
- **Go** v1.25.0  
  - [Goose](https://github.com/pressly/goose) v3.25.0 – Pure Go migration tool  
  - [Protoc](https://grpc.io/docs/protoc-installation/) v28.3 – Protobuf generation tool  
- **PostgreSQL** 17.6  
- **OpenSSL** 3.2.1  

---

## ⚙️ Setup Instructions

### 1. Install dependencies
Make sure all dependencies listed above are installed.  

### 2. Environment setup
Copy the example environment file:  
```bash
cp .env.example .env
```

Update the `.env` file with the following:

- **Database password** (generate with OpenSSL):  
  ```bash
  openssl rand -hex 16
  ```

- **JWT secret** (HS256, base64):  
  ```bash
  openssl rand -base64 32
  ```

- **Admin password** (for first admin seeding):  
  ```bash
  openssl rand -hex 8
  ```

### 3. Run migrations
Apply database schemas using Goose:  
```bash
goose up
```

### 4. Seed admin user
Run the seeder to create the first admin:  
```bash
make seed
```

---

✅ You’re ready to run **Hydrocast**!
