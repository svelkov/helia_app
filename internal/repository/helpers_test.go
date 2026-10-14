package repository

import "helia/internal/infrastructure/db/querybuilder"

func queryFor(base string) *querybuilder.Builder { return querybuilder.New(base) }
