package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.
import (
    "sync"
    repopkg "chorequest/backend/internal/repo"
)

type Resolver struct{
    mu sync.Mutex
    Repo repopkg.Repo
}
