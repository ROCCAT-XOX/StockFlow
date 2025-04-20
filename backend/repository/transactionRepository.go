// backend/repository/transactionRepository.go
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

// TransactionRepository enthält alle Datenbankoperationen für das Transaction-Modell
type TransactionRepository struct {
	collection *mongo.Collection
}

// NewTransactionRepository erstellt ein neues TransactionRepository
func NewTransactionRepository() *TransactionRepository {
	return &TransactionRepository{
		collection: db.GetCollection("transactions"),
	}
}

// Create erstellt eine neue Transaktion
func (r *TransactionRepository) Create(transaction *model.Transaction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Standardwerte setzen für fehlende Zeitstempel
	if transaction.Timestamp.IsZero() {
		transaction.Timestamp = time.Now()
	}

	result, err := r.collection.InsertOne(ctx, transaction)
	if err != nil {
		return err
	}

	transaction.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByID findet eine Transaktion anhand ihrer ID
func (r *TransactionRepository) FindByID(id string) (*model.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var transaction model.Transaction
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&transaction)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// FindAll findet alle Transaktionen
func (r *TransactionRepository) FindAll() ([]*model.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optionen für die Sortierung nach Zeitstempel (absteigend)
	opts := options.Find().SetSort(bson.D{
		{Key: "timestamp", Value: -1},
	})

	var transactions []*model.Transaction
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var transaction model.Transaction
		if err := cursor.Decode(&transaction); err != nil {
			return nil, err
		}
		transactions = append(transactions, &transaction)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// FindWithFilters findet Transaktionen mit optionalen Filtern und Paginierung
func (r *TransactionRepository) FindWithFilters(articleID, transactionType string, page, perPage int) ([]*model.Transaction, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Filter aufbauen
	filter := bson.M{}

	if articleID != "" {
		objID, err := primitive.ObjectIDFromHex(articleID)
		if err == nil { // Ignoriere Fehler bei der ID-Konvertierung
			filter["articleId"] = objID
		}
	}

	if transactionType != "" {
		filter["type"] = transactionType
	}

	// Optionen für die Sortierung und Paginierung
	findOptions := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64((page - 1) * perPage)).
		SetLimit(int64(perPage))

	// Gesamtzahl der Ergebnisse
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Transaktionen abrufen
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var transactions []*model.Transaction
	for cursor.Next(ctx) {
		var transaction model.Transaction
		if err := cursor.Decode(&transaction); err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, &transaction)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// FindByArticleID findet alle Transaktionen für einen bestimmten Artikel
func (r *TransactionRepository) FindByArticleID(articleID string) ([]*model.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(articleID)
	if err != nil {
		return nil, err
	}

	// Optionen für die Sortierung nach Zeitstempel (absteigend)
	opts := options.Find().SetSort(bson.D{
		{Key: "timestamp", Value: -1},
	})

	var transactions []*model.Transaction
	cursor, err := r.collection.Find(ctx, bson.M{"articleId": objID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var transaction model.Transaction
		if err := cursor.Decode(&transaction); err != nil {
			return nil, err
		}
		transactions = append(transactions, &transaction)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// FindRecent findet die neuesten n Transaktionen
func (r *TransactionRepository) FindRecent(limit int) ([]*model.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optionen für die Sortierung nach Zeitstempel (absteigend) und Limitierung
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit))

	var transactions []*model.Transaction
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var transaction model.Transaction
		if err := cursor.Decode(&transaction); err != nil {
			return nil, err
		}
		transactions = append(transactions, &transaction)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// CountSince zählt die Anzahl der Transaktionen seit einem bestimmten Zeitpunkt
func (r *TransactionRepository) CountSince(since time.Time) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"timestamp": bson.M{"$gte": since},
	}

	return r.collection.CountDocuments(ctx, filter)
}

// GetStockMovementSummary berechnet eine Zusammenfassung der Lagerbewegungen pro Monat
func (r *TransactionRepository) GetStockMovementSummary() (map[string][]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aktuelles Jahr für das Filter
	currentYear := time.Now().Year()
	startOfYear := time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(currentYear+1, 1, 1, 0, 0, 0, 0, time.UTC)

	// Aggregation für monatliche Bewegungen nach Typ
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"timestamp": bson.M{
					"$gte": startOfYear,
					"$lt":  endOfYear,
				},
			},
		},
		{
			"$project": bson.M{
				"type":     "$type",
				"quantity": "$quantity",
				"month":    bson.M{"$month": "$timestamp"},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"month": "$month",
					"type":  "$type",
				},
				"total": bson.M{"$sum": "$quantity"},
			},
		},
		{
			"$sort": bson.M{"_id.month": 1},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID struct {
			Month int                   `bson:"month"`
			Type  model.TransactionType `bson:"type"`
		} `bson:"_id"`
		Total float64 `bson:"total"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Daten für das Chart vorbereiten (ein Array pro Transaktionstyp)
	stockIn := make([]float64, 12)    // Ein Wert pro Monat
	stockOut := make([]float64, 12)   // Ein Wert pro Monat
	adjustment := make([]float64, 12) // Ein Wert pro Monat

	for _, result := range results {
		monthIndex := result.ID.Month - 1 // 0-basierter Index

		switch result.ID.Type {
		case model.TransactionTypeStockIn:
			stockIn[monthIndex] = result.Total
		case model.TransactionTypeStockOut:
			stockOut[monthIndex] = result.Total
		case model.TransactionTypeAdjust, model.TransactionTypeInventory:
			adjustment[monthIndex] = result.Total
		}
	}

	// Die Variable 'labels' wird nicht benötigt, da sie nicht zurückgegeben wird
	// Also entfernen wir sie aus dem Code

	return map[string][]float64{
		"stockIn":    stockIn,
		"stockOut":   stockOut,
		"adjustment": adjustment,
	}, nil
}
