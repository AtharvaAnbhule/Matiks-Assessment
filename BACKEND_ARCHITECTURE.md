# Backend Architecture Documentation

## System Design

### High-Level Flow

```
HTTP Request
    ↓
[Middleware] - CORS, Rate Limiting, Logging
    ↓
[Controller] - Parse request, call service
    ↓
[Service] - Business logic, caching logic
    ↓
[Repository] - Database queries
    ↓
[Cache] - Redis caching layer
    ↓
[Database] - PostgreSQL
```

## Component Details

### 1. Controller Layer (`controller/user_controller.go`)

**Responsibilities:**

- Parse HTTP requests
- Validate request format
- Call service methods
- Return HTTP responses
- Handle errors gracefully

**Key Methods:**

- `CreateUser()`: POST /users
- `GetUser()`: GET /users/:user_id
- `UpdateRating()`: PUT /users/:user_id/rating
- `SearchUser()`: GET /users/search?username=query
- `GetLeaderboard()`: GET /leaderboard
- `Health()`: GET /health

**Error Handling:**

- 400: Bad request (validation errors)
- 404: Not found
- 429: Rate limit exceeded
- 500: Server error

### 2. Service Layer (`service/user_service.go`)

**Responsibilities:**

- Implement business logic
- Cache management (cache-aside pattern)
- Input validation
- Rank calculations
- Non-blocking operations

**Caching Strategy:**

```
GetUser Request
    ↓
Check Cache
    ├─ Hit → Return cached user
    └─ Miss → Query DB
              ↓
         Cache result (async, fire-and-forget)
              ↓
         Return user
```

**Rank Calculation:**

```
Get User Rank
    ↓
Check Cache
    ├─ Hit → Return cached rank
    └─ Miss → Acquire per-user lock
             ↓
        Double-check cache (prevent thundering herd)
             ↓
        Query: COUNT(*) WHERE rating > user_rating
             ↓
        Cache result (3 min TTL)
             ↓
        Return rank
```

**Thread Safety:**

- Per-user mutex locks for rank calculation
- Prevents concurrent rank calculations for same user
- Minimal lock contention (lock only for rank, not entire user)

### 3. Repository Layer (`repository/user_repository.go`)

**Responsibilities:**

- Database operations
- Query optimization
- Index utilization

**Key Operations:**

| Operation         | Time Complexity | Index Used                |
| ----------------- | --------------- | ------------------------- |
| CreateUser        | O(1)            | Primary Key               |
| GetUserByID       | O(1)            | Primary Key               |
| GetUserByUsername | O(log n)        | idx_users_username_lower  |
| CalculateRank     | O(log n)        | idx_users_rating          |
| GetLeaderboard    | O(k log n)      | idx_users_rating_username |
| SearchUser        | O(log n + m)    | idx_users_username_lower  |

### 4. Cache Layer (`cache/cache.go`)

**Caching Pattern:** Cache-Aside

```
Pseudocode:
  value = cache.get(key)
  if value is null:
    value = fetch_from_database(key)
    cache.set(key, value)
  return value
```

**Cache Keys:**

- `user:{user_id}` - User object (5 min TTL)
- `rank:{user_id}` - User rank (3 min TTL)
- `leaderboard` - Leaderboard data (2 min TTL)

**Benefits:**

- Simple to implement
- Avoids cache inconsistency
- Transparent to business logic

**Invalidation Strategy:**

- TTL-based (automatic expiry)
- Event-based (on updates)
- Manual (fire-and-forget async invalidation)

### 5. Database Layer (`database/database.go`)

**Schema:**

```sql
CREATE TABLE users (
  id VARCHAR(255) PRIMARY KEY,
  username VARCHAR(255) NOT NULL UNIQUE,
  rating INT32 NOT NULL DEFAULT 1000,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_rating ON users(rating DESC);
CREATE INDEX idx_users_rating_username
  ON users(rating DESC, username);
CREATE INDEX idx_users_username_lower
  ON users(LOWER(username));
```

**Why These Indexes?**

1. `idx_users_rating`: Fast range queries for ranking
2. `idx_users_rating_username`: Composite index for sorting
3. `idx_users_username_lower`: Fast case-insensitive search

## Performance Optimization

### 1. Query Optimization

**Leaderboard Query:**

```sql
-- ✅ GOOD: Uses composite index, pre-sorted
SELECT id, username, rating
FROM users
ORDER BY rating DESC, username ASC
LIMIT 100 OFFSET 0;

-- ❌ BAD: No index, requires full table scan + sort
SELECT * FROM users
ORDER BY rating DESC
LIMIT 100;
```

### 2. Caching Impact

**Without Cache:**

- 10,000 users
- 100 requests/sec
- Response time: 50-100ms
- DB load: 100 QPS

**With Cache (85% hit ratio):**

- Same 10,000 users
- Same 100 requests/sec
- Response time: <10ms (cached)
- DB load: 15 QPS (15% miss rate)

### 3. Rank Calculation Optimization

**Inefficient:** Fetch all users, sort in memory

```go
// ❌ This loads all users into memory
users, _ := repo.GetAllUsers()
rank := 0
for _, u := range users {
  if u.Rating > targetRating {
    rank++
  }
}
```

**Efficient:** Single COUNT query

```go
// ✅ This uses index, returns single row
rank, _ := repo.CalculateRank(userID)
```

## Scalability Considerations

### Current Approach (10K - 100K users)

- **Pagination**: Offset-based (simple, works for small datasets)
- **Caching**: In-memory Redis (single instance)
- **Database**: Single PostgreSQL instance

### For 1M+ Users

1. **Replace Offset Pagination with Keyset Pagination**

```go
// Current: Offset-based (slow for large offsets)
SELECT * FROM users
ORDER BY rating DESC
LIMIT 100 OFFSET 1000000; // Scans 1M rows!

// Better: Keyset pagination (uses index)
SELECT * FROM users
WHERE rating < 4500 OR (rating = 4500 AND username > 'last_username')
ORDER BY rating DESC, username ASC
LIMIT 100;
```

2. **Distributed Caching**
   - Use Redis Cluster instead of single instance
   - Consistent hashing for key distribution

3. **Read Replicas**
   - Separate read and write operations
   - Read leaderboard from replica
   - Write/update to primary

4. **Horizontal Scaling**
   - Load balance multiple backend instances
   - Shared database
   - Shared Redis cache

## Concurrency & Thread Safety

### Race Condition Example

**Without Locks (UNSAFE):**

```go
// Two goroutines trying to calculate same user's rank
func GetUserRank(userID string) {
  cachedRank, _ := cache.GetRank(userID)
  if cachedRank == 0 {
    // Two goroutines reach here simultaneously
    rank := db.CalculateRank(userID) // DB called twice!
    cache.SetRank(userID, rank)
  }
  return rank
}
```

**With Per-User Lock (SAFE):**

```go
// Each user gets a separate mutex
rankMu := getRankMutex(userID) // Get or create mutex for user
rankMu.Lock()
defer rankMu.Unlock()

// Only one goroutine executes this critical section
cachedRank, _ := cache.GetRank(userID)
if cachedRank == 0 {
  rank := db.CalculateRank(userID)
  cache.SetRank(userID, rank)
}
```

## Non-Blocking Operations

### Problem

```
API Request → Update DB → Invalidate Cache → Return Response
                              (slow)
```

Cache invalidation can take 5-10ms, blocking API response.

### Solution: Fire-and-Forget

```go
// Invalidate cache asynchronously
go func() {
  ctx := context.Background()
  cache.InvalidateRank(ctx, userID) // Happens async
}()

// Return immediately
return userDTO, nil
```

**Benefits:**

- API response returns in <100ms
- Cache invalidation happens in background
- Short cache TTL (3-5 min) ensures eventual consistency

## Error Handling

### Database Errors

```go
if err := db.UpdateUser(ctx, user); err != nil {
  logger.Error("Failed to update user", zap.Error(err))
  return nil, err
}
```

### Cache Errors (Non-Critical)

```go
if err := cache.SetUser(ctx, user); err != nil {
  logger.Warn("Failed to cache user", zap.Error(err)) // Log but continue
  // Data is still returned, cache will be warmed on next read
}
```

## Monitoring & Observability

### Key Metrics

1. **Performance**
   - API response times
   - Cache hit ratio
   - Database query latency

2. **Availability**
   - Error rates
   - Service uptime
   - Database connectivity

3. **Scalability**
   - Concurrent users
   - Requests per second
   - Database connections

### Logging

```go
logger.Info("User search",
  zap.String("search_term", username),
  zap.String("found_user", user.Username),
  zap.Duration("latency", duration),
)
```

## Testing Strategy

### Unit Tests

```bash
go test ./service -v
go test ./repository -v
```

### Integration Tests

```bash
# Start Docker containers
docker-compose up

# Run integration tests
go test ./... -integration
```

### Load Tests

```bash
# Apache Bench
ab -n 10000 -c 100 http://localhost:8080/leaderboard

# Expected: 9000+ requests/sec with caching
```

## Security Considerations

1. **Input Validation**
   - Username length and format
   - Rating range (100-5000)
   - SQL injection prevention

2. **Rate Limiting**
   - Token bucket algorithm
   - 100 req/sec per IP
   - Burst capacity: 200 requests

3. **Error Messages**
   - No stack traces in production
   - Generic error messages
   - Detailed logs server-side only

4. **CORS**
   - Restrict to frontend origin
   - No credentials allowed for public API

## Deployment Checklist

- [ ] Database indexes created
- [ ] Redis configured with persistence
- [ ] Environment variables set
- [ ] Rate limiting configured
- [ ] Logging level set to INFO (production)
- [ ] Health check endpoint tested
- [ ] Load testing completed
- [ ] Error monitoring setup
- [ ] Backup strategy implemented
