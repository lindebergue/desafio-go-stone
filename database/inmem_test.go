package database

import "testing"

func TestInMemDB(t *testing.T) {
	runDBTests(t, NewInMemDB())
}
