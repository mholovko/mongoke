package mongoke

import (
	"net/http"

	"github.com/graphql-go/handler"
)

type Config struct {
	schemaString string
}

func main(config Config) {
	schema, _ := generateSchema(config.schemaString)

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/", h)
	http.ListenAndServe(":8080", nil)
}
