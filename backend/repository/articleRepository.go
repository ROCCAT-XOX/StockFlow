// backend/repository/articleRepository.go (erweiterte Version)
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

	// Standardmäßig aktiv setzen, falls nicht anders angegeben
	if !article.IsActive {
		article.IsActive = true
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

// FindLowStock findet Artikel, deren Bestand unter dem Mindestbestand liegt
func (r *ArticleRepository) FindLowStock(limit int) ([]*model.Article, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Filter für Artikel mit Bestand unter oder gleich Mindestbestand
	filter := bson.M{
		"$expr": bson.M{
			"$lte": []interface{}{"$stockCurrent", "$minimumStock"},
		},
		"minimumStock": bson.M{"$gt": 0}, // Nur Artikel mit definiertem Mindestbestand
		"isActive":     true,             // Nur aktive Artikel
	}

	options := options.Find().
		SetSort(bson.D{{Key: "stockCurrent", Value: 1}}).
		SetLimit(int64(limit))

	var articles []*model.Article
	cursor, err := r.collection.Find(ctx, filter, options)
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

// FindBySupplierID findet alle Artikel eines bestimmten Lieferanten
func (r *ArticleRepository) FindBySupplierID(supplierID string) ([]*model.Article, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var articles []*model.Article
	suppID, err := primitive.ObjectIDFromHex(supplierID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"supplierId": suppID})
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

// FindByStockStatus findet Artikel basierend auf ihrem Bestandsstatus
func (r *ArticleRepository) FindByStockStatus(status string) ([]*model.Article, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var filter bson.M

	switch status {
	case "low":
		filter = bson.M{
			"$expr": bson.M{
				"$lte": []interface{}{"$stockCurrent", "$minimumStock"},
			},
			"minimumStock": bson.M{"$gt": 0},
		}
	case "high":
		filter = bson.M{
			"$expr": bson.M{
				"$gte": []interface{}{"$stockCurrent", "$maximumStock"},
			},
			"maximumStock": bson.M{"$gt": 0},
		}
	case "ok":
		filter = bson.M{
			"$expr": bson.M{
				"$and": []bson.M{
					{"$gt": []interface{}{"$stockCurrent", "$minimumStock"}},
					{"$or": []bson.M{
						{"$lt": []interface{}{"$stockCurrent", "$maximumStock"}},
						{"$eq": []interface{}{"$maximumStock", 0}},
					}},
				},
			},
			"minimumStock": bson.M{"$gt": 0},
		}
	case "zero":
		filter = bson.M{"stockCurrent": 0, "isActive": true}
	default:
		filter = bson.M{} // Alle Artikel
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

// FindByCategoryAndStockStatus kombiniert die Filter für Kategorie und Bestandsstatus
func (r *ArticleRepository) FindByCategoryAndStockStatus(category, status string) ([]*model.Article, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var baseFilter bson.M

	switch status {
	case "low":
		baseFilter = bson.M{
			"$expr": bson.M{
				"$lte": []interface{}{"$stockCurrent", "$minimumStock"},
			},
			"minimumStock": bson.M{"$gt": 0},
		}
	case "high":
		baseFilter = bson.M{
			"$expr": bson.M{
				"$gte": []interface{}{"$stockCurrent", "$maximumStock"},
			},
			"maximumStock": bson.M{"$gt": 0},
		}
	case "ok":
		baseFilter = bson.M{
			"$expr": bson.M{
				"$and": []bson.M{
					{"$gt": []interface{}{"$stockCurrent", "$minimumStock"}},
					{"$or": []bson.M{
						{"$lt": []interface{}{"$stockCurrent", "$maximumStock"}},
						{"$eq": []interface{}{"$maximumStock", 0}},
					}},
				},
			},
			"minimumStock": bson.M{"$gt": 0},
		}
	case "zero":
		baseFilter = bson.M{"stockCurrent": 0}
	default:
		baseFilter = bson.M{} // Kein Status-Filter
	}

	// Kategorie-Filter hinzufügen
	baseFilter["category"] = category

	var articles []*model.Article
	cursor, err := r.collection.Find(ctx, baseFilter)
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

// Count zählt die Gesamtzahl der Artikel
func (r *ArticleRepository) Count() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{})
}

// CountLowStock zählt die Anzahl der Artikel unter Mindestbestand
func (r *ArticleRepository) CountLowStock() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$expr": bson.M{
			"$lte": []interface{}{"$stockCurrent", "$minimumStock"},
		},
		"minimumStock": bson.M{"$gt": 0},
		"isActive":     true,
	}

	return r.collection.CountDocuments(ctx, filter)
}

// GetAllCategories gibt eine Liste aller vorhandenen Kategorien zurück
func (r *ArticleRepository) GetAllCategories() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregation für eindeutige Kategorien
	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$category"}},
		{"$match": bson.M{"_id": bson.M{"$ne": ""}}}, // Leere Kategorien ausschließen
		{"$sort": bson.M{"_id": 1}},                  // Aufsteigend sortieren
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID string `bson:"_id"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	categories := make([]string, len(results))
	for i, result := range results {
		categories[i] = result.ID
	}

	return categories, nil
}

// CountCategories zählt die Anzahl eindeutiger Kategorien
func (r *ArticleRepository) CountCategories() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregation für die Anzahl eindeutiger Kategorien
	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$category"}},
		{"$match": bson.M{"_id": bson.M{"$ne": ""}}}, // Leere Kategorien ausschließen
		{"$count": "count"},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Count int64 `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	return results[0].Count, nil
}

// CalculateTotalStockValue berechnet den Gesamtwert aller Artikel im Lager
func (r *ArticleRepository) CalculateTotalStockValue() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregation für den Gesamtwert (Bestand * Einkaufspreis)
	pipeline := []bson.M{
		{
			"$project": bson.M{
				"value": bson.M{
					"$multiply": []string{"$stockCurrent", "$purchasePriceNet"},
				},
			},
		},
		{
			"$group": bson.M{
				"_id":        nil,
				"totalValue": bson.M{"$sum": "$value"},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalValue float64 `bson:"totalValue"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	return results[0].TotalValue, nil
}

// CalculateTotalStock berechnet die Gesamtmenge aller Artikel im Lager
func (r *ArticleRepository) CalculateTotalStock() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregation für die Gesamtmenge
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":        nil,
				"totalStock": bson.M{"$sum": "$stockCurrent"},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalStock float64 `bson:"totalStock"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	return results[0].TotalStock, nil
}

// GetCategorySummary gibt eine Zusammenfassung der Artikel nach Kategorien zurück
func (r *ArticleRepository) GetCategorySummary() (map[string][]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregation für Kategorie-Zusammenfassung
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"category": bson.M{"$ne": ""}, // Artikel ohne Kategorie ausschließen
			},
		},
		{
			"$group": bson.M{
				"_id":   "$category",
				"count": bson.M{"$sum": 1},
				"totalValue": bson.M{
					"$sum": bson.M{
						"$multiply": []string{"$stockCurrent", "$purchasePriceNet"},
					},
				},
			},
		},
		{
			"$sort": bson.M{"count": -1}, // Absteigend nach Anzahl sortieren
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID         string  `bson:"_id"`
		Count      int     `bson:"count"`
		TotalValue float64 `bson:"totalValue"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Ergebnisse in für das Template geeignetes Format umwandeln
	labels := make([]interface{}, len(results))
	counts := make([]interface{}, len(results))
	values := make([]interface{}, len(results))

	for i, result := range results {
		labels[i] = result.ID
		counts[i] = result.Count
		values[i] = result.TotalValue
	}

	// Wenn keine Kategorien gefunden wurden, Dummy-Daten zurückgeben
	if len(results) == 0 {
		return map[string][]interface{}{
			"labels": {"Keine Kategorie"},
			"counts": {0},
			"values": {0.0},
		}, nil
	}

	return map[string][]interface{}{
		"labels": labels,
		"counts": counts,
		"values": values,
	}, nil
}
