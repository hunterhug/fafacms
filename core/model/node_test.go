package model

import "testing"

func TestContentNode_Get(t *testing.T) {
	setupTestDB(t)

	n := new(ContentNode)
	n.Id = 2
	n.Get()
}
