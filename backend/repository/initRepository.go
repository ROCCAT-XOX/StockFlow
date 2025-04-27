// backend/repository/initRepository.go
package repository

import (
	"context"
	"log"
	"time"

	"StockFlow/backend/db"
	"go.mongodb.org/mongo-driver/bson"
)

// InitRepository ist für die Initialisierung der Datenbank zuständig
type InitRepository struct {
	articleRepo *ArticleRepository
	userRepo    *UserRepository
}

// NewInitRepository erstellt ein neues InitRepository
func NewInitRepository() *InitRepository {
	return &InitRepository{
		articleRepo: NewArticleRepository(),
		userRepo:    NewUserRepository(),
	}
}

// InitializeDatabase initialisiert die Datenbank mit Beispieldaten, falls nötig
func (r *InitRepository) InitializeDatabase() error {
	// Zuerst Admin-Benutzer erstellen, falls keiner existiert
	err := r.userRepo.CreateAdminUserIfNotExists()
	if err != nil {
		log.Printf("Warnung: Admin-Benutzer konnte nicht erstellt werden: %v", err)
	} else {
		log.Println("Admin-Benutzer wurde überprüft/erstellt")
	}

	return nil
}

// countArticles zählt die Anzahl der Artikel in der Datenbank
func (r *InitRepository) countArticles() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := db.GetCollection("articles")
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}

	return count, nil
}
