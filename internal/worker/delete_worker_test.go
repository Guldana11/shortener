package worker

import (
	"context"
	"sync"
	"testing"
	"time"
)

type mockRepo struct {
	mu      sync.Mutex
	deleted map[string][]string
}

func newMockRepo() *mockRepo {
	return &mockRepo{deleted: make(map[string][]string)}
}

func (m *mockRepo) MarkAsDeleted(userID string, ids []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleted[userID] = append(m.deleted[userID], ids...)
	return nil
}

func (m *mockRepo) Create(originalURL string) (string, error)                { return "", nil }
func (m *mockRepo) CreateWithID(id, originalURL string)                      {}
func (m *mockRepo) Get(id string) (string, error)                            { return "", nil }
func (m *mockRepo) Ping(ctx context.Context) error                           { return nil }
func (m *mockRepo) CreateForUser(userID, originalURL string) (string, error) { return "", nil }
func (m *mockRepo) GetAllForUser(userID string) map[string]string            { return nil }
func (m *mockRepo) GetStats(ctx context.Context) (int, int, error)           { return 0, 0, nil }

func TestNewDeleteWorker(t *testing.T) {
	repo := newMockRepo()
	w := NewDeleteWorker(repo, 10)
	if w == nil {
		t.Fatal("NewDeleteWorker вернул nil")
	}
	if w.Repo != repo {
		t.Error("Repo не установлен")
	}
	w.Stop()
}

func TestEnqueueDeletion(t *testing.T) {
	repo := newMockRepo()
	w := NewDeleteWorker(repo, 10)

	w.EnqueueDeletion("user1", []string{"id1", "id2"})
	w.EnqueueDeletion("user2", []string{"id3"})
	w.Stop()

	repo.mu.Lock()
	defer repo.mu.Unlock()

	if len(repo.deleted["user1"]) != 2 {
		t.Errorf("ожидали 2 удалённых ID для user1, получили %d", len(repo.deleted["user1"]))
	}
	if len(repo.deleted["user2"]) != 1 {
		t.Errorf("ожидали 1 удалённый ID для user2, получили %d", len(repo.deleted["user2"]))
	}
}

func TestStopIdempotent(t *testing.T) {
	repo := newMockRepo()
	w := NewDeleteWorker(repo, 10)

	w.Stop()
	w.Stop() // повторный вызов не должен паниковать
}

func TestStopWaitsForTasks(t *testing.T) {
	repo := newMockRepo()
	w := NewDeleteWorker(repo, 100)

	for i := 0; i < 50; i++ {
		w.EnqueueDeletion("user1", []string{"id"})
	}

	done := make(chan struct{})
	go func() {
		w.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop не завершился за 5 секунд")
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(repo.deleted["user1"]) != 50 {
		t.Errorf("ожидали 50 удалённых ID, получили %d", len(repo.deleted["user1"]))
	}
}
