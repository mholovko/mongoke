package mongoke

import (
	"fmt"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"
	"github.com/remorses/mongoke/src/testutil"
	"github.com/stretchr/testify/require"
)

var config = Config{
	Schema: `
	type User {
		name: String
		age: Int
	}
	`,
	DatabaseUri: testutil.MONGODB_URI,
	Types: map[string]*TypeConfig{
		"User": {Collection: "users"},
	},
}

func TestQueryArgs(t *testing.T) {
	databaseMock := &DatabaseInterfaceMock{
		FindManyFunc: func(p FindManyParams) ([]Map, error) {
			return nil, nil
		},
		FindOneFunc: func(p FindOneParams) (interface{}, error) {
			return nil, nil
		},
	}
	schema, err := MakeMongokeSchema(config, databaseMock)
	if err != nil {
		t.Error(err)
	}
	t.Run("introspect schema", func(t *testing.T) {
		if err != nil {
			t.Error(err)
		}
		testutil.QuerySchema(t, schema, testutil.IntrospectionQuery)
	})
	t.Run("findOne query without args", func(t *testing.T) {
		query := `
		{
			User {
				name
				age
			}
		}
		`
		testutil.QuerySchema(t, schema, query)
		calls := len(databaseMock.FindOneCalls())
		require.Equal(t, 1, calls)
		where := databaseMock.FindOneCalls()[calls-1].P.Where
		// require.Equal(t, nil, where)
		t.Log(where)
	})
	t.Run("findOne query with eq", func(t *testing.T) {
		databaseMock.calls.FindOne = nil
		query := `
		{
			User(where: {name: {eq: "xxx"}}) {
				name
				age
			}
		}
		`
		testutil.QuerySchema(t, schema, query)
		calls := len(databaseMock.FindOneCalls())
		require.Equal(t, 1, calls)
		where := databaseMock.FindOneCalls()[0].P.Where
		t.Log(pretty(where))
		require.Equal(t, "xxx", where["name"].Eq)
	})
	t.Run("findMany query with first, after, where", func(t *testing.T) {
		databaseMock.calls.FindMany = nil
		query := `
		{
			UserNodes(first: 10, after: "xxx", where: {name: {eq: "xxx"}}) {
				nodes {
					name
				}
			}
		}
		`
		testutil.QuerySchema(t, schema, query)
		calls := len(databaseMock.calls.FindMany)
		require.Equal(t, 1, calls)
		p := databaseMock.calls.FindMany[0].P
		t.Log("params", pretty(p))
		// + 1 because we need to know hasNextPage
		require.Equal(t, 10+1, p.Pagination.First)
		require.Equal(t, "xxx", p.Pagination.After)
	})
}

func TestQueryReturnValues(t *testing.T) {
	type user struct {
		Name string `json:name`
		Age  int    `json:age`
	}

	type userStruct struct {
		Name string
		Age  int
	}

	var exampleUsers = []Map{
		{"name": "01", "age": 1},
		{"name": "02", "age": 2},
		{"name": "03", "age": 3},
	}
	exampleUser := exampleUsers[0]
	databaseMock := &DatabaseInterfaceMock{
		FindManyFunc: func(p FindManyParams) ([]Map, error) {
			return exampleUsers, nil
		},
		FindOneFunc: func(p FindOneParams) (interface{}, error) {
			return exampleUser, nil
		},
	}
	schema, err := MakeMongokeSchema(config, databaseMock)
	if err != nil {
		t.Error(err)
	}

	t.Run("findOne query without args", func(t *testing.T) {
		query := `
		{
			User {
				name
				age
			}
		}
		`
		type Res struct {
			User userStruct
		}
		res := testutil.QuerySchema(t, schema, query)
		var x Res
		mapstructure.Decode(res, &x)
		require.Equal(t, exampleUser["name"], x.User.Name)
	})

	t.Run("findOne query with permissions false", func(t *testing.T) {
		config := Config{
			Schema: `
			type User {
				name: String
				age: Int
			}
			`,
			DatabaseUri: testutil.MONGODB_URI,
			Types: map[string]*TypeConfig{
				"User": {
					Collection: "users",
					Permissions: []AuthGuard{
						{
							Expression: "false",
						},
					},
				},
			},
		}
		schema, err := MakeMongokeSchema(config, databaseMock)
		if err != nil {
			t.Error(err)
		}
		query := `
		{
			User {
				name
				age
			}
		}
		`
		err = testutil.QuerySchemaShouldFail(t, schema, query)
		fmt.Println(err)
		// require.Equal(t, err, "")
	})

	t.Run("findMany query without args", func(t *testing.T) {
		query := `
		{
			UserNodes {
				nodes {
					name
					age
				}
			}
		}
		`
		type Res struct {
			UserNodes struct {
				Nodes []userStruct
			}
		}
		res := testutil.QuerySchema(t, schema, query)
		var x Res
		mapstructure.Decode(res, &x)
		require.Equal(t, exampleUsers[0]["name"], x.UserNodes.Nodes[0].Name)
		require.Equal(t, exampleUsers[0]["age"], x.UserNodes.Nodes[0].Age)
	})

	t.Run("findMany query with permissions false", func(t *testing.T) {
		config := Config{
			Schema: `
			type User {
				name: String
				age: Int
			}
			`,
			DatabaseUri: testutil.MONGODB_URI,
			Types: map[string]*TypeConfig{
				"User": {
					Collection: "users",
					Permissions: []AuthGuard{
						{
							Expression: "false",
						},
					},
				},
			},
		}
		schema, err := MakeMongokeSchema(config, databaseMock)
		if err != nil {
			t.Error(err)
		}
		query := `
		{
			UserNodes {
				nodes {
					name
					age
				}
			}
		}
		`
		type Res struct {
			UserNodes struct {
				Nodes []userStruct
			}
		}
		res := testutil.QuerySchema(t, schema, query)
		t.Log(pretty(res))
		var x Res
		mapstructure.Decode(res, &x)
		require.Equal(t, 0, len(x.UserNodes.Nodes))
	})
}

func TestExtractClaims(t *testing.T) {
	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1ODkwMzI0MjksImV4cCI6MTYyMDU2ODQyOSwiYXVkIjoid3d3LmV4YW1wbGUuY29tIiwic3ViIjoianJvY2tldEBleGFtcGxlLmNvbSIsIkdpdmVuTmFtZSI6IkpvaG5ueSIsIlN1cm5hbWUiOiJSb2NrZXQiLCJFbWFpbCI6Impyb2NrZXRAZXhhbXBsZS5jb20iLCJSb2xlIjpbIk1hbmFnZXIiLCJQcm9qZWN0IEFkbWluaXN0cmF0b3IiXX0.Qt_BmT2lADjJSwKhxCureJED-RDoDDrUF2bHnYGqfOo"
	secret := "qwertyuiopasdfghjklzxcvbnm123456"
	claims, err := extractClaims(token, secret)
	if err != nil {
		t.Error(err)
	}
	t.Log(pretty(claims))

	require.Equal(t, "jrocket@example.com", claims["Email"])
}
func TestGetJwt(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		res := getJwt(graphql.ResolveParams{
			Info: graphql.ResolveInfo{
				RootValue: Map{
					"jwt": jwt.MapClaims{
						"email": "email",
					},
				},
			},
		})
		t.Log(pretty(res))
		require.Equal(t, "email", res["email"])
	})
	t.Run("jwt made with extractClaims", func(t *testing.T) {
		token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1ODkwMzI0MjksImV4cCI6MTYyMDU2ODQyOSwiYXVkIjoid3d3LmV4YW1wbGUuY29tIiwic3ViIjoianJvY2tldEBleGFtcGxlLmNvbSIsIkdpdmVuTmFtZSI6IkpvaG5ueSIsIlN1cm5hbWUiOiJSb2NrZXQiLCJFbWFpbCI6Impyb2NrZXRAZXhhbXBsZS5jb20iLCJSb2xlIjpbIk1hbmFnZXIiLCJQcm9qZWN0IEFkbWluaXN0cmF0b3IiXX0.Qt_BmT2lADjJSwKhxCureJED-RDoDDrUF2bHnYGqfOo"
		secret := "qwertyuiopasdfghjklzxcvbnm123456"
		claims, err := extractClaims(token, secret)
		if err != nil {
			t.Error(err)
		}
		res := getJwt(graphql.ResolveParams{
			Info: graphql.ResolveInfo{
				RootValue: Map{
					"jwt": claims,
				},
			},
		})
		t.Log(pretty(res))
		require.Equal(t, "jrocket@example.com", res["Email"])
	})

}
