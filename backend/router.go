package backend

import (
	"StockFlow/backend/db"
	"StockFlow/backend/handler"
	"StockFlow/backend/middleware"
	"StockFlow/backend/model"
	"StockFlow/backend/repository"
	"StockFlow/backend/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// InitializeRoutes setzt alle Routen für die Anwendung auf
func InitializeRoutes(router *gin.Engine) {
	// Stelle sicher, dass die Datenbankverbindung hergestellt ist
	if err := db.ConnectDB(); err != nil {
		panic("Fehler beim Verbinden zur Datenbank")
	}

	// Public routes (keine Authentifizierung erforderlich)
	router.GET("/login", func(c *gin.Context) {
		// Token aus dem Cookie extrahieren
		tokenString, err := c.Cookie("token")
		if err == nil && tokenString != "" {
			// Token validieren
			_, err := utils.ValidateJWT(tokenString)
			if err == nil {
				// Gültiges Token, zum Dashboard umleiten
				c.Redirect(http.StatusFound, "/dashboard")
				return
			}
		}

		// Kein Token oder ungültiges Token, Login-Seite anzeigen
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login",
			"year":  time.Now().Year(),
		})
	})

	// Auth-Handler erstellen
	authHandler := handler.NewAuthHandler()
	router.POST("/auth", authHandler.Login)
	router.GET("/logout", authHandler.Logout)

	// Auth middleware für geschützte Routen
	authorized := router.Group("/")
	authorized.Use(middleware.AuthMiddleware())
	{
		// User-Handler
		userHandler := handler.NewUserHandler()

		// Root-Pfad zum Dashboard umleiten
		router.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/dashboard")
		})

		// Dashboard
		// Die Variable dashboardHandler wird nicht verwendet, wir entfernen sie
		authorized.GET("/dashboard", func(c *gin.Context) {
			user, _ := c.Get("user")
			userModel := user.(*model.User)

			// Repositories und Services initialisieren
			articleRepo := repository.NewArticleRepository()
			transactionRepo := repository.NewTransactionRepository()
			supplierRepo := repository.NewSupplierRepository()
			activityRepo := repository.NewActivityRepository()
			// stockService wird nicht verwendet, wir entfernen es

			// Artikel-Statistiken abrufen
			allArticles, err := articleRepo.FindAll()
			if err != nil {
				allArticles = []*model.Article{} // Leere Liste im Fehlerfall
			}

			totalArticles := len(allArticles)

			// Werte für das Dashboard berechnen
			var totalStock float64
			var totalStockValue float64
			var lowStockCount int
			var totalCategories = make(map[string]bool)

			for _, article := range allArticles {
				totalStock += article.StockCurrent
				totalStockValue += article.StockCurrent * article.PurchasePriceNet

				if article.StockCurrent < article.MinimumStock && article.MinimumStock > 0 {
					lowStockCount++
				}

				if article.Category != "" {
					totalCategories[article.Category] = true
				}
			}

			// Kategorien zählen
			categoryCount := len(totalCategories)

			// Lieferantenzählung
			supplierCount, _ := supplierRepo.Count()

			// Artikel unter Mindestbestand
			lowStockArticles, err := articleRepo.FindLowStock(10) // Die Top 10
			if err != nil {
				lowStockArticles = []*model.Article{} // Leere Liste im Fehlerfall
			}

			// Neueste Transaktionen
			recentTransactions, err := transactionRepo.FindRecent(5)
			if err != nil {
				recentTransactions = []*model.Transaction{} // Leere Liste im Fehlerfall
			}

			// Neueste Aktivitäten
			recentActivitiesData, err := activityRepo.FindRecent(10)
			if err != nil {
				recentActivitiesData = []*model.Activity{} // Leere Liste im Fehlerfall
			}

			// Aktivitäten in ein Template-freundlicheres Format konvertieren
			var recentActivities []gin.H
			for i, activity := range recentActivitiesData {
				var message string

				switch activity.Type {
				case model.ActivityTypeArticleAdded:
					message = fmt.Sprintf("<a href=\"/articles/view/%s\" class=\"font-medium text-gray-900\">%s</a> wurde als neuer Artikel hinzugefügt",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeArticleUpdated:
					message = fmt.Sprintf("Artikel <a href=\"/articles/view/%s\" class=\"font-medium text-gray-900\">%s</a> wurde aktualisiert",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeArticleDeleted:
					message = fmt.Sprintf("Artikel <span class=\"font-medium text-gray-900\">%s</span> wurde gelöscht",
						activity.TargetName)
				case model.ActivityTypeStockAdjusted:
					message = fmt.Sprintf("Bestand für <a href=\"/articles/view/%s\" class=\"font-medium text-gray-900\">%s</a> wurde angepasst",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeStockTaking:
					message = fmt.Sprintf("Inventur für <a href=\"/articles/view/%s\" class=\"font-medium text-gray-900\">%s</a> wurde durchgeführt",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeUserAdded:
					message = fmt.Sprintf("Benutzer <span class=\"font-medium text-gray-900\">%s</span> wurde hinzugefügt",
						activity.TargetName)
				case model.ActivityTypeUserUpdated:
					message = fmt.Sprintf("Benutzer <span class=\"font-medium text-gray-900\">%s</span> wurde aktualisiert",
						activity.TargetName)
				case model.ActivityTypeUserDeleted:
					message = fmt.Sprintf("Benutzer <span class=\"font-medium text-gray-900\">%s</span> wurde entfernt",
						activity.TargetName)
				case model.ActivityTypeSupplierAdded:
					message = fmt.Sprintf("Lieferant <a href=\"/suppliers/view/%s\" class=\"font-medium text-gray-900\">%s</a> wurde hinzugefügt",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeSupplierUpdated:
					message = fmt.Sprintf("Lieferant <a href=\"/suppliers/view/%s\" class=\"font-medium text-gray-900\">%s</a> wurde aktualisiert",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeSupplierDeleted:
					message = fmt.Sprintf("Lieferant <span class=\"font-medium text-gray-900\">%s</span> wurde entfernt",
						activity.TargetName)
				default:
					message = activity.Description
				}

				recentActivities = append(recentActivities, gin.H{
					"IconBgClass": activity.GetIconClass(),
					"IconSVG":     activity.GetIconSVG(),
					"Message":     message,
					"Time":        activity.FormatTimeAgo(),
					"IsLast":      i == len(recentActivitiesData)-1,
				})
			}

			// Lager-Bewegungsdaten für Diagramme
			stockMovementData, err := transactionRepo.GetStockMovementSummary()
			if err != nil {
				stockMovementData = map[string][]float64{
					"stockIn":    {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					"stockOut":   {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					"adjustment": {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				}
			}

			// Monatsbezeichnungen für Charts
			monthLabels := []string{"Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"}

			// Verteilung nach Kategorien
			categorySummary, err := articleRepo.GetCategorySummary()
			if err != nil {
				categorySummary = map[string][]interface{}{
					"labels": {"Keine Kategorie"},
					"counts": {0},
					"values": {0.0},
				}
			}

			// Transaktionen der letzten 30 Tage
			thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
			recentTransactionsCount, _ := transactionRepo.CountSince(thirtyDaysAgo)

			// Daten an das Template übergeben
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"title":                   "Dashboard",
				"active":                  "dashboard",
				"user":                    userModel.FirstName + " " + userModel.LastName,
				"email":                   userModel.Email,
				"year":                    time.Now().Year(),
				"userRole":                c.GetString("userRole"),
				"totalArticles":           totalArticles,
				"totalStock":              totalStock,
				"totalStockValue":         totalStockValue,
				"lowStockCount":           lowStockCount,
				"categoryCount":           categoryCount,
				"supplierCount":           supplierCount,
				"recentTransactionsCount": recentTransactionsCount,
				"lowStockArticles":        lowStockArticles,
				"recentTransactions":      recentTransactions,
				"recentActivities":        recentActivities,
				"stockMovementData":       stockMovementData,
				"monthLabels":             monthLabels,
				"categoryLabels":          categorySummary["labels"],
				"categoryCounts":          categorySummary["counts"],
				"categoryValues":          categorySummary["values"],
			})
		})

		// Benutzerprofilrouten
		authorized.GET("/profile", userHandler.ShowUserProfile)

		// Einstellungsrouten (für alle Benutzer)
		authorized.GET("/settings", userHandler.ShowSettings)

		// Benutzerverwaltungsrouten (für Administratoren)
		authorized.POST("/users/add", middleware.RoleMiddleware(model.RoleAdmin), userHandler.AddUser)
		authorized.POST("/users/edit/:id", middleware.RoleMiddleware(model.RoleAdmin), userHandler.UpdateUser)
		authorized.DELETE("/users/delete/:id", middleware.RoleMiddleware(model.RoleAdmin), userHandler.DeleteUser)

		// Passwortänderungsroute - ein Benutzer kann nur sein eigenes Passwort ändern
		authorized.POST("/users/change-password", middleware.SelfOrAdminMiddleware(), userHandler.ChangePassword)

		// Artikel-Routen
		articleHandler := handler.NewArticleHandler()
		authorized.GET("/articles", articleHandler.ListArticles)
		authorized.GET("/articles/add", articleHandler.ShowAddArticleForm)
		authorized.POST("/articles/add", articleHandler.AddArticle)
		authorized.GET("/articles/view/:id", articleHandler.GetArticleDetails)
		authorized.GET("/articles/edit/:id", articleHandler.ShowEditArticleForm)
		authorized.POST("/articles/edit/:id", articleHandler.UpdateArticle)
		authorized.DELETE("/articles/delete/:id", articleHandler.DeleteArticle)
		authorized.GET("/stock", articleHandler.ShowStockOverview)

		// Transaktions-Routen
		transactionHandler := handler.NewTransactionHandler()
		authorized.GET("/transactions", transactionHandler.ListTransactions)
		authorized.GET("/transactions/add", transactionHandler.ShowAddTransactionForm)
		authorized.POST("/transactions/add", transactionHandler.AddTransaction)
		authorized.GET("/transactions/view/:id", transactionHandler.GetTransactionDetails)

		// Lieferanten-Routen
		supplierHandler := handler.NewSupplierHandler()
		authorized.GET("/suppliers", supplierHandler.ListSuppliers)
		authorized.GET("/suppliers/add", supplierHandler.ShowAddSupplierForm)
		authorized.POST("/suppliers/add", supplierHandler.AddSupplier)
		authorized.GET("/suppliers/view/:id", supplierHandler.GetSupplierDetails)
		authorized.GET("/suppliers/edit/:id", supplierHandler.ShowEditSupplierForm)
		authorized.POST("/suppliers/edit/:id", supplierHandler.UpdateSupplier)
		authorized.DELETE("/suppliers/delete/:id", supplierHandler.DeleteSupplier)

		// Optionale API-Endpoints für AJAX-Anfragen
		api := router.Group("/api")
		api.Use(middleware.AuthMiddleware())
		{
			api.DELETE("/articles/:id", articleHandler.DeleteArticle)
			api.DELETE("/suppliers/:id", supplierHandler.DeleteSupplier)
		}
	}
}
