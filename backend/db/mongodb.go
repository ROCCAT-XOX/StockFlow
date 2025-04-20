// backend/db/mongodb.go
package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB Connection-URI
const mongoURI = "mongodb://localhost:27017"
const dbName = "StockFlow" // Datenbankname

// DBClient ist der shared MongoDB-Client
var DBClient *mongo.Client

// ConnectDB stellt eine Verbindung zur MongoDB her
func ConnectDB() error {
	// Verbindungskontext mit Timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verbindung zur MongoDB herstellen
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Printf("Fehler beim Verbinden zur MongoDB: %v", err)
		return err
	}

	// Ping zur Überprüfung der Verbindung
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("Fehler beim Pingen der MongoDB: %v", err)
		return err
	}

	DBClient = client
	log.Println("Erfolgreich mit MongoDB verbunden")

	// Stellen sicher, dass die benötigten Collections existieren
	EnsureCollections()

	return nil
}

// GetCollection gibt eine Kollektion aus der Datenbank zurück
func GetCollection(collectionName string) *mongo.Collection {
	return DBClient.Database(dbName).Collection(collectionName)
}

// DisconnectDB trennt die Verbindung zur MongoDB
func DisconnectDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := DBClient.Disconnect(ctx); err != nil {
		log.Printf("Fehler beim Trennen der MongoDB-Verbindung: %v", err)
		return err
	}

	log.Println("Verbindung zur MongoDB getrennt")
	return nil
}

// EnsureCollections stellt sicher, dass alle benötigten Collections in der Datenbank existieren
func EnsureCollections() {
	// Liste der Collections, die in der Datenbank existieren sollten
	collections := []string{
		"users",        // Benutzer
		"articles",     // Artikel
		"activities",   // Aktivitäten
		"suppliers",    // Lieferanten (für zukünftige Erweiterung)
		"transactions", // Bewegungen/Transaktionen (für zukünftige Erweiterung)
	}

	// Mit einfachen Anfragen sicherstellen, dass die Collections existieren
	for _, collName := range collections {
		// Eine leere Anfrage ausführen, um die Collection zu erstellen falls sie nicht existiert
		GetCollection(collName).FindOne(context.Background(), map[string]interface{}{})
	}
}
