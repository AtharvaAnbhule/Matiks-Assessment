# Quick Start Guide

## Prerequisites

### System Requirements

- **OS**: Windows, macOS, or Linux
- **RAM**: 4GB minimum (8GB recommended)
- **Disk**: 2GB free space

### Required Software

**Backend:**

- Go 1.21+: https://golang.org/dl
- Neon DB Account: https://console.neon.tech (free)
- Redis 6.0+: https://redis.io/download (or use cloud Redis)

**Frontend:**

- Node.js 16+: https://nodejs.org
- npm 8+: Comes with Node.js
- Expo CLI: `npm install -g expo-cli`

### Verify Installations

```bash
# Go
go version

# Node.js
node --version
npm --version

# Redis (if installed locally)
redis-cli --version
```

## Setting Up Neon DB (Required First)

### Step 1: Create Neon Account

1. Go to https://console.neon.tech
2. Sign up (free account, no credit card needed)
3. Create a new project

### Step 2: Get Connection String

1. In Neon Console, go to your project
2. Click on the database you created
3. Go to "Connection string" tab
4. Copy the full connection string (looks like: `postgresql://user:password@ep-xxxxx.neon.tech:5432/neon?sslmode=require`)
5. Save it somewhere safe - you'll need it in a moment

### Step 3: Set Up Environment

```bash
cd backend

# Create .env file with your Neon connection string
# Windows PowerShell:
@"
DATABASE_URL=postgresql://your-user:your-password@ep-your-project.neon.tech:5432/neon?sslmode=require
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
PORT=8080
ENV=development
"@ | Out-File -Encoding UTF8 .env
```

**Important**: Replace `postgresql://...` with your actual Neon connection string!

## 5-Minute Setup

### 1. Neon DB Setup (2 minutes)

- Create free account at https://console.neon.tech
- Create new project (auto-creates database)
- Copy connection string from "Connection string" tab

### 2. Backend Setup

```bash
cd backend

# Download dependencies
go mod download

# Create .env file with your Neon connection string
# Windows PowerShell:
@"
DATABASE_URL=postgresql://user:password@ep-xxxxx.neon.tech:5432/neon?sslmode=require
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
PORT=8080
ENV=development
"@ | Out-File -Encoding UTF8 .env

# Run backend
go run main.go
```

**Expected Output:**

```
Server starting address=:8080
Database connected
Cache connected
```

### 3. Start Redis (Terminal 2)

```bash
# Windows
redis-server

# macOS (if installed with Homebrew)
redis-server /usr/local/etc/redis.conf

# Docker
docker run -d -p 6379:6379 redis:7-alpine
```

### 4. Frontend Setup (Terminal 3)

```bash
cd frontend

# Install dependencies
npm install

# Start Expo
npm start
```

**Expected Output:**

```
Expo Dev Server running at http://localhost:19000
```

### 6. Run Frontend

Choose one:

```bash
# Web browser (fastest for testing)
w

# Android emulator
a

# iOS simulator (macOS only)
i
```

## Testing the System

### 1. Backend Health Check

```bash
curl http://localhost:8080/health

# Expected:
# {"status":"healthy","timestamp":"2026-01-20T10:30:45Z"}
```

### 2. Create User

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "username": "john_doe",
    "initial_rating": 1500
  }'

# Expected:
# {"success":true,"data":{"id":"user1","username":"john_doe","rating":1500}}
```

### 3. Get Leaderboard

```bash
curl http://localhost:8080/leaderboard?page=1&page_size=10

# Expected:
# Paginated leaderboard with 10 entries
```

### 4. Search User

```bash
curl http://localhost:8080/users/search?username=john_doe

# Expected:
# User details with rank
```

### 5. Update Rating

```bash
curl -X PUT http://localhost:8080/users/user1/rating \
  -H "Content-Type: application/json" \
  -d '{"rating": 2000}'

# Expected:
# Updated user with new rating
```

## Frontend Testing

### 1. Leaderboard Tab

- Should see top players
- Try pagination (next/prev)
- Pull to refresh

### 2. Search Tab

- Type username (e.g., "john_doe")
- Should find user and show rank
- Should poll and update rank in real-time

### 3. Profile Tab

- Click "Create User"
- Fill in user ID, username, rating
- Create user
- Update rating
- See rank update

## Troubleshooting

### Backend Won't Start

**Error: "Failed to connect to database"**

```bash
# Check your Neon connection string in .env
# Go to https://console.neon.tech
# Copy connection string from your project
# Example format:
DATABASE_URL=postgresql://user:password@ep-xxxxx.neon.tech:5432/neon?sslmode=require

# Test connection:
psql "postgresql://user:password@ep-xxxxx.neon.tech:5432/neon?sslmode=require" -c "SELECT 1"

# Expected output: 1 (number)

# If connection fails:
# 1. Verify you copied the correct connection string
# 2. Check the password is correct (case-sensitive!)
# 3. Ensure `sslmode=require` is in the string
# 4. Check your IP is not blocked in Neon firewall
```

**Error: "Failed to connect to redis"**

```bash
# Check Redis is running
redis-cli ping

# Expected output: PONG

# If not running, start Redis:
redis-server
```

### Frontend Won't Connect to Backend

**Error: "Search failed" or "API unavailable"**

1. **Check backend is running**

   ```bash
   curl http://localhost:8080/health
   ```

2. **Verify API URL in `frontend/app.json`**

   ```json
   {
     "extra": {
       "apiBaseUrl": "http://localhost:8080"
     }
   }
   ```

3. **For Android/iOS on physical device**
   - Use machine IP instead of localhost
   - `http://192.168.1.100:8080` (replace with your IP)
   - Get IP: `ipconfig` (Windows) or `ifconfig` (Mac/Linux)

4. **Clear cache**
   ```bash
   npm start -- -c
   ```

### Database Errors

**Error: "Database already exists"**
**Error: "Invalid connection string" or "Certificate verify failed"**

```bash
# Make sure sslmode=require is in your connection string
# Example correct:
DATABASE_URL=postgresql://user:password@ep-xxxxx.neon.tech:5432/neon?sslmode=require

# Also verify the connection string from Neon Console:
# 1. Go to https://console.neon.tech
# 2. Select your project
# 3. Go to "Connection string" tab
# 4. Click the dropdown and select "Go"
# 5. Copy the exact string shown
```

**Error: "Database does not exist"**

```bash
# Neon auto-creates a database when you create a project
# If not, create one in Neon Console:
# 1. Go to https://console.neon.tech
# 2. Click "Create database"
# 3. Name it "neon" or your preferred name
# 4. Copy the updated connection string
```

**Error: "Relations not created"**

```bash
# Tables are created automatically on backend startup
# Just ensure backend is running and connected to Neon
# Check logs for confirmation:
go run main.go  # Should show "Database connected"
```

## Load Testing

### Test with Apache Bench

```bash
# Test search endpoint
ab -n 1000 -c 10 http://localhost:8080/leaderboard

# Test with data
ab -n 1000 -c 10 -p data.json \
  -T application/json \
  http://localhost:8080/users/search?username=john
```

### Test with Hey (faster)

```bash
# Install: go install github.com/rakyll/hey@latest

hey -n 10000 -c 100 http://localhost:8080/leaderboard
```

### Performance Metrics to Check

```bash
# Expected response times:
# - Search: < 100ms (cached)
# - Leaderboard: < 200ms
# - Update: < 100ms

# Expected throughput:
# - 5000+ requests/sec (with caching)
```

## Create Test Data

### Using Curl

```bash
# Create 10 test users
for i in {1..10}; do
  curl -X POST http://localhost:8080/users \
    -H "Content-Type: application/json" \
    -d "{
      \"user_id\": \"user$i\",
      \"username\": \"player_$i\",
      \"initial_rating\": $((1000 + i * 100))
    }"
done
```

### Using PowerShell

```powershell
$baseUrl = "http://localhost:8080/users"

for ($i = 1; $i -le 10; $i++) {
    $body = @{
        user_id = "user$i"
        username = "player_$i"
        initial_rating = 1000 + ($i * 100)
    } | ConvertTo-Json

    Invoke-RestMethod -Uri $baseUrl -Method Post `
        -Body $body -ContentType "application/json"
}
```

## Development Workflow

### Making Code Changes

**Backend:**

```bash
# Install air for hot reload (optional)
go install github.com/cosmtrek/air@latest

# Run with hot reload
air

# Or manually restart
go run main.go
```

**Frontend:**

- Changes auto-reload in Expo
- Press 'r' to reload app
- Press 'R' to restart bundler

### Debugging

**Backend:**

```bash
# Print debug info
fmt.Printf("Debug: %+v\n", variable)

# Or use debugger (VS Code)
# Set breakpoints in IDE
```

**Frontend:**

```javascript
console.log("Debug:", variable);

// View in Expo dev tools
// Press 'j' to open debugger
```

## Next Steps

1. **Explore the code**
   - Read `BACKEND_ARCHITECTURE.md`
   - Read `FRONTEND_ARCHITECTURE.md`

2. **Customize**
   - Modify ratings range (currently 100-5000)
   - Change cache TTL values
   - Adjust pagination size

3. **Deploy**
   - Read deployment section in README.md
   - Build Docker container
   - Deploy to cloud (AWS, Azure, GCP)

4. **Scale**
   - Add more test data
   - Run load tests
   - Optimize based on profiling

## Useful Commands

```bash
# Backend
go test ./...              # Run tests
go build -o app .          # Build binary
go mod tidy               # Clean dependencies

# Frontend
npm test                  # Run tests
npm run ios              # Build for iOS
npm run android          # Build for Android
expo build:web           # Build for web

# Database (Neon)
psql "postgresql://user:password@ep-xxxxx.neon.tech:5432/neon?sslmode=require"
\dt                             # List tables
SELECT * FROM users;            # Query users

# Redis
redis-cli                 # Connect to Redis
KEYS *                    # List all keys
FLUSHDB                   # Clear database
```

## Performance Monitoring

### Backend Metrics

```bash
# Check database performance
psql "postgresql://..." -c "EXPLAIN ANALYZE SELECT * FROM users ORDER BY rating DESC LIMIT 100;"

# Check cache hit ratio
redis-cli INFO stats
```

### Frontend Metrics

- Open DevTools in browser
- Network tab: Check response times
- Performance tab: Check frame rate

````

### Change Cache TTL

**Backend** (`cache/cache.go`):
```go
const (
    CacheUserTTL = 5 * time.Minute     // Change here
    CacheRankTTL = 3 * time.Minute     // Or here
)
````

### Change Pagination Size

**Frontend** (`hooks/useAPI.ts`):

```typescript
const pageSize = 50; // Change this
```

## Support

For issues:

1. Check error messages carefully
2. Review troubleshooting section
3. Check backend logs
4. Check frontend console (Expo)
5. Review architecture docs

## Performance Targets Met

✅ Handle 10,000+ users  
✅ Sub-second search (with caching)  
✅ Real-time rank updates (5s polling)  
✅ Tie-aware ranking  
✅ Non-blocking updates  
✅ Input validation & sanitization  
✅ Rate limiting  
✅ Clean architecture with separation of concerns

Good luck! 🚀
