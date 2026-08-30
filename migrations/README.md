# Database Migrations

This directory contains all database migration scripts for the Helia application.

## Migration Files

### 001_create_users_table.sql
Creates the `users` table with columns for:
- User authentication (id, username, email, password_hash)
- 2FA support (two_fa_enabled, two_fa_secret, backup_codes)
- Timestamps (created_at, updated_at)

Includes:
- Unique constraints on username and email
- Indexes for faster lookups
- Auto-updating timestamp trigger
- Column documentation

## Running Migrations

### Using psql (PostgreSQL CLI)

```bash
# Connect to your database
psql -U postgres -d helia -h localhost

# Run the migration file
\i migrations/001_create_users_table.sql

# Verify the table was created
\dt users
\d users
```

### Using Go (from application code)

You can add a migration runner in your Go code:

```bash
psql -U postgres -d helia -h localhost < migrations/001_create_users_table.sql
```

### Using Docker

If running PostgreSQL in Docker:

```bash
# Copy migration file to container and run it
docker cp migrations/001_create_users_table.sql postgres_container:/tmp/

docker exec -i postgres_container psql -U postgres -d helia < migrations/001_create_users_table.sql
```

## Rollback

To rollback (drop the users table):

```sql
DROP TRIGGER IF EXISTS users_update_timestamp ON users;
DROP FUNCTION IF EXISTS update_users_timestamp();
DROP TABLE IF EXISTS users;
```

## Migration Naming Convention

Follow this naming convention for new migrations:
```
NNN_description_of_migration.sql
```

Where NNN is a 3-digit sequential number (001, 002, 003, etc.)

## Best Practices

1. Always make migrations idempotent using `IF NOT EXISTS` or `IF EXISTS`
2. Never modify existing migrations - create new ones for changes
3. Test migrations in a development environment first
4. Include documentation comments in the SQL
5. Keep migrations focused and single-purpose
6. Always include both forward and rollback statements

## Future Migrations

When adding new features, create new migration files:

- `002_add_user_roles_table.sql` - For role-based access control
- `003_add_audit_log_table.sql` - For audit logging
- `004_add_user_sessions_table.sql` - For session management
- etc.
