# PostgreSQL Migration Runner (PowerShell)
# Usage: .\run_migrations.ps1 -Username postgres -Database helia -Host localhost -Port 5432

param(
    [string]$Username = "postgres",
    [string]$Database = "helia",
    [string]$Host = "localhost",
    [int]$Port = 5432
)

Write-Host "Running migrations for database: $Database"
Write-Host "User: $Username"
Write-Host "Host: $Host`:$Port"
Write-Host ""

# Check if psql is installed
try {
    $psqlVersion = psql --version 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "psql not found"
    }
} catch {
    Write-Host "Error: psql (PostgreSQL client) is not installed or not in PATH."
    Write-Host "Please install PostgreSQL and add it to your system PATH."
    exit 1
}

# Get all migration files in order
$migrationFiles = Get-ChildItem -Path "migrations" -Filter "*.sql" -File | Sort-Object Name

if ($migrationFiles.Count -eq 0) {
    Write-Host "No migration files found in the migrations directory."
    exit 1
}

foreach ($migrationFile in $migrationFiles) {
    Write-Host "Running migration: $($migrationFile.FullName)"
    
    # Run the migration
    & psql -U $Username -d $Database -h $Host -p $Port -f $migrationFile.FullName
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ $($migrationFile.Name) completed successfully" -ForegroundColor Green
    } else {
        Write-Host "✗ $($migrationFile.Name) failed" -ForegroundColor Red
        exit 1
    }
    Write-Host ""
}

Write-Host "All migrations completed successfully!" -ForegroundColor Green
