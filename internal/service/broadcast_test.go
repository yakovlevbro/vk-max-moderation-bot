package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
)

func TestModerationService_GetBroadcastSelections(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name      string
		userID    int64
		setupMock func() *MockBroadcastSelectionRepository
		want      []int64
		wantErr   bool
	}{
		{
			name:   "Returns selections",
			userID: 1,
			setupMock: func() *MockBroadcastSelectionRepository {
				return &MockBroadcastSelectionRepository{
					GetSelectionsFunc: func(userID int64) ([]int64, error) {
						return []int64{100, 200, 300}, nil
					},
				}
			},
			want:    []int64{100, 200, 300},
			wantErr: false,
		},
		{
			name:   "Returns empty list",
			userID: 2,
			setupMock: func() *MockBroadcastSelectionRepository {
				return &MockBroadcastSelectionRepository{
					GetSelectionsFunc: func(userID int64) ([]int64, error) {
						return nil, nil
					},
				}
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "Repo error",
			userID: 3,
			setupMock: func() *MockBroadcastSelectionRepository {
				return &MockBroadcastSelectionRepository{
					GetSelectionsFunc: func(userID int64) ([]int64, error) {
						return nil, errors.New("db error")
					},
				}
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broadcastRepo := tt.setupMock()
			svc := NewModerationService(logger, nil, nil, nil, nil, nil, &MockViolationRepository{}, broadcastRepo, nil, nil)

			got, err := svc.GetBroadcastSelections(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBroadcastSelections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("GetBroadcastSelections() len = %d, want %d", len(got), len(tt.want))
			}
		})
	}
}

func TestModerationService_ToggleBroadcastSelection(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name        string
		userID      int64
		chatID      int64
		setupMock   func() *MockBroadcastSelectionRepository
		wantSelected bool
		wantErr     bool
	}{
		{
			name:   "Toggle on",
			userID: 1,
			chatID: 100,
			setupMock: func() *MockBroadcastSelectionRepository {
				return &MockBroadcastSelectionRepository{
					ToggleFunc: func(userID, chatID int64) (bool, error) {
						return true, nil
					},
				}
			},
			wantSelected: true,
			wantErr:      false,
		},
		{
			name:   "Toggle off",
			userID: 1,
			chatID: 100,
			setupMock: func() *MockBroadcastSelectionRepository {
				return &MockBroadcastSelectionRepository{
					ToggleFunc: func(userID, chatID int64) (bool, error) {
						return false, nil
					},
				}
			},
			wantSelected: false,
			wantErr:      false,
		},
		{
			name:   "Repo error",
			userID: 1,
			chatID: 100,
			setupMock: func() *MockBroadcastSelectionRepository {
				return &MockBroadcastSelectionRepository{
					ToggleFunc: func(userID, chatID int64) (bool, error) {
						return false, errors.New("db error")
					},
				}
			},
			wantSelected: false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broadcastRepo := tt.setupMock()
			svc := NewModerationService(logger, nil, nil, nil, nil, nil, &MockViolationRepository{}, broadcastRepo, nil, nil)

			got, err := svc.ToggleBroadcastSelection(context.Background(), tt.userID, tt.chatID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToggleBroadcastSelection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantSelected {
				t.Errorf("ToggleBroadcastSelection() = %v, want %v", got, tt.wantSelected)
			}
		})
	}
}

func TestModerationService_ClearBroadcastSelections(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name      string
		userID    int64
		setupMock func() *MockBroadcastSelectionRepository
		wantErr   bool
	}{
		{
			name:   "Success",
			userID: 1,
			setupMock: func() *MockBroadcastSelectionRepository {
				return &MockBroadcastSelectionRepository{
					ClearFunc: func(userID int64) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:   "Repo error",
			userID: 1,
			setupMock: func() *MockBroadcastSelectionRepository {
				return &MockBroadcastSelectionRepository{
					ClearFunc: func(userID int64) error {
						return errors.New("db error")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broadcastRepo := tt.setupMock()
			svc := NewModerationService(logger, nil, nil, nil, nil, nil, &MockViolationRepository{}, broadcastRepo, nil, nil)

			err := svc.ClearBroadcastSelections(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClearBroadcastSelections() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestModerationService_SelectAllChatsForBroadcast(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name       string
		userID     int64
		setupMocks func() (*MockChatAdminRepository, *MockBroadcastSelectionRepository)
		wantErr    bool
	}{
		{
			name:   "Success - selects all managed chats",
			userID: 1,
			setupMocks: func() (*MockChatAdminRepository, *MockBroadcastSelectionRepository) {
				adminRepo := &MockChatAdminRepository{
					GetManagedChatsFunc: func(userID int64) ([]int64, error) {
						return []int64{100, 200, 300}, nil
					},
				}
				broadcastRepo := &MockBroadcastSelectionRepository{
					SelectAllFunc: func(userID int64, chatIDs []int64) error {
						if len(chatIDs) != 3 {
							t.Errorf("expected 3 chatIDs, got %d", len(chatIDs))
						}
						return nil
					},
				}
				return adminRepo, broadcastRepo
			},
			wantErr: false,
		},
		{
			name:   "Admin repo error",
			userID: 1,
			setupMocks: func() (*MockChatAdminRepository, *MockBroadcastSelectionRepository) {
				adminRepo := &MockChatAdminRepository{
					GetManagedChatsFunc: func(userID int64) ([]int64, error) {
						return nil, errors.New("db error")
					},
				}
				return adminRepo, &MockBroadcastSelectionRepository{}
			},
			wantErr: true,
		},
		{
			name:   "Broadcast repo error",
			userID: 1,
			setupMocks: func() (*MockChatAdminRepository, *MockBroadcastSelectionRepository) {
				adminRepo := &MockChatAdminRepository{
					GetManagedChatsFunc: func(userID int64) ([]int64, error) {
						return []int64{100}, nil
					},
				}
				broadcastRepo := &MockBroadcastSelectionRepository{
					SelectAllFunc: func(userID int64, chatIDs []int64) error {
						return errors.New("db error")
					},
				}
				return adminRepo, broadcastRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adminRepo, broadcastRepo := tt.setupMocks()
			svc := NewModerationService(logger, nil, adminRepo, nil, nil, nil, &MockViolationRepository{}, broadcastRepo, nil, nil)

			err := svc.SelectAllChatsForBroadcast(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectAllChatsForBroadcast() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestModerationService_SaveBroadcastDraft(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name      string
		userID    int64
		text      string
		setupMock func() *MockBroadcastDraftRepository
		wantErr   bool
	}{
		{
			name:   "Success",
			userID: 1,
			text:   "Hello chats!",
			setupMock: func() *MockBroadcastDraftRepository {
				return &MockBroadcastDraftRepository{
					SaveFunc: func(userID int64, text string) error {
						if text != "Hello chats!" {
							t.Errorf("expected text 'Hello chats!', got '%s'", text)
						}
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:   "Repo error",
			userID: 1,
			text:   "Hello",
			setupMock: func() *MockBroadcastDraftRepository {
				return &MockBroadcastDraftRepository{
					SaveFunc: func(userID int64, text string) error {
						return errors.New("db error")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draftRepo := tt.setupMock()
			svc := NewModerationService(logger, nil, nil, nil, nil, nil, &MockViolationRepository{}, nil, draftRepo, nil)

			err := svc.SaveBroadcastDraft(context.Background(), tt.userID, tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveBroadcastDraft() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestModerationService_GetBroadcastDraft(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name      string
		userID    int64
		setupMock func() *MockBroadcastDraftRepository
		wantText  string
		wantErr   bool
	}{
		{
			name:   "Returns draft",
			userID: 1,
			setupMock: func() *MockBroadcastDraftRepository {
				return &MockBroadcastDraftRepository{
					GetFunc: func(userID int64) (string, error) {
						return "Saved draft text", nil
					},
				}
			},
			wantText: "Saved draft text",
			wantErr:  false,
		},
		{
			name:   "Repo error",
			userID: 1,
			setupMock: func() *MockBroadcastDraftRepository {
				return &MockBroadcastDraftRepository{
					GetFunc: func(userID int64) (string, error) {
						return "", errors.New("not found")
					},
				}
			},
			wantText: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draftRepo := tt.setupMock()
			svc := NewModerationService(logger, nil, nil, nil, nil, nil, &MockViolationRepository{}, nil, draftRepo, nil)

			got, err := svc.GetBroadcastDraft(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBroadcastDraft() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantText {
				t.Errorf("GetBroadcastDraft() = %v, want %v", got, tt.wantText)
			}
		})
	}
}

func TestModerationService_DeleteBroadcastDraft(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name      string
		userID    int64
		setupMock func() *MockBroadcastDraftRepository
		wantErr   bool
	}{
		{
			name:   "Success",
			userID: 1,
			setupMock: func() *MockBroadcastDraftRepository {
				return &MockBroadcastDraftRepository{
					DeleteFunc: func(userID int64) error { return nil },
				}
			},
			wantErr: false,
		},
		{
			name:   "Repo error",
			userID: 1,
			setupMock: func() *MockBroadcastDraftRepository {
				return &MockBroadcastDraftRepository{
					DeleteFunc: func(userID int64) error { return errors.New("db error") },
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draftRepo := tt.setupMock()
			svc := NewModerationService(logger, nil, nil, nil, nil, nil, &MockViolationRepository{}, nil, draftRepo, nil)

			err := svc.DeleteBroadcastDraft(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteBroadcastDraft() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestModerationService_SendBroadcast(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name          string
		userID        int64
		chatIDs       []int64
		text          string
		managedChats  []int64
		adminRepoErr  error
		wantSent      int
		wantFailed    int
		wantErr       bool
	}{
		{
			name:         "Bot not initialized",
			userID:       1,
			chatIDs:      []int64{100},
			text:         "hello",
			managedChats: []int64{100},
			wantSent:     0,
			wantFailed:   0,
			wantErr:      true,
		},
		{
			name:         "Admin repo error",
			userID:       1,
			chatIDs:      []int64{100},
			text:         "hello",
			adminRepoErr: errors.New("db error"),
			wantSent:     0,
			wantFailed:   0,
			wantErr:      true,
		},
		{
			name:         "All chats not managed - all failed",
			userID:       1,
			chatIDs:      []int64{100, 200},
			text:         "hello",
			managedChats: []int64{999},
			wantSent:     0,
			wantFailed:   2,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adminRepo := &MockChatAdminRepository{
				GetManagedChatsFunc: func(userID int64) ([]int64, error) {
					if tt.adminRepoErr != nil {
						return nil, tt.adminRepoErr
					}
					return tt.managedChats, nil
				},
			}

			var svc Service
			if tt.name == "Bot not initialized" {
				svc = NewModerationService(logger, nil, adminRepo, nil, nil, nil, &MockViolationRepository{}, nil, nil, nil)
			} else {
				svc = NewModerationService(logger, nil, adminRepo, nil, nil, nil, &MockViolationRepository{}, nil, nil, nil)
			}

			sent, failed, err := svc.SendBroadcast(context.Background(), tt.userID, tt.chatIDs, tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendBroadcast() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if sent != tt.wantSent {
					t.Errorf("SendBroadcast() sent = %d, want %d", sent, tt.wantSent)
				}
				if failed != tt.wantFailed {
					t.Errorf("SendBroadcast() failed = %d, want %d", failed, tt.wantFailed)
				}
			}
		})
	}
}
