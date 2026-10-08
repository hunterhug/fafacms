package model

import (
	"testing"
)

func TestGroup_GetById(t *testing.T) {
	setupTestDB(t)

	g := new(Group)
	g.Id = 2
	FaFaRdb.Client.ID(g.Id).Get(g)
}
