package worker

import "github.com/Guldana11/shortener/internal/repository"

type DeleteTask struct {
	UserID string
	IDs    []string
}

type DeleteWorker struct {
	Repo      repository.Repository
	TaskQueue chan DeleteTask
}

func NewDeleteWorker(repo repository.Repository, bufferSize int) *DeleteWorker {
	w := &DeleteWorker{
		Repo:      repo,
		TaskQueue: make(chan DeleteTask, bufferSize),
	}

	go w.run()
	return w
}

func (w *DeleteWorker) run() {
	for task := range w.TaskQueue {
		_ = w.Repo.MarkAsDeleted(task.UserID, task.IDs)
	}
}

func (w *DeleteWorker) EnqueueDeletion(userID string, ids []string) {
	w.TaskQueue <- DeleteTask{UserID: userID, IDs: ids}
}
