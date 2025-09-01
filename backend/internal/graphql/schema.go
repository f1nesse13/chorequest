package graphql

import (
    "net/http"

    gql "github.com/graphql-go/graphql"
    "github.com/graphql-go/handler"
)

// NewHandler constructs an HTTP handler serving a simple GraphQL schema.
func NewHandler() http.Handler {
    // Define types
    userType := gql.NewObject(gql.ObjectConfig{
        Name: "User",
        Fields: gql.Fields{
            "id":   &gql.Field{Type: gql.NewNonNull(gql.String)},
            "name": &gql.Field{Type: gql.NewNonNull(gql.String)},
        },
    })

    todoType := gql.NewObject(gql.ObjectConfig{
        Name: "Todo",
        Fields: gql.Fields{
            "id":   &gql.Field{Type: gql.NewNonNull(gql.String)},
            "text": &gql.Field{Type: gql.NewNonNull(gql.String)},
            "done": &gql.Field{Type: gql.NewNonNull(gql.Boolean)},
            "user": &gql.Field{Type: gql.NewNonNull(userType)},
        },
    })

    // In-memory sample data
    todos := []map[string]any{
        {"id": "1", "text": "Try GraphQL", "done": false, "user": map[string]any{"id": "u1", "name": "Ada"}},
    }

    // Root query
    queryType := gql.NewObject(gql.ObjectConfig{
        Name: "Query",
        Fields: gql.Fields{
            "health": &gql.Field{
                Type: gql.NewNonNull(gql.String),
                Resolve: func(p gql.ResolveParams) (any, error) {
                    return "ok", nil
                },
            },
            "todos": &gql.Field{
                Type: gql.NewList(gql.NewNonNull(todoType)),
                Resolve: func(p gql.ResolveParams) (any, error) {
                    return todos, nil
                },
            },
        },
    })

    // Mutations
    mutationType := gql.NewObject(gql.ObjectConfig{
        Name: "Mutation",
        Fields: gql.Fields{
            "createTodo": &gql.Field{
                Type: todoType,
                Args: gql.FieldConfigArgument{
                    "text":  &gql.ArgumentConfig{Type: gql.NewNonNull(gql.String)},
                    "userId": &gql.ArgumentConfig{Type: gql.NewNonNull(gql.String)},
                },
                Resolve: func(p gql.ResolveParams) (any, error) {
                    id := "todo_" + p.Args["text"].(string)
                    item := map[string]any{
                        "id":   id,
                        "text": p.Args["text"],
                        "done": false,
                        "user": map[string]any{"id": p.Args["userId"], "name": "User"},
                    }
                    todos = append(todos, item)
                    return item, nil
                },
            },
        },
    })

    schema, _ := gql.NewSchema(gql.SchemaConfig{Query: queryType, Mutation: mutationType})

    return handler.New(&handler.Config{
        Schema:   &schema,
        Pretty:   true,
        GraphiQL: true,
    })
}

