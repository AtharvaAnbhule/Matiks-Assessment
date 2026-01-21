# Frontend Architecture Documentation

## System Design

### High-Level Flow

```
┌──────────────────────────────────────────┐
│        React Native UI Components         │
│  (LeaderboardScreen, SearchScreen, etc)  │
└───────────────────┬──────────────────────┘
                    │
        ┌───────────▼──────────┐
        │  Custom React Hooks   │
        │  (useLeaderboard,     │
        │   useSearch,          │
        │   useUserRank)        │
        └───────────┬──────────┘
                    │
        ┌───────────▼──────────┐
        │   API Service Layer   │
        │   (axios instance)    │
        └───────────┬──────────┘
                    │
              ┌─────▼─────┐
              │  Backend   │
              │  (Golang)  │
              └───────────┘
```

## Component Architecture

### 1. Navigation (`src/navigation/RootNavigator.tsx`)

**Structure:**

```
RootNavigator (Root)
├── SplashScreen
└── TabNavigator (Bottom Tab Navigation)
    ├── LeaderboardScreen
    ├── SearchScreen
    └── ProfileScreen
```

**Features:**

- Bottom tab navigation for easy access
- Splash screen for API health check
- Navigation state persistence

### 2. Screens

#### LeaderboardScreen.tsx

**Purpose:** Display paginated global leaderboard

**Features:**

- Pull-to-refresh
- Pagination (prev/next)
- Highlight top 10 players
- Loading states
- Error handling

**Data Flow:**

```
Component Mount
    ↓
Call useLeaderboard hook
    ↓
Fetch paginated data from API
    ↓
Render with pagination controls
    ↓
User pulls refresh → Call refresh()
    ↓
User clicks next → Fetch next page
```

#### SearchScreen.tsx

**Purpose:** Search for players by username and view rank

**Features:**

- Debounced search (500ms)
- Real-time rank updates (5s polling)
- Case-insensitive search
- Loading/error states
- Live indicator while polling

**Data Flow:**

```
User types username
    ↓
Debounce for 500ms
    ↓
Call API search
    ↓
Render search result
    ↓
Start polling for rank updates (5s intervals)
    ↓
Update rank when new data arrives
```

#### ProfileScreen.tsx

**Purpose:** Manage user profile and rating updates

**Features:**

- Create new user
- Load existing user
- View profile details
- Update rating
- Input validation

**Data Flow:**

```
User fills form
    ↓
Validate inputs
    ↓
Call API to create/load user
    ↓
Show user details
    ↓
User updates rating
    ↓
Call update API
    ↓
Refresh user details
```

### 3. API Service Layer (`src/services/api.ts`)

**Architecture:**

```
React Components
        ↓
useAPI Hooks
        ↓
API Service (Axios wrapper)
        ↓
Axios Interceptors
        ├─ Request logging
        └─ Response error handling
        ↓
Backend REST API
```

**Key Features:**

1. **Axios Configuration**
   - Base URL from environment
   - 10s timeout
   - Content-Type: application/json

2. **Interceptors**
   - Request logging (development only)
   - Error logging and formatting

3. **API Methods**
   - `userAPI.createUser()` - Create new user
   - `userAPI.getUser()` - Get user with rank
   - `userAPI.updateRating()` - Update rating
   - `userAPI.searchUser()` - Search by username
   - `userAPI.getLeaderboard()` - Get paginated leaderboard
   - `userAPI.getLeaderboardAroundUser()` - Context around user
   - `userAPI.checkHealth()` - Health check

### 4. Custom Hooks (`src/hooks/useAPI.ts`)

**Hook Pattern:**

```
Custom Hook
    ↓
Manage state (data, loading, error)
    ↓
Provide methods (fetch, refresh, etc)
    ↓
Handle side effects (polling, debouncing)
    ↓
Return to component
```

#### useLeaderboard Hook

**State:**

```typescript
{
  data: LeaderboardResponse | null,
  loading: boolean,
  error: string | null,
  currentPage: number
}
```

**Methods:**

- `fetchLeaderboard(page)` - Fetch specific page
- `nextPage()` - Go to next page
- `prevPage()` - Go to previous page
- `refresh()` - Reload current page

**Usage:**

```typescript
const { data, loading, error, nextPage, prevPage, refresh } = useLeaderboard();
```

#### useSearch Hook

**State:**

```typescript
{
  query: string,
  result: SearchResult | null,
  loading: boolean,
  error: string | null
}
```

**Methods:**

- `debouncedSearch(query)` - Search with debounce
- `clearSearch()` - Clear search state

**Debounce Logic:**

```
User types → Clear previous timer → Set 500ms timer → Call API
```

**Usage:**

```typescript
const { result, loading, debouncedSearch } = useSearch();
```

#### useUserRank Hook

**State:**

```typescript
{
  user: User | null,
  loading: boolean,
  error: string | null,
  refreshing: boolean
}
```

**Features:**

- Initial fetch on mount
- Automatic polling every 5 seconds
- Cleanup on unmount

**Polling Implementation:**

```typescript
useEffect(() => {
  const timer = setInterval(() => {
    fetchUser(); // Refresh every 5 seconds
  }, 5000);

  return () => clearInterval(timer); // Cleanup on unmount
}, [userId, fetchUser]);
```

**Usage:**

```typescript
const { user, loading, refreshing, refresh } = useUserRank(userId);
```

## State Management Strategy

**No Redux/MobX - Local Component State**

Why?

- Simple requirements (no complex state)
- Custom hooks sufficient
- Reduces boilerplate
- Easier to understand data flow

**Data Flow:**

```
Component → Hook → API Service → Backend → Hook → Component
```

## Performance Optimizations

### 1. Debounced Search

```typescript
const debouncedSearch = useCallback((query) => {
  clearTimeout(timer); // Clear previous
  timer = setTimeout(() => search(query), 500); // Wait 500ms
}, []);
```

**Benefits:**

- Reduces API calls from N to ~1-2 per user action
- Better UX (no jank while typing)
- Server load reduction

### 2. Pagination

**Why not infinite scroll?**

- Simpler to implement
- Better memory usage
- Clear page boundaries
- User knows how many total pages

**Page Size:**

- Default: 100 users per page
- Can fetch 50-1000 users

### 3. Memoization

```typescript
const fetchLeaderboard = useCallback(async (page) => {
  // Only recreated if dependencies change
}, []);
```

### 4. Component Memoization

```typescript
const LeaderboardEntry = React.memo(({ item }) => {
  // Only re-renders if item props change
});
```

## Error Handling Strategy

**API Errors:**

```typescript
try {
  const result = await userAPI.getUser(userId);
  setUser(result);
} catch (err) {
  setError(err?.response?.data?.message || "Unknown error");
  // Show error UI to user
}
```

**Fallback UI:**

- Show loading spinner
- Show error message
- Provide retry button
- Graceful degradation

## Real-Time Updates

### Polling Strategy (Current)

```typescript
useEffect(() => {
  // Poll every 5 seconds
  const interval = setInterval(fetchUser, 5000);
  return () => clearInterval(interval);
}, [fetchUser]);
```

**Pros:**

- Simple to implement
- Works with REST API
- No WebSocket needed

**Cons:**

- Latency (up to 5 seconds)
- Wasted requests

### Future: WebSocket Support

```typescript
// Pseudo-code for future improvement
useEffect(() => {
  const ws = new WebSocket("ws://...");
  ws.onmessage = (event) => {
    const update = JSON.parse(event.data);
    setUser(update);
  };
  return () => ws.close();
}, []);
```

## UI/UX Design Decisions

### 1. Color Scheme

```typescript
const Colors = {
  PRIMARY: "#007AFF", // Apple blue
  SUCCESS: "#4CAF50", // Green
  WARNING: "#FFD700", // Gold
  ERROR: "#D32F2F", // Red
  BACKGROUND: "#F5F5F5", // Light gray
  SURFACE: "#FFFFFF", // White
};
```

### 2. Responsive Layout

- Flexible layouts with `flex: 1`
- Percentage-based widths
- Device-agnostic typography

### 3. Accessibility

- Semantic text labels
- Touch-friendly button sizes (44+ points)
- Color contrast ratios (AA standard)
- Screen reader support (from React Native)

## Navigation Flow

```
Splash Screen (2s)
    ↓
Main Navigation (Bottom Tabs)
    ├─ Leaderboard Tab
    │  └─ View top 100 players, paginate
    ├─ Search Tab
    │  └─ Find player, see rank, live updates
    └─ Profile Tab
       └─ Create/manage user, update rating
```

## Testing Strategy

### Component Testing (Jest + React Native Testing Library)

```typescript
describe('LeaderboardScreen', () => {
  it('displays leaderboard entries', async () => {
    const { getByText } = render(<LeaderboardScreen />);
    await waitFor(() => {
      expect(getByText('john_doe')).toBeTruthy();
    });
  });
});
```

### Hook Testing (Jest)

```typescript
describe("useLeaderboard", () => {
  it("fetches leaderboard on mount", async () => {
    const { result } = renderHook(() => useLeaderboard());
    await waitFor(() => {
      expect(result.current.data).toBeDefined();
    });
  });
});
```

### Integration Testing

- Mock API responses
- Test navigation flows
- Test error scenarios

## Deployment

### Build for iOS

```bash
eas build --platform ios
```

### Build for Android

```bash
eas build --platform android
```

### Release to App Stores

- Register with Apple App Store
- Register with Google Play Store
- Create app listings
- Submit builds for review

## Performance Benchmarks

### Target Response Times

- Search: < 200ms (with debounce)
- Leaderboard: < 300ms
- Profile: < 200ms
- UI update: < 60ms (smooth 60 FPS)

### Memory Benchmarks

- App memory: < 100MB
- Leaderboard page: < 10MB (100 entries)

## Troubleshooting

### Common Issues

1. **API Connection Fails**
   - Check backend is running
   - Verify API base URL in `app.json`
   - Check network connectivity

2. **Search Returns Empty**
   - Username might not exist
   - Server rate limiting (wait a few seconds)
   - Check console logs for errors

3. **Slow Performance**
   - Check network latency (DevTools)
   - Verify backend caching is working
   - Check for memory leaks (Profile tool)
