package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"

	"chorequest/backend/graph"
	"chorequest/backend/graph/model"
	appauth "chorequest/backend/internal/auth"
	repopkg "chorequest/backend/internal/repo"
	sqlitedb "chorequest/backend/internal/sqlite"
)

func main() {
	// Load environment from .env if present (for local dev)
	_ = godotenv.Load()

	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "./data/dev.sqlite"
	}
	sqliteDB, err := sqlitedb.Open(context.Background(), sqlitePath)
	if err != nil {
		log.Fatalf("sqlite open error: %v", err)
	}
	defer sqliteDB.Close()
	appRepo := repopkg.NewSQLiteRepo(sqliteDB)
	log.Printf("using SQLite repository at %s", sqlitePath)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Dev auth endpoint: issues a JWT with sub + role (and inferred family) for quick testing
	r.Post("/auth/dev", func(w http.ResponseWriter, r *http.Request) {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			http.Error(w, "JWT_SECRET not set", http.StatusPreconditionFailed)
			return
		}
		role := r.URL.Query().Get("role")
		if role == "" {
			role = "PARENT"
		}
		sub := r.URL.Query().Get("sub")
		if sub == "" {
			sub = "dev-user"
		}

		familyID := r.URL.Query().Get("familyId")
		if familyID == "" {
			switch role {
			case "PARENT":
				if fams, err := appRepo.ListFamiliesForOwner(r.Context(), sub); err == nil && len(fams) > 0 {
					familyID = fams[0].ID
				} else {
					if fam, err := appRepo.CreateFamily(r.Context(), model.NewFamily{OwnerID: sub, Name: fmt.Sprintf("Family %s", sub)}); err == nil {
						familyID = fam.ID
					} else if err != nil {
						log.Printf("dev auth: create family failed: %v", err)
					}
				}
			case "CHILD":
				if child, err := appRepo.GetChild(r.Context(), sub); err == nil && child != nil {
					familyID = child.FamilyID
				}
			}
		}

		claims := jwt.MapClaims{
			"sub":  sub,
			"role": role,
			"iat":  time.Now().Unix(),
		}
		if familyID != "" {
			claims["familyId"] = familyID
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		s, err := token.SignedString([]byte(secret))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"token":    s,
			"familyId": familyID,
		})
	})

	// GraphQL endpoint (gqlgen)
	gql := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Repo: appRepo}}))
	r.Method("POST", "/query", gql)
	r.Method("GET", "/query", gql) // allow GET for basic tests
	// GraphQL Playground (legacy) — keep available for reference
	r.Get("/play", func(w http.ResponseWriter, r *http.Request) {
		playground.Handler("GraphQL", "/query").ServeHTTP(w, r)
	})

	// Example protected route using JWT middleware (not required yet by schema)
	r.Group(func(pr chi.Router) {
		pr.Use(appauth.JWTMiddleware(os.Getenv("JWT_SECRET")))
		pr.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			sub := appauth.SubjectFromContext(r.Context())
			if sub == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{\"sub\": \"" + sub + "\"}"))
		})
	})

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}

	server := &http.Server{Addr: addr, Handler: r, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second}
	log.Printf("GraphQL server listening on %s/query", addr)
	log.Fatal(server.ListenAndServe())
}
