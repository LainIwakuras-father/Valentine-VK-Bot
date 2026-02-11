package handlers

import (
	"sync"
)

// UserState хранит состояние пользователя
type UserState struct {
	CurrentStep string
	Data        map[string]interface{}
}

// StateManager управляет состояниями пользователей
type StateManager struct {
	states map[int]*UserState
	mu     sync.RWMutex
}

// NewStateManager создает новый менеджер состояний
func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[int]*UserState),
	}
}

// SetState устанавливает состояние пользователя
func (sm *StateManager) SetState(userID int, step string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.states[userID]; !exists {
		sm.states[userID] = &UserState{
			CurrentStep: step,
			Data:        make(map[string]interface{}),
		}
	} else {
		sm.states[userID].CurrentStep = step
	}
}

// GetState возвращает состояние пользователя
func (sm *StateManager) GetState(userID int) (string, map[string]interface{}) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if state, exists := sm.states[userID]; exists {
		return state.CurrentStep, state.Data
	}
	return "", nil
}

// SetData устанавливает данные для пользователя
func (sm *StateManager) SetData(userID int, key string, value interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.states[userID]; !exists {
		sm.states[userID] = &UserState{
			Data: make(map[string]interface{}),
		}
	}
	sm.states[userID].Data[key] = value
}

// ClearState очищает состояние пользователя
func (sm *StateManager) ClearState(userID int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.states, userID)
}
