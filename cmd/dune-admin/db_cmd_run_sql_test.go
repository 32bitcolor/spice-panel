package main

import (
	"strings"
	"testing"
)

func TestFormatSQLRow(t *testing.T) {
	t.Parallel()

	row := formatSQLRow([]any{int64(7), "name", nil})
	if row != "7 │ name │ <nil>" {
		t.Fatalf("unexpected row format: %q", row)
	}
}

func TestFormatSQLStringRows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   [][]any
		want [][]string
	}{
		{
			name: "integers and strings convert to string representation",
			in:   [][]any{{int64(1), "alpha"}, {int64(2), "beta"}},
			want: [][]string{{"1", "alpha"}, {"2", "beta"}},
		},
		{
			name: "nil becomes <nil>",
			in:   [][]any{{nil, "x"}},
			want: [][]string{{"<nil>", "x"}},
		},
		{
			name: "empty input returns empty slice",
			in:   [][]any{},
			want: [][]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatSQLStringRows(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("len=%d want=%d", len(got), len(tt.want))
			}
			for i, row := range got {
				if len(row) != len(tt.want[i]) {
					t.Fatalf("row %d: len=%d want=%d", i, len(row), len(tt.want[i]))
				}
				for j, cell := range row {
					if cell != tt.want[i][j] {
						t.Fatalf("[%d][%d]: got %q want %q", i, j, cell, tt.want[i][j])
					}
				}
			}
		})
	}
}

func TestBuildSQLResult(t *testing.T) {
	t.Parallel()

	result := buildSQLResult(
		[]string{"id", "name"},
		[][]any{
			{int64(1), "alpha"},
			{int64(2), "beta"},
		},
		false,
	)
	if !strings.Contains(result, "id │ name\n") {
		t.Fatalf("expected header line in result: %q", result)
	}
	if !strings.Contains(result, "1 │ alpha\n") || !strings.Contains(result, "2 │ beta\n") {
		t.Fatalf("expected row lines in result: %q", result)
	}
	if strings.Contains(result, "limited to 200 rows") {
		t.Fatalf("did not expect truncation marker in non-truncated result")
	}

	truncated := buildSQLResult([]string{"id"}, [][]any{{1}}, true)
	if !strings.Contains(truncated, "… (limited to 200 rows)\n") {
		t.Fatalf("expected truncation marker in truncated result: %q", truncated)
	}
}
