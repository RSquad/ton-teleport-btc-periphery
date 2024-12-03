//go:build ignore

package main

import (
	"log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	ex, err := entgql.NewExtension(
		entgql.WithWhereInputs(true),
		entgql.WithConfigPath("./gqlgen.yml"),
		entgql.WithSchemaGenerator(),
		entgql.WithSchemaPath("./ent.graphql"),
	)
	if err != nil {
		log.Fatalf("%v", err)
	}
	opts := []entc.Option{
		entc.Extensions(ex),
	}
	if err := entc.Generate("./internal/ent/schema", &gen.Config{
		Target:  "./internal/ent/generated",
		Package: "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated",
	}, opts...); err != nil {
		log.Fatalf("%v", err)
	}
}
