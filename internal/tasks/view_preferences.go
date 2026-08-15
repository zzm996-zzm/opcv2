package tasks

import (
	"context"
	"encoding/json"
	"strings"
)

type TaskViewPreferenceRepository interface {
	GetTaskViewPreference(context.Context, int64, string) (TaskViewPreference, error)
	SaveTaskViewPreference(context.Context, int64, string, []string) (TaskViewPreference, error)
}

func (s *Service) taskViewPreferenceRepository() (TaskViewPreferenceRepository, error) {
	repository, ok := s.repository.(TaskViewPreferenceRepository)
	if !ok {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func (s *Service) GetTaskViewPreference(ctx context.Context, userID int64, view string) (TaskViewPreference, error) {
	repository, err := s.taskViewPreferenceRepository()
	if err != nil {
		return TaskViewPreference{}, err
	}
	if userID <= 0 || normalizeTaskView(view) != TaskViewList {
		return TaskViewPreference{}, ErrInvalidTaskViewPreference
	}
	return repository.GetTaskViewPreference(ctx, userID, TaskViewList)
}

func (s *Service) SaveTaskViewPreference(ctx context.Context, input UpdateTaskViewPreferenceInput) (TaskViewPreference, error) {
	repository, err := s.taskViewPreferenceRepository()
	if err != nil {
		return TaskViewPreference{}, err
	}
	if input.UserID <= 0 || normalizeTaskView(input.View) != TaskViewList || !validTaskListColumns(input.Columns) {
		return TaskViewPreference{}, ErrInvalidTaskViewPreference
	}
	return repository.SaveTaskViewPreference(ctx, input.UserID, TaskViewList, input.Columns)
}

func normalizeTaskView(view string) string {
	view = strings.TrimSpace(view)
	if view == "" {
		return TaskViewList
	}
	return view
}

func validTaskListColumns(columns []string) bool {
	if len(columns) == 0 || len(columns) > len(taskListAllowedColumns) {
		return false
	}
	allowed := make(map[string]struct{}, len(taskListAllowedColumns))
	for _, column := range taskListAllowedColumns {
		allowed[column] = struct{}{}
	}
	seen := make(map[string]struct{}, len(columns))
	for _, column := range columns {
		column = strings.TrimSpace(column)
		if _, ok := allowed[column]; !ok {
			return false
		}
		if _, ok := seen[column]; ok {
			return false
		}
		seen[column] = struct{}{}
	}
	_, hasTitle := seen["title"]
	return hasTitle
}

func marshalTaskListColumns(columns []string) ([]byte, error) {
	return json.Marshal(columns)
}
