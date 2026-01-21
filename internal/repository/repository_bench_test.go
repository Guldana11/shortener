package repository

import (
	"strconv"
	"testing"
)

func BenchmarkCreate(b *testing.B) {
	repo := NewURLRepository("")
	url := "https://example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Create(url)
	}
}

func BenchmarkCreateWithID(b *testing.B) {
	repo := NewURLRepository("")
	url := "https://example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.CreateWithID("id"+strconv.Itoa(i), url)
	}
}

func BenchmarkCreateForUser(b *testing.B) {
	repo := NewURLRepository("")
	url := "https://example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.CreateForUser("user123", url)
	}
}

func BenchmarkGet(b *testing.B) {
	repo := NewURLRepository("")
	id, _ := repo.CreateForUser("user123", "https://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Get(id)
	}
}

func BenchmarkGetAllForUser(b *testing.B) {
	repo := NewURLRepository("")
	for i := 0; i < 10; i++ {
		_, _ = repo.CreateForUser("user123", "https://example.com/"+strconv.Itoa(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.GetAllForUser("user123")
	}
}

func BenchmarkMarkAsDeleted(b *testing.B) {
	repo := NewURLRepository("")
	ids := make([]string, 10)
	for i := 0; i < 10; i++ {
		id, _ := repo.CreateForUser("user123", "https://example.com/"+strconv.Itoa(i))
		ids[i] = id
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.MarkAsDeleted("user123", ids)
	}
}
