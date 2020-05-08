package mongoke

import (
	"errors"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"
)

const TIMEOUT_FIND = 10

type findOneFieldConfig struct {
	collection string
	returnType *graphql.Object
}

func (mongoke *Mongoke) findOneField(conf findOneFieldConfig) *graphql.Field {
	if conf.collection == "" {
		panic(errors.New("missing collection name for " + conf.returnType.Name() + " findOneField"))
	}
	resolver := func(params graphql.ResolveParams) (interface{}, error) {
		args := params.Args
		opts := FindOneParams{
			Collection:  "users", // TODO collection based on config and current type name
			DatabaseUri: mongoke.databaseUri,
		}
		err := mapstructure.Decode(args, &opts)
		if err != nil {
			return nil, err
		}
		document, err := mongoke.databaseFunctions.FindOne(opts)
		if err != nil {
			return nil, err
		}
		// document, err := mongoke.database.findOne()
		// prettyPrint(args)
		return document, nil
	}
	whereArg, err := mongoke.getWhereArg(conf.returnType)
	if err != nil {
		panic(err)
	}
	return &graphql.Field{
		Type: conf.returnType,
		Args: graphql.FieldConfigArgument{
			"where": &graphql.ArgumentConfig{Type: whereArg},
		},
		Resolve: resolver,
	}
}

type findManyFieldConfig struct {
	collection string
	returnType *graphql.Object
}

func (mongoke *Mongoke) findManyField(conf findManyFieldConfig) *graphql.Field {
	resolver := func(params graphql.ResolveParams) (interface{}, error) {
		args := params.Args
		opts := FindManyParams{
			DatabaseUri: mongoke.databaseUri,
			Collection:  "users", // here i set the defaults
			Direction:   ASC,
			CursorField: "_id",
			Pagination:  paginationFromArgs(args),
		}
		err := mapstructure.Decode(args, &opts)
		if err != nil {
			return nil, err
		}
		document, err := mongoke.databaseFunctions.FindMany(
			opts,
		)
		if err != nil {
			return nil, err
		}
		// document, err := mongoke.database.findOne()
		// prettyPrint(args)
		return document, nil
	}
	whereArg, err := mongoke.getWhereArg(conf.returnType)
	if err != nil {
		panic(err)
	}
	connectionType, err := mongoke.getConnectionType(conf.returnType)
	if err != nil {
		panic(err)
	}
	return &graphql.Field{
		Type: connectionType,
		Args: graphql.FieldConfigArgument{
			"where":       &graphql.ArgumentConfig{Type: whereArg},
			"first":       &graphql.ArgumentConfig{Type: graphql.Int},
			"last":        &graphql.ArgumentConfig{Type: graphql.Int},
			"direction":   &graphql.ArgumentConfig{Type: directionEnum},
			"cursorField": &graphql.ArgumentConfig{Type: graphql.String}, // TODO make cursorField as the indexable fields enum
		},
		Resolve: resolver,
	}
}

func paginationFromArgs(args interface{}) Pagination {
	var pag Pagination
	err := mapstructure.Decode(args, &pag)
	if err != nil {
		fmt.Println(err)
		return Pagination{}
	}
	// prettyPrint(pag)
	return pag
}
