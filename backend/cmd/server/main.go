package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/99designs/gqlgen/graphql/playground"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"

    appauth "chorequest/backend/internal/auth"
    "chorequest/backend/graph"
    "chorequest/backend/internal/db"
    repopkg "chorequest/backend/internal/repo"
    "github.com/golang-jwt/jwt/v5"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment from .env if present (for local dev)
    _ = godotenv.Load()

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

    // Dev auth endpoint: issues a JWT with sub + role for quick testing
    r.Post("/auth/dev", func(w http.ResponseWriter, r *http.Request) {
        secret := os.Getenv("JWT_SECRET")
        if secret == "" {
            http.Error(w, "JWT_SECRET not set", http.StatusPreconditionFailed)
            return
        }
        role := r.URL.Query().Get("role")
        if role == "" { role = "PARENT" }
        sub := r.URL.Query().Get("sub")
        if sub == "" { sub = "dev-user" }
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "sub":  sub,
            "role": role,
            "iat":  time.Now().Unix(),
        })
        s, err := token.SignedString([]byte(secret))
        if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
        w.Header().Set("Content-Type", "application/json")
        _, _ = w.Write([]byte("{\"token\":\"" + s + "\"}"))
    })

    // Dependencies
    dbClient, err := db.New(context.Background())
    if err != nil {
        log.Printf("dynamo client error (non-fatal): %v", err)
    }
    var appRepo repopkg.Repo
    if dbClient != nil {
        appRepo = repopkg.NewDynamoRepo(dbClient.Dynamo, os.Getenv("DYNAMO_TABLE_NAME"))
    }

    // GraphQL endpoint (gqlgen)
    gql := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Repo: appRepo}}))
    r.Method("POST", "/query", gql)
    r.Method("GET", "/query", gql) // allow GET for basic tests
    // GraphQL Playground (legacy) — keep available for reference
    r.Get("/play", func(w http.ResponseWriter, r *http.Request) {
        playground.Handler("GraphQL", "/query").ServeHTTP(w, r)
    })

    // Optional GraphiQL UI in dev: enable by setting ENABLE_GRAPHIQL=1
    if os.Getenv("ENABLE_GRAPHIQL") == "1" {
        r.Get("/graphiql", func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            _, _ = w.Write([]byte(graphiqlHTML))
        })
    }

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

    // Optional: ensure Dynamo table(s) when requested
    if os.Getenv("DYNAMO_AUTO_MIGRATE") == "1" {
        if dbClient != nil {
            if err := db.EnsureSingleTable(context.Background(), dbClient, os.Getenv("DYNAMO_TABLE_NAME")); err != nil {
                log.Printf("dynamo ensure table error: %v", err)
            } else {
                log.Printf("dynamo ensure table ok")
            }
        }
    }

    addr := ":8080"
    if v := os.Getenv("PORT"); v != "" {
        addr = ":" + v
    }

    server := &http.Server{Addr: addr, Handler: r, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second}
    log.Printf("GraphQL server listening on %s/query", addr)
    if os.Getenv("ENABLE_GRAPHIQL") == "1" {
        log.Printf("GraphiQL UI enabled at %s/graphiql", addr)
    }
    log.Fatal(server.ListenAndServe())
}

// Minimal GraphiQL HTML served in dev
const graphiqlHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <title>GraphiQL</title>
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphiql@2.7.1/graphiql.min.css" />
  <style>
    body { margin: 0; height: 100vh; }
    #graphiql { height: 100vh; }
    .topbar { position: fixed; top:0; left:0; right:0; height:44px; background: linear-gradient(90deg,#4f46e5,#06b6d4); color:#fff; display:flex; align-items:center; padding:0 12px; font: 14px/1.4 system-ui,-apple-system,Segoe UI,Roboto; z-index: 10; }
    .pad { height:44px; }
    .token-input { margin-left:auto; display:flex; gap:6px; align-items:center; }
    .token-input input { width: 360px; max-width: 50vw; padding: 6px 8px; border-radius: 6px; border: none; }
  </style>
</head>
<body>
  <div class="topbar">
    <strong>Chorequest GraphiQL</strong>
    <div class="token-input">
      <label for="auth">Authorization</label>
      <input id="auth" type="text" placeholder="Bearer <token> (optional)" />
    </div>
  </div>
  <div class="pad"></div>
  <div id="graphiql">Loading…</div>
  <script src="https://cdn.jsdelivr.net/npm/react@18/umd/react.production.min.js" crossorigin></script>
  <script src="https://cdn.jsdelivr.net/npm/react-dom@18/umd/react-dom.production.min.js" crossorigin></script>
  <script src="https://cdn.jsdelivr.net/npm/graphiql@2.7.1/graphiql.min.js" crossorigin></script>
  <script>
    const e = React.createElement;
    let extraHeaders = {};
    const input = document.getElementById('auth');
    input.addEventListener('input', () => {
      const v = input.value.trim();
      if (v) extraHeaders['Authorization'] = v; else delete extraHeaders['Authorization'];
    });
    const fetcher = GraphiQL.createFetcher({ url: '/query', headers: () => extraHeaders });
    const root = ReactDOM.createRoot(document.getElementById('graphiql'));
    root.render(e(GraphiQL, { fetcher }));
  </script>
</body>
</html>`
