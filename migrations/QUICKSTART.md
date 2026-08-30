# Quick Migration Commands

## Prerequisites

Make sure PostgreSQL is installed and you have:
- Username (default: `postgres`)
- Database name (default: `helia`)
- Host (default: `localhost`)
- Port (default: `5432`)

## Windows (PowerShell)

### Option 1: Using the PowerShell script (Recommended)

```powershell
cd d:\helia_app
.\migrations\run_migrations.ps1

# Or with custom parameters
.\migrations\run_migrations.ps1 -Username youruser -Database yourdb -Host localhost -Port 5432
```

### Option 2: Direct psql command

```powershell
psql -U postgres -d helia -h localhost -p 5432 -f migrations/001_create_users_table.sql
```

### Option 3: Using Command Prompt

```cmd
cd d:\helia_app
psql -U postgres -d helia -h localhost -p 5432 -f migrations/001_create_users_table.sql
```

## Linux/Mac

### Option 1: Using the bash script (Recommended)

```bash
cd /path/to/helia_app
chmod +x migrations/run_migrations.sh
./migrations/run_migrations.sh

# Or with custom parameters
./migrations/run_migrations.sh postgres helia localhost 5432
```

### Option 2: Direct psql command

```bash
psql -U postgres -d helia -h localhost -p 5432 -f migrations/001_create_users_table.sql
```

## Docker

### Option 1: Direct execution in running container

```bash
docker exec -it <container_name> psql -U postgres -d helia -f /migrations/001_create_users_table.sql
```

### Option 2: Copy and run

```bash
docker cp migrations/001_create_users_table.sql <container_name>:/tmp/
docker exec -it <container_name> psql -U postgres -d helia -f /tmp/001_create_users_table.sql
```

### Option 3: Using docker-compose (if available)

```bash
docker-compose exec postgres psql -U postgres -d helia -f /migrations/001_create_users_table.sql
```

## Verify Migration

After running the migration, verify the table was created:

```sql
-- Connect to your database
psql -U postgres -d helia -h localhost

-- List all tables
\dt

-- Describe the users table
\d users

-- Or with SQL
SELECT * FROM information_schema.tables WHERE table_name = 'users';

-- Check columns
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'users';
```

## Connection String Format

If you need the full connection string:

```
postgresql://username:password@host:port/database
```

Example:
```
postgresql://postgres:password@localhost:5432/helia
```

## Troubleshooting

### Connection refused
- Check if PostgreSQL is running
- Verify host and port are correct
- Check if database exists: `psql -l`

### Permission denied
- Check username and permissions
- Make sure user has CREATE TABLE privilege
- Try connecting as superuser

### Migration already exists error
- The migration is idempotent (uses `IF NOT EXISTS`)
- Safe to run multiple times
- Check for syntax errors in the SQL file

### Column already exists
- Check if the table was partially created
- Drop and recreate: `DROP TABLE IF EXISTS users CASCADE;`
- Then re-run the migration

## Additional Resources

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [psql Documentation](https://www.postgresql.org/docs/current/app-psql.html)
- [SQL Best Practices](https://www.postgresql.org/docs/current/sql.html)
