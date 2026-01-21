# Neon DB Setup Guide

## What is Neon?

**Neon** is a serverless PostgreSQL database service. It's perfect for development and production:

- ✅ **Free tier**: Includes 5GB storage and instant provisioning
- ✅ **No credit card required**: Sign up and get running immediately
- ✅ **Scalable**: Handles 10K+ users easily
- ✅ **Secure**: Built-in SSL, automatic backups
- ✅ **Auto-suspend**: Pauses when idle to save resources
- ✅ **Compatible**: Standard PostgreSQL - no code changes needed

## Step-by-Step Setup (5 minutes)

### Step 1: Create Neon Account

1. Go to **https://console.neon.tech**
2. Click **"Sign Up"**
3. Choose sign-up method:
   - Email/Password
   - GitHub
   - Google
4. No credit card required! ✨

### Step 2: Create Your First Project

1. After signing up, you'll see the Neon Console
2. Click **"Create a project"**
3. Enter project name: `leaderboard` (or your preferred name)
4. Select region closest to you
5. Click **"Create project"**

Neon automatically:

- Creates a PostgreSQL database
- Generates a superuser account
- Provides connection string

### Step 3: Get Your Connection String

1. In Neon Console, find your project
2. Click on your project name
3. Go to the **"Connection string"** tab
4. Click the dropdown and select **"Go"**
5. Copy the full connection string

**Example connection string:**

```
postgresql://neon_user:xxxxx@ep-cool-moon-123.neon.tech:5432/neon?sslmode=require
```

Keep this safe! You'll need it in the next step.

### Step 4: Configure Your Backend

1. Navigate to the backend folder:

   ```bash
   cd backend
   ```

2. Create `.env` file with your connection string:

   ```bash
   @"
   DATABASE_URL=postgresql://neon_user:xxxxx@ep-cool-moon-123.neon.tech:5432/neon?sslmode=require
   REDIS_HOST=localhost
   REDIS_PORT=6379
   REDIS_PASSWORD=
   PORT=8080
   ENV=development
   "@ | Out-File -Encoding UTF8 .env
   ```

3. Replace the connection string with your actual one!

### Step 5: Run Backend

```bash
# Download dependencies
go mod download

# Run backend
go run main.go
```

**Expected output:**

```
Server starting address=:8080
Database connected
Cache connected
```

✅ **Done!** Your backend is now connected to Neon!

---

## Important Neon Features

### 1. Connection String Format

```
postgresql://username:password@endpoint:5432/database?sslmode=require
```

- `username` - Auto-created user (usually `neon_user`)
- `password` - Your password (set during account creation)
- `endpoint` - Your Neon endpoint (e.g., `ep-cool-moon-123.neon.tech`)
- `5432` - PostgreSQL port (always this)
- `database` - Database name (usually `neon`)
- `sslmode=require` - **Important**: Neon requires SSL connections

### 2. Multiple Connections

Your `.env` file can have different connection strings:

```bash
# Development (local development box)
DATABASE_URL=postgresql://user:pass@ep-xxxxx.neon.tech:5432/neon?sslmode=require

# Production (deployed on cloud)
DATABASE_URL=postgresql://prod_user:prod_pass@ep-yyyyy.neon.tech:5432/prod_db?sslmode=require
```

### 3. Auto-Suspend Feature

Neon automatically suspends inactive databases:

- **After**: 5 minutes of inactivity (free tier)
- **Resume**: On next connection (automatic, no manual action)
- **Data**: Never lost - all data persists

This is great for development - you don't pay for unused compute!

### 4. Password Management

To change your Neon database password:

1. Go to Neon Console
2. Select your project
3. Go to **"Connection"** tab
4. Click **"Reset password"**
5. Copy new password
6. Update `.env` with new connection string

### 5. Multiple Branches

Create separate databases for different features:

```bash
# In Neon Console: Click "Create a branch"

# Main database for production
DATABASE_URL=postgresql://...@ep-main.neon.tech:5432/neon?sslmode=require

# Staging database for testing
DATABASE_URL=postgresql://...@ep-staging.neon.tech:5432/neon?sslmode=require

# Development database for experiments
DATABASE_URL=postgresql://...@ep-dev.neon.tech:5432/neon?sslmode=require
```

---

## Troubleshooting

### Error: "Host is not allowed to connect"

**Cause**: Your IP address isn't whitelisted

**Solution**:

1. Go to Neon Console
2. Project Settings → **"IP Whitelist"**
3. Add your IP address
4. Or allow all IPs (for development only): `0.0.0.0/0`

### Error: "FATAL: invalid_authorization_specification"

**Cause**: Wrong password in connection string

**Solution**:

1. Go to Neon Console
2. Copy the connection string again
3. Make sure password is correct (case-sensitive!)
4. Update `.env`

### Error: "psql: could not translate host name"

**Cause**: Connection string has typo in endpoint

**Solution**:

1. Verify endpoint in connection string
2. Copy from Neon Console (don't type manually)
3. Make sure it starts with `ep-` (like `ep-cool-moon-123`)

### Error: "SSL connection error"

**Cause**: Missing `?sslmode=require` in connection string

**Solution**:

1. Check connection string ends with `?sslmode=require`
2. **Important**: Neon requires SSL - don't remove this!

### Backend Shows "Database connected" but Tables Missing

**Cause**: Tables are auto-created but might take a few seconds

**Solution**:

1. Wait 10 seconds
2. Restart backend: `go run main.go`
3. Check Neon Console to verify tables exist

---

## Advanced: Neon SQL Editor

You can query your database directly in Neon Console:

1. Go to **"SQL Editor"** tab
2. Write SQL:
   ```sql
   SELECT * FROM users LIMIT 10;
   SELECT COUNT(*) FROM users;
   SELECT username, rating, rank FROM users ORDER BY rating DESC;
   ```
3. Click **"Execute"**
4. See results instantly

## Advanced: Neon Web Console

Monitor your database in real-time:

1. Go to Neon Console
2. Click **"Monitoring"** tab
3. See:
   - Query performance
   - Connection count
   - Disk usage
   - Last activity

---

## Cost Breakdown

### Free Tier

- **Storage**: 5 GB
- **Compute**: 0.25 CPU hours/month
- **Perfect for**: Development, small projects
- **Cost**: **FREE** ✨

### Pro Tier (when you need more)

- **Pay-as-you-go**: Only pay for what you use
- **Unlimited storage**: As much as you need
- **Auto-scaling compute**: Scales with traffic
- **Starting**: ~$0.27/hour when active

### For This Project

- **Users**: 10K+ (easily fits in free tier)
- **Storage**: ~100 MB for test data
- **Estimated cost**: **FREE on Neon free tier** 🎉

---

## Docker Compose with Neon

Update `docker-compose.yml` to use your Neon database:

```yaml
services:
  backend:
    environment:
      DATABASE_URL: postgresql://user:pass@ep-xxxxx.neon.tech:5432/neon?sslmode=require
      REDIS_HOST: redis
      PORT: 8080
    depends_on:
      redis:
        condition: service_healthy
```

Then run:

```bash
docker-compose up
```

Neon database is in the cloud, so only Redis runs locally!

---

## Deployment with Neon

When deploying to production:

### 1. Create Production Database in Neon

1. In Neon Console, create a new branch
2. Get the production connection string
3. Name it `prod_db` or similar

### 2. Set Environment Variables

**On AWS Lambda, Vercel, Railway, etc:**

```bash
DATABASE_URL=postgresql://prod_user:prod_pass@ep-xxxxx.neon.tech:5432/prod_db?sslmode=require
REDIS_HOST=redis.example.com  # Your production Redis
PORT=8080
ENV=production
```

### 3. Deploy Backend

Backend automatically:

- Connects to Neon
- Creates tables if missing
- Starts serving requests

### 4. Test Production

```bash
# From your machine:
curl https://your-production-url.com/health

# Should return:
# {"status":"ok"}
```

---

## Next Steps

1. ✅ Set up Neon account and database
2. ✅ Copy connection string to `.env`
3. ✅ Run backend: `go run main.go`
4. ✅ Follow [QUICK_START.md](QUICK_START.md) for Redis setup
5. ✅ Run frontend
6. 🚀 You're ready to go!

---

## More Resources

- **Neon Console**: https://console.neon.tech
- **Neon Documentation**: https://neon.tech/docs
- **Neon API**: https://neon.tech/docs/api-reference
- **Contact Support**: https://neon.tech/support

---

## FAQ

**Q: Is my data safe in Neon?**
A: Yes! Neon uses enterprise-grade security with automatic backups, encryption at rest, and SSL for data in transit.

**Q: Can I migrate from local PostgreSQL to Neon?**
A: Yes! Use `pg_dump` and `pg_restore`:

```bash
pg_dump postgresql://localhost/leaderboard | psql postgresql://...@neon.tech/neon
```

**Q: What happens if I exceed free tier limits?**
A: You'll be notified before exceeding limits. You can upgrade anytime.

**Q: Can I use Neon for testing?**
A: Perfect for testing! Create a test branch in Neon Console for unit tests.

**Q: How do I delete a project?**
A: Go to **Project Settings** → **Delete project**. This deletes all data.

**Q: Can I have multiple projects?**
A: Yes! Create as many as you want in Neon Console.

---

**Last Updated**: January 20, 2026  
**Status**: Production Ready ✅
