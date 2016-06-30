package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/kr/pretty"
)

var hierarchical = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "hierarchical",
	Description: "everything.",
	Serialize: func(value interface{}) interface{} {
		pretty.Log("serialize\n", value)
		return 1
	},
	ParseValue: func(value interface{}) interface{} {
		pretty.Log("parsevalue\n", value)
		return 2
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		pretty.Log("parseliteral\n", valueAST)
		return 3
	},
})

func MakeSchema() (graphql.Schema, error) {
	return graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "RootQuery",
			Fields: graphql.Fields{
				"hello": &graphql.Field{
					Type: hierarchical,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return map[string]interface{}{
							"x": 23,
							"y": "lskdnfs",
						}, nil
					},
				},
			},
		}),
	})
}
