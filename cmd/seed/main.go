package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Define dummy data generators
func generateDummyUsers(count int) []map[string]interface{} {
	firstNames := []string{"John", "Jane", "Michael", "Emily", "David", "Sarah", "Robert", "Lisa"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Miller", "Davis", "Garcia"}

	users := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		userId := uuid.New()
		firstNameIndex := i % len(firstNames)
		lastNameIndex := i % len(lastNames)

		firstName := firstNames[firstNameIndex]
		lastName := lastNames[lastNameIndex]
		email := fmt.Sprintf("%s.%s@example.com", strings.ToLower(firstName), strings.ToLower(lastName))

		users[i] = map[string]interface{}{
			"id":        userId,
			"firstname": firstName,
			"lastname":  lastName,
			"email":     email,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}
	}
	return users
}

func generateDummyMetrics(count int) []map[string]interface{} {
	metrics := make([]map[string]interface{}, count)
	metricTypes := []string{"P/E Ratio", "EPS", "Market Cap", "Revenue Growth", "Dividend Yield", "Debt/Equity Ratio"}
	metricDesc := []string{
		"Price-to-Earnings Ratio: A measure of valuation",
		"Earnings Per Share: A measure of profitability",
		"Market Capitalization: Total value of a company's shares",
		"Growth in company revenue over time",
		"The dividend income relative to the stock price",
		"A measure of a company's financial leverage",
	}

	for i := 0; i < count; i++ {
		metricId := uuid.New()
		index := i % len(metricTypes)

		metrics[i] = map[string]interface{}{
			"id":          metricId,
			"name":        metricTypes[index],
			"description": metricDesc[index],
		}
	}
	return metrics
}

func generateDummyStocks(count int) []map[string]interface{} {
	stocks := make([]map[string]interface{}, count)
	tickers := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA", "META", "NVDA", "JPM", "V", "JNJ"}
	companies := []string{
		"Apple Inc.",
		"Microsoft Corporation",
		"Alphabet Inc.",
		"Amazon.com Inc.",
		"Tesla Inc.",
		"Meta Platforms Inc.",
		"NVIDIA Corporation",
		"JPMorgan Chase & Co.",
		"Visa Inc.",
		"Johnson & Johnson",
	}

	for i := 0; i < count; i++ {
		stockId := uuid.New()
		index := i % len(tickers)

		stocks[i] = map[string]interface{}{
			"id":      stockId,
			"symbol":  tickers[index],
			"company": companies[index],
		}
	}
	return stocks
}

func generateDummyStockMetrics(stocks []map[string]interface{}, metrics []map[string]interface{}) []map[string]interface{} {
	stockMetrics := []map[string]interface{}{}

	for _, stock := range stocks {
		for _, metric := range metrics {
			stockMetric := map[string]interface{}{
				"id":         uuid.New(),
				"stock_id":   stock["id"],
				"metric_id":  metric["id"],
				"value":      rand.Float64() * 100, // Random metric value
				"recorded_at": time.Now(),
			}
			stockMetrics = append(stockMetrics, stockMetric)
		}
	}
	return stockMetrics
}

func generateDummyAnalyses(stocks []map[string]interface{}, users []map[string]interface{}) []map[string]interface{} {
	analyses := []map[string]interface{}{}
	analysisTypes := []string{"Technical", "Fundamental", "Sentiment", "Financial"}

	for _, stock := range stocks {
		for i := 0; i < 2; i++ { // 2 analyses per stock
			userIndex := rand.Intn(len(users))
			analysisType := analysisTypes[rand.Intn(len(analysisTypes))]

			analysis := map[string]interface{}{
				"id":         uuid.New(),
				"stock_id":   stock["id"],
				"user_id":    users[userIndex]["id"],
				"type":       analysisType,
				"content":    fmt.Sprintf("Analysis of %s using %s approach", stock["ticker"], analysisType),
				"created_at": time.Now(),
			}
			analyses = append(analyses, analysis)
		}
	}
	return analyses
}

func generateDummyRecommendations(analyses []map[string]interface{}) []map[string]interface{} {
	recommendations := []map[string]interface{}{}
	actions := []string{"Buy", "Sell", "Hold", "Strong Buy", "Strong Sell"}

	for _, analysis := range analyses {
		recommendation := map[string]interface{}{
			"id":           uuid.New(),
			"analysis_id":  analysis["id"],
			"action":       actions[rand.Intn(len(actions))],
			"target_price": rand.Float64() * 1000,
			"confidence":   rand.Float64(),
			"created_at":   time.Now(),
		}
		recommendations = append(recommendations, recommendation)
	}
	return recommendations
}

// Insert data into database
func insertDummyData(ctx context.Context, pool *pgxpool.Pool) error {
	// Generate dummy data
	users := generateDummyUsers(5)
	metrics := generateDummyMetrics(8)
	stocks := generateDummyStocks(10)
	stockMetrics := generateDummyStockMetrics(stocks, metrics)
	analyses := generateDummyAnalyses(stocks, users)
	recommendations := generateDummyRecommendations(analyses)

	// Start a transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Will be ignored if transaction is committed

	// Insert users
	fmt.Println("Inserting dummy users...")
	for _, user := range users {
		_, err := tx.Exec(ctx, `
			INSERT INTO users (id, firstname, lastname, email, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, user["id"], user["firstname"], user["lastname"], user["email"], 
		   user["created_at"], user["updated_at"])

		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}
	}

	// Insert metrics
	fmt.Println("Inserting dummy metrics...")
	for _, metric := range metrics {
		_, err := tx.Exec(ctx, `
			INSERT INTO metric (id, name, description)
			VALUES ($1, $2, $3)
		`, metric["id"], metric["name"], metric["description"])

		if err != nil {
			return fmt.Errorf("failed to insert metric: %w", err)
		}
	}

	// Insert stocks
	fmt.Println("Inserting dummy stocks...")
	for _, stock := range stocks {
		_, err := tx.Exec(ctx, `
			INSERT INTO stock (id, symbol, company)
			VALUES ($1, $2, $3)
		`, stock["id"], stock["symbol"], stock["company"])

		if err != nil {
			return fmt.Errorf("failed to insert stock: %w", err)
		}
	}

	// Insert stock metrics
	fmt.Println("Inserting dummy stock metrics...")
	for _, stockMetric := range stockMetrics {
		_, err := tx.Exec(ctx, `
			INSERT INTO stock_metric (id, stock_id, metric_id, value, recorded_at)
			VALUES ($1, $2, $3, $4, $5)
		`, stockMetric["id"], stockMetric["stock_id"], stockMetric["metric_id"], 
		   stockMetric["value"], stockMetric["recorded_at"])

		if err != nil {
			return fmt.Errorf("failed to insert stock metric: %w", err)
		}
	}

	// Insert analyses
	fmt.Println("Inserting dummy analyses...")
	for _, analysis := range analyses {
		_, err := tx.Exec(ctx, `
			INSERT INTO analysis (id, stock_id, user_id, type, content, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, analysis["id"], analysis["stock_id"], analysis["user_id"],
			analysis["type"], analysis["content"], analysis["created_at"])

		if err != nil {
			return fmt.Errorf("failed to insert analysis: %w", err)
		}
	}

	// Insert recommendations
	fmt.Println("Inserting dummy recommendations...")
	for _, recommendation := range recommendations {
		_, err := tx.Exec(ctx, `
			INSERT INTO recommendation (id, analysis_id, action, target_price, confidence, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, recommendation["id"], recommendation["analysis_id"], recommendation["action"],
			recommendation["target_price"], recommendation["confidence"], recommendation["created_at"])

		if err != nil {
			return fmt.Errorf("failed to insert recommendation: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
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
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Get connection info
	connString := promptConnectionInfo()

	// Set up context
	ctx := context.Background()
	fmt.Println("Starting to seed Supabase database with dummy data...")

	// Connect to database
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Ping database failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully connected to Supabase database")

	// Insert dummy data
	if err := insertDummyData(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to insert dummy data: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Dummy data successfully seeded!")
}
