# API Documentation

## Base URL

```
http://localhost:8080
```

## Authentication

Currently no authentication. Add API key or JWT in production.

## Response Format

### Success Response

```json
{
  "success": true,
  "data": {
    // Response data
  }
}
```

### Error Response

```json
{
  "error": "ERROR_CODE",
  "message": "Human-readable error message",
  "timestamp": "2026-01-20T10:30:45Z"
}
```

---

## Endpoints

### Health Check

**Endpoint**: `GET /health`

**Purpose**: Check service health

**Response**:

```json
{
  "status": "healthy",
  "timestamp": "2026-01-20T10:30:45Z"
}
```

**Status Codes**:

- `200 OK` - Service is healthy
- `503 Service Unavailable` - Service is unhealthy

**Example**:

```bash
curl http://localhost:8080/health
```

---

### Create User

**Endpoint**: `POST /users`

**Purpose**: Create a new user

**Request Body**:

```json
{
  "user_id": "usr_123",
  "username": "john_doe",
  "initial_rating": 1500
}
```

**Validation**:

- `user_id`: Required, string
- `username`: Required, 3-50 characters, alphanumeric + underscore/hyphen
- `initial_rating`: Required, integer 100-5000

**Response**:

```json
{
  "success": true,
  "data": {
    "id": "usr_123",
    "username": "john_doe",
    "rating": 1500
  }
}
```

**Status Codes**:

- `201 Created` - User created successfully
- `400 Bad Request` - Validation failed or user exists
- `429 Too Many Requests` - Rate limited

**Example**:

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "usr_123",
    "username": "john_doe",
    "initial_rating": 1500
  }'
```

---

### Get User

**Endpoint**: `GET /users/:user_id`

**Purpose**: Get user details with current rank

**Path Parameters**:

- `user_id`: User identifier (required)

**Response**:

```json
{
  "success": true,
  "data": {
    "id": "usr_123",
    "username": "john_doe",
    "rating": 1500,
    "rank": 42
  }
}
```

**Status Codes**:

- `200 OK` - User found
- `404 Not Found` - User doesn't exist
- `429 Too Many Requests` - Rate limited

**Example**:

```bash
curl http://localhost:8080/users/usr_123
```

---

### Update User Rating

**Endpoint**: `PUT /users/:user_id/rating`

**Purpose**: Update user's rating (non-blocking operation)

**Path Parameters**:

- `user_id`: User identifier (required)

**Request Body**:

```json
{
  "rating": 2000
}
```

**Validation**:

- `rating`: Required, integer 100-5000

**Response**:

```json
{
  "success": true,
  "data": {
    "id": "usr_123",
    "username": "john_doe",
    "rating": 2000,
    "rank": 38
  }
}
```

**Note**: Cache invalidation happens asynchronously (non-blocking)

**Status Codes**:

- `200 OK` - Rating updated
- `400 Bad Request` - Invalid rating or user not found
- `429 Too Many Requests` - Rate limited

**Example**:

```bash
curl -X PUT http://localhost:8080/users/usr_123/rating \
  -H "Content-Type: application/json" \
  -d '{"rating": 2000}'
```

---

### Search User

**Endpoint**: `GET /users/search?username=query`

**Purpose**: Search for user by username (case-insensitive)

**Query Parameters**:

- `username`: Username to search for (required, min 1 character)

**Response (Found)**:

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "usr_123",
      "username": "john_doe",
      "rating": 1500
    },
    "rank": 42,
    "found": true
  }
}
```

**Response (Not Found)**:

```json
{
  "success": true,
  "data": {
    "user": null,
    "rank": 0,
    "found": false
  }
}
```

**Status Codes**:

- `200 OK` - Search completed (regardless of result)
- `400 Bad Request` - Username parameter missing
- `429 Too Many Requests` - Rate limited

**Notes**:

- Search is case-insensitive
- Uses indexed database query
- Implements 500ms debounce on frontend

**Example**:

```bash
curl "http://localhost:8080/users/search?username=john_doe"

# URL encoding if needed
curl "http://localhost:8080/users/search?username=john%20doe"
```

---

### Get Leaderboard

**Endpoint**: `GET /leaderboard?page=1&page_size=100`

**Purpose**: Get paginated leaderboard

**Query Parameters**:

- `page`: Page number, 1-based (default: 1, optional)
- `page_size`: Items per page (default: 100, optional, max: 1000)

**Response**:

```json
{
  "success": true,
  "data": {
    "entries": [
      {
        "rank": 1,
        "username": "alice_pro",
        "rating": 5000
      },
      {
        "rank": 2,
        "username": "bob_player",
        "rating": 4500
      }
    ],
    "total": 10000,
    "page": 1,
    "page_size": 100,
    "has_more": true
  }
}
```

**Status Codes**:

- `200 OK` - Leaderboard retrieved
- `500 Internal Server Error` - Database error
- `429 Too Many Requests` - Rate limited

**Notes**:

- Results sorted by rating (DESC) then username (ASC)
- Implements tie-aware ranking
- Pagination uses offset-limit (simple, upgradeable to keyset)
- Uses composite index for performance

**Example**:

```bash
# Get first 100 players
curl "http://localhost:8080/leaderboard?page=1&page_size=100"

# Get next page
curl "http://localhost:8080/leaderboard?page=2&page_size=100"

# Different page size
curl "http://localhost:8080/leaderboard?page=1&page_size=50"
```

---

### Get Leaderboard Around User

**Endpoint**: `GET /users/:user_id/leaderboard-context?context_size=10`

**Purpose**: Get leaderboard entries around user's position

**Path Parameters**:

- `user_id`: User identifier (required)

**Query Parameters**:

- `context_size`: How many entries before/after (default: 10, optional, max: 100)

**Response**:

```json
{
  "success": true,
  "data": {
    "entries": [
      {
        "rank": 38,
        "username": "player_above_1",
        "rating": 1510
      },
      {
        "rank": 39,
        "username": "player_above_2",
        "rating": 1505
      },
      {
        "rank": 42,
        "username": "john_doe",
        "rating": 1500
      },
      {
        "rank": 43,
        "username": "player_below_1",
        "rating": 1490
      },
      {
        "rank": 44,
        "username": "player_below_2",
        "rating": 1485
      }
    ],
    "total": 10000,
    "page": 1,
    "page_size": 20,
    "has_more": true
  }
}
```

**Status Codes**:

- `200 OK` - Leaderboard context retrieved
- `404 Not Found` - User doesn't exist
- `429 Too Many Requests` - Rate limited

**Notes**:

- Shows users ranked before and after target user
- Useful for displaying ranking context in UI
- `context_size=10` shows ~20 entries (10 before, 10 after)

**Example**:

```bash
# Get context around user (default 10 before/after)
curl "http://localhost:8080/users/usr_123/leaderboard-context"

# Get custom context size (5 before/after = 10 entries)
curl "http://localhost:8080/users/usr_123/leaderboard-context?context_size=5"
```

---

## Error Codes

| Code              | Meaning                   | Action                |
| ----------------- | ------------------------- | --------------------- |
| `INVALID_REQUEST` | Request validation failed | Check request format  |
| `NOT_FOUND`       | Resource not found        | Verify user ID        |
| `CREATE_FAILED`   | User creation failed      | Check if user exists  |
| `UPDATE_FAILED`   | Update failed             | Verify input data     |
| `SEARCH_FAILED`   | Search failed             | Check username format |
| `FETCH_FAILED`    | Data fetch failed         | Retry request         |
| `RATE_LIMITED`    | Too many requests         | Wait before retrying  |

---

## Rate Limiting

**Policy**: Token bucket algorithm

**Limits**:

- 100 requests/second per IP
- Burst capacity: 200 requests
- Window: 1 second

**Response When Limited**:

```json
{
  "error": "RATE_LIMITED",
  "message": "Too many requests, please try again later"
}
```

**Status Code**: `429 Too Many Requests`

---

## Performance Guidelines

### Recommended Query Patterns

| Operation       | Expected Time | Notes                  |
| --------------- | ------------- | ---------------------- |
| Search user     | <100ms        | Uses index, cached     |
| Get rank        | <50ms         | Cached for 3 minutes   |
| Get leaderboard | <200ms        | Paginated, indexed     |
| Update rating   | <100ms        | Non-blocking           |
| Create user     | <200ms        | With index maintenance |

### Pagination Guidelines

```
Total users: 10,000
Recommended page_size: 50-100
Max page_size: 1000

For 10,000 users:
- Page size 100: 100 pages
- Page size 50: 200 pages
- Page size 1: 10,000 pages (don't do this)
```

### Caching

- **User data**: Cached for 5 minutes
- **Rank**: Cached for 3 minutes
- **Leaderboard**: Not cached (always fresh)

---

## Example Client Code

### JavaScript/TypeScript

```typescript
const BASE_URL = "http://localhost:8080";

// Create user
async function createUser(userId, username, rating) {
  const res = await fetch(`${BASE_URL}/users`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      user_id: userId,
      username,
      initial_rating: rating,
    }),
  });
  return res.json();
}

// Search user
async function searchUser(username) {
  const res = await fetch(
    `${BASE_URL}/users/search?username=${encodeURIComponent(username)}`,
  );
  return res.json();
}

// Get leaderboard
async function getLeaderboard(page = 1, pageSize = 100) {
  const res = await fetch(
    `${BASE_URL}/leaderboard?page=${page}&page_size=${pageSize}`,
  );
  return res.json();
}
```

### CURL

```bash
# Create user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"user_id":"u1","username":"john","initial_rating":1500}'

# Search
curl "http://localhost:8080/users/search?username=john"

# Get leaderboard
curl "http://localhost:8080/leaderboard?page=1&page_size=50"

# Update rating
curl -X PUT http://localhost:8080/users/u1/rating \
  -H "Content-Type: application/json" \
  -d '{"rating":2000}'
```

---

## Changelog

### v1.0.0 (Initial Release)

- ✅ User management (create, read, update)
- ✅ Leaderboard ranking with tie-awareness
- ✅ User search with case-insensitive matching
- ✅ Pagination support
- ✅ Real-time rank calculation
- ✅ Rate limiting
- ✅ Input validation
- ✅ Non-blocking updates

### Future Versions

- [ ] JWT authentication
- [ ] WebSocket for real-time updates
- [ ] Batch operations
- [ ] Analytics endpoints
- [ ] GraphQL API
