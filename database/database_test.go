package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeathlyHand(t *testing.T) {
	require := require.New(t)

	db, err := Open("postgres",
		"postgres://ladybug:something_stupid@localhost/ladybug")
	require.NoError(err)
	require.NoError(db.Ping())

	ss := func(sp *string) string {
		if sp == nil {
			return "<NULL>"
		}
		return *sp
	}

	sb := func(sb *bool) string {
		if sb == nil {
			return "<NULL>"
		}
		if *sb {
			return "true"
		}
		return "false"
	}

	rows, err := db.Query("SELECT * FROM pg_catalog.pg_tables")
	require.NoError(err)
	defer rows.Close()
	fmt.Println("schemaname", "tablename", "tableowner", "tablespace", "hasindexes", "hasrules", "hastriggers", "rowsecurity")
	for rows.Next() {
		var schemaname *string
		var tablename *string
		var tableowner *string
		var tablespace *string
		var hasindexes *bool
		var hasrules *bool
		var hastriggers *bool
		var rowsecurity *bool

		err = rows.Scan(&schemaname, &tablename, &tableowner, &tablespace, &hasindexes, &hasrules, &hastriggers, &rowsecurity)
		require.NoError(err)
		fmt.Println(ss(schemaname), ss(tablename), ss(tableowner), ss(tablespace), sb(hasindexes), sb(hasrules), sb(hastriggers), sb(rowsecurity))
	}
	require.NoError(rows.Err())

	buyer, err := db.Create_Buyer(context.Background(),
		Buyer_Id("ID"),
		Buyer_FirstName("FIRST"),
		Buyer_LastName("LAST"),
	)
	require.NoError(err)

	fmt.Println(buyer)
}
