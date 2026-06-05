package core

import "testing"

func TestIsInsertStatement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		sql  string
		want bool
	}{
		{
			name: "plain insert",
			sql:  " INSERT INTO users(name) VALUES ('Ada')",
			want: true,
		},
		{
			name: "replace",
			sql:  "REPLACE INTO users(id, name) VALUES (1, 'Ada')",
			want: true,
		},
		{
			name: "leading comments",
			sql:  "-- INSERT is only in a comment\n/* another comment */ INSERT INTO users(name) VALUES ('Ada')",
			want: true,
		},
		{
			name: "cte insert with quoted text",
			sql:  "WITH incoming AS (SELECT 'not ) INSERT' AS name) INSERT INTO users(name) SELECT name FROM incoming",
			want: true,
		},
		{
			name: "cte insert with materialized hint",
			sql:  "WITH incoming AS NOT MATERIALIZED (SELECT 'Ada' AS name) INSERT INTO users(name) SELECT name FROM incoming",
			want: true,
		},
		{
			name: "cte update",
			sql:  "WITH selected(id) AS (SELECT 1) UPDATE users SET name = 'Ada' WHERE id = (SELECT id FROM selected)",
			want: false,
		},
		{
			name: "commented insert before update",
			sql:  "/* INSERT */ UPDATE users SET name = 'INSERT'",
			want: false,
		},
		{
			name: "cte select",
			sql:  "WITH selected(id) AS (SELECT 1) SELECT id FROM selected",
			want: false,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isInsertStatement(tt.sql); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
