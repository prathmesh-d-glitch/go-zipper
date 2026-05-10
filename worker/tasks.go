package worker

import (
	"errors"
	"fmt"
)

type TaskType string

const (
	TaskCompress   TaskType = "compress"
	TaskDecompress TaskType = "decompress"
)

type Task struct {
	ID        string
	Type      TaskType
	InputPath string
	InputData []byte
}

func (t *Task) Validate() error {
	if t.ID == "" {
		return errors.New("worker: task ID is required")
	}

	switch t.Type {
	case TaskCompress:
		if t.InputPath == "" {
			return fmt.Errorf("worker: compress task %q requires InputPath", t.ID)
		}
	case TaskDecompress:
		if len(t.InputData) == 0 {
			return fmt.Errorf("worker: decompress task %q requires InputData", t.ID)
		}
	default:
		return fmt.Errorf("worker: unknown task type %q for task %q", t.Type, t.ID)
	}

	return nil
}
