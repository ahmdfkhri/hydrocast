package main

import (
	"context"
	"log"

	"github.com/ahmdfkhri/hydrocast/backend/config"
	"github.com/ahmdfkhri/hydrocast/backend/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.New()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminConfig.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	// update first (if both username & email match)
	query := `
		INSERT INTO users (username, email, password, role)
		VALUES ($1, $2, $3, 'admin')
		ON CONFLICT (email) 
		DO UPDATE SET 
				username = EXCLUDED.username,
				password = EXCLUDED.password,
				role = EXCLUDED.role
		RETURNING id
	`

	db := database.NewConnection(&cfg.DatabaseConfig)
	defer db.Close()

	ctx := context.Background()
	_, err = db.Exec(ctx, query,
		cfg.AdminConfig.Username, cfg.AdminConfig.Email, string(hashedPassword))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("admin user seeded/updated")
}
