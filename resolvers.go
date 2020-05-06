package mongoke

import (
	"github.com/graphql-go/graphql"
)

type findOneResolverConfig struct {
	collection   string
	database     string
	resolvedType *graphql.Object
}

func findOneResolver(conf findOneResolverConfig) *graphql.Field {
	// TODO create the where argument based on the object fields

	resolver := func(params graphql.ResolveParams) (interface{}, error) {
		args := params.Args
		// TODO get item from database
		// check authorization guards
		// if interface or union set the right __typeName
		prettyPrint(args)
		return "world", nil
	}
	return &graphql.Field{
		Type:    conf.resolvedType,
		Args:    graphql.FieldConfigArgument{"where": &graphql.ArgumentConfig{Type: graphql.String}},
		Resolve: resolver,
	}
}
