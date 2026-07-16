package crud_test

import (
	"testing"
)

func BenchmarkExecutingSQL(b *testing.B) {
	for b.Loop() {
		_, err := DB.Client.Exec("SHOW TABLES LIKE 'shouldnotexist'")
		if err != nil {
			panic(err)
		}
	}
}
