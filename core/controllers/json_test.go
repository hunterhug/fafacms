package controllers

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestActivateUser(t *testing.T) {
	req := new(ListFileAdminRequest)
	req.CreateTimeEnd = time.Now().Unix()

	raw, _ := json.MarshalIndent(req, " ", " ")
	fmt.Println(string(raw))
}
