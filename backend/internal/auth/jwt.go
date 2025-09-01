package auth

import (
    "context"
    "net/http"
    "strings"

    "github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const subjectKey ctxKey = "sub"
const roleKey ctxKey = "role"

// JWTMiddleware validates Bearer tokens using the provided HMAC secret.
// If secret is empty, the middleware is a no-op (allows all requests).
func JWTMiddleware(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        if secret == "" {
            return next
        }
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authz := r.Header.Get("Authorization")
            if authz == "" || !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
                next.ServeHTTP(w, r)
                return
            }
            tokenString := strings.TrimSpace(authz[len("Bearer "):])
            token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
                if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, jwt.ErrInvalidKeyType
                }
                return []byte(secret), nil
            })
            if err == nil && token != nil && token.Valid {
                if claims, ok := token.Claims.(jwt.MapClaims); ok {
                    if sub, ok := claims["sub"].(string); ok {
                        ctx := context.WithValue(r.Context(), subjectKey, sub)
                        // attach role if present
                        if role, ok := claims["role"].(string); ok {
                            ctx = context.WithValue(ctx, roleKey, role)
                        }
                        r = r.WithContext(ctx)
                    }
                }
            }
            next.ServeHTTP(w, r)
        })
    }
}

func SubjectFromContext(ctx context.Context) string {
    v, _ := ctx.Value(subjectKey).(string)
    return v
}

func RoleFromContext(ctx context.Context) string {
    v, _ := ctx.Value(roleKey).(string)
    return v
}
