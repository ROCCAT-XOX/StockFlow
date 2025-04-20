// backend/repository/articleRepository.go
package repository

import (
	"context"
	"time"

	"StockFlow/backend/db"
	"StockFlow/backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ArticleRepository enthält alle Datenbankoperationen für das Article-Modell
type ArticleRepository struct {
	collection *mongo.Collection
}

// NewArticleRepository erstellt ein neues ArticleRepository
func NewArticleRepository() *ArticleRepository {
	return &ArticleRepository{
		collection: db.GetCollection("articles"),
	}
}

// Create erstellt einen neuen Artikel
func (r *ArticleRepository) Create(article *model.Article) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prüfen, ob bereits ein Artikel mit dieser Artikelnummer existiert
	count, err := r.collection.CountDocuments(ctx, bson.M{"articleNumber": article.ArticleNumber})
	if err != nil {
		return err
	}
	if count > 0 {
		return mongo.ErrNoDocuments // Fehlertyp hier nicht optimal, aber einfach zu erkennen
	}

	// Standardwerte setzen für fehlende Zeitstempel
	if article.CreatedAt.IsZero() {
		article.CreatedAt = time.Now()
	}
	if article.UpdatedAt.IsZero() {
		article.UpdatedAt = time.Now()
	}

	result, err := r.collection.InsertOne(ctx, article)
	if err != nil {
		return err
	}

	article.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByID findet einen Artikel anhand seiner ID
func (r *ArticleRepository) FindByID(id string) (*model.Article, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var article model.Article
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&article)
	if err != nil {
		return nil, err
	}

	return &article, nil
}

// FindByArticleNumber findet einen Artikel anhand seiner Artikelnummer
func (r *ArticleRepository) FindByArticleNumber(articleNumber string) (*model.Article, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var article model.Article
	err := r.collection.FindOne(ctx, bson.M{"articleNumber": articleNumber}).Decode(&article)
	if err != nil {
		return nil, err
	}

	return &article, nil
}

// FindAll findet alle Artikel
func (r *ArticleRepository) FindAll() ([]*model.Article, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optionen für die Sortierung nach Artikelnummer
	opts := options.Find().SetSort(bson.D{
		{Key: "articleNumber", Value: 1},
	})

	var articles []*model.Article
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var article model.Article
		if err := cursor.Decode(&article); err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

// Update aktualisiert einen Artikel
func (r *ArticleRepository) Update(article *model.Article) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// UpdatedAt-Zeitstempel aktualisieren
	article.UpdatedAt = time.Now()

	// Prüfen, ob bereits ein anderer Artikel mit dieser Artikelnummer existiert
	var existingArticle model.Article
	err := r.collection.FindOne(ctx, bson.M{
		"articleNumber": article.ArticleNumber,
		"_id":           bson.M{"$ne": article.ID},
	}).Decode(&existingArticle)

	// Wenn ein Dokument gefunden wurde, bedeutet das, dass die Artikelnummer bereits verwendet wird
	if err == nil {
		return mongo.ErrNoDocuments
	}
	// Wenn der Fehler nicht "nicht gefunden" ist, ist es ein anderer Fehler
	if err != mongo.ErrNoDocuments {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": article.ID},
		bson.M{"$set": article},
	)
	return err
}

// Delete löscht einen Artikel
func (r *ArticleRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// FindByCategory findet alle Artikel einer bestimmten Kategorie
func (r *ArticleRepository) FindByCategory(category string) ([]*model.Article, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var articles []*model.Article
	cursor, err := r.collection.Find(ctx, bson.M{"category": category})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var article model.Article
		if err := cursor.Decode(&article); err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

// SearchArticles sucht Artikel anhand verschiedener Kriterien
func (r *ArticleRepository) SearchArticles(query string) ([]*model.Article, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Suche in mehreren Feldern
	filter := bson.M{
		"$or": []bson.M{
			{"articleNumber": bson.M{"$regex": query, "$options": "i"}},
			{"shortName": bson.M{"$regex": query, "$options": "i"}},
			{"longName": bson.M{"$regex": query, "$options": "i"}},
			{"ean": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	var articles []*model.Article
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var article model.Article
		if err := cursor.Decode(&article); err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}
