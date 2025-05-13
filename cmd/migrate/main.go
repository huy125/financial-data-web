package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MigrationFile represents a migration file
type MigrationFile struct {
	Path     string
	Name     string
	Content  string
	Executed bool
}

// Migrate is the main function to handle the migration process
func Migrate(ctx context.Context, connString string) error {
	// Connect to database
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database failed: %w", err)
	}

	fmt.Println("Successfully connected to Supabase database")

	// Create migrations table if it doesn't exist
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get already applied migrations
	appliedMigrations := make(map[string]bool)
	rows, err := pool.Query(ctx, "SELECT name FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("error querying applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return fmt.Errorf("error scanning migration name: %w", err)
		}
		appliedMigrations[name] = true
	}

	// Find migration files
	migrations, err := getMigrationFiles("migrations")
	if err != nil {
		return err
	}

	fmt.Printf("Found %d migration files to process\n", len(migrations))

	// Apply each migration in a transaction
	for _, migration := range migrations {
		// Skip if already applied
		if appliedMigrations[migration.Name] {
			fmt.Printf("Migration already applied: %s\n", migration.Name)
			continue
		}

		fmt.Printf("Applying migration: %s\n", migration.Name)

		// Begin transaction
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("error starting transaction: %w", err)
		}

		// Execute migration
		_, err = tx.Exec(ctx, migration.Content)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("error executing migration %s: %w", migration.Name, err)
		}

		// Record migration
		_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (name) VALUES ($1)", migration.Name)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("error recording migration %s: %w", migration.Name, err)
		}

		// Commit transaction
		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("error committing transaction for %s: %w", migration.Name, err)
		}

		fmt.Printf("Successfully applied migration: %s\n", migration.Name)
	}

	return nil
}

// getMigrationFiles returns a list of migration files to execute
func getMigrationFiles(migrationDir string) ([]MigrationFile, error) {
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations directory: %w", err)
	}

	var migrations []MigrationFile
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		path := filepath.Join(migrationDir, file.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error reading migration file %s: %w", file.Name(), err)
		}

		migrations = append(migrations, MigrationFile{
			Path:    path,
			Name:    file.Name(),
			Content: string(content),
		})
	}

	// Sort migrations by name (which includes the timestamp)
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	return migrations, nil
}

// promptConnectionInfo prompts the user for Supabase connection details
func promptConnectionInfo() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter your Supabase database connection details:")

	// Offer user to input a complete connection string first
	fmt.Println("Would you like to enter a complete connection string from the Supabase dashboard?")
	fmt.Print("Enter 'Y' for full connection string or 'N' to input details separately [Y]: ")
	fullConnOption, _ := reader.ReadString('\n')
	fullConnOption = strings.TrimSpace(fullConnOption)

	if fullConnOption == "" || fullConnOption == "Y" || fullConnOption == "y" {
		fmt.Println("\nGet the connection string from:")
		fmt.Println("Supabase Dashboard -> Project Settings -> Database -> Connection string -> Show URI")
		fmt.Print("\nEnter your connection string: ")
		connString, _ := reader.ReadString('\n')
		connString = strings.TrimSpace(connString)
		
		return connString
	}

	fmt.Print("\nProject Reference ID (from Supabase URL, e.g. 'xyzproject'): ")
	projectRef, _ := reader.ReadString('\n')
	projectRef = strings.TrimSpace(projectRef)

	fmt.Print("Database Password (from Supabase dashboard): ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	fmt.Print("Connection Type (1=Direct connection, 2=Connection pooler) [2]: ")
	connType, _ := reader.ReadString('\n')
	connType = strings.TrimSpace(connType)

	if connType == "" || connType == "2" {
		// Use connection pooler (port 6543) - newer connection string format
		fmt.Print("Connection Mode (1=read-only, 2=read-write) [2]: ")
		mode, _ := reader.ReadString('\n')
		mode = strings.TrimSpace(mode)
		
		var poolerEndpoint string
		if mode == "1" {
			poolerEndpoint = "connection-pool-ro"
		} else {
			poolerEndpoint = "connection-pool"
		}
		
		// Modern Supabase pooler format
		connString := fmt.Sprintf(
			"postgres://postgres.%s:%s@%s.supabase.co:6543/postgres?sslmode=require",
			projectRef, password, poolerEndpoint,
		)
		return connString
	} else {
		// Use direct connection with connection string
		fmt.Print("Database Name (usually 'postgres'): ")
		dbName, _ := reader.ReadString('\n')
		dbName = strings.TrimSpace(dbName)
		if dbName == "" {
			dbName = "postgres"
		}

		fmt.Print("Username (usually 'postgres' or 'postgres.{project_ref}'): ")
		user, _ := reader.ReadString('\n')
		user = strings.TrimSpace(user)
		if user == "" {
			user = fmt.Sprintf("postgres.%s", projectRef)
		}

		// Build connection string for direct connection
		connString := fmt.Sprintf(
			"postgres://%s:%s@db.%s.supabase.co:5432/%s?sslmode=require",
			user, password, projectRef, dbName,
		)
		return connString
	}
}

func main() {
	// Get connection info
	connString := promptConnectionInfo()

	// Run migrations
	ctx := context.Background()
	fmt.Println("Starting migration to Supabase...")
	
	if err := Migrate(ctx, connString); err != nil {
		fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Migration to Supabase completed successfully!")
}
