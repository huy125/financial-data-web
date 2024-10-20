package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	model "github.com/huy125/financial-data-web/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) Create(user model.User) model.User {
	args := m.Called(user)
	return args.Get(0).(model.User)
}

func (m *MockStore) List() []model.User {
	args := m.Called()
	return args.Get(0).([]model.User)
}

func TestOKCreateUserHandler(t *testing.T) {
	mockStore := new(MockStore)
	handler := &UserHandler{Store: mockStore}

	t.Run("Method not allowed", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequest(http.MethodGet, "/users", nil)
		rr := httptest.NewRecorder()

		handler.CreateUserHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
		assert.Equal(t, "Method not allowed\n", rr.Body.String())
	})

	t.Run("Successful user creation", func(t *testing.T) {
		t.Parallel()
		validPayload := `
		{
			"email":"test@example.com",
			"hash":"hashedpassword"
		}
	`
		req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(validPayload))
		rr := httptest.NewRecorder()

		expectedUser := model.User{
			ID:       1,
			Username: "test@example.com",
			Hash:     "hashedpassword",
		}

		mockStore.On("Create", mock.AnythingOfType("model.User")).Return(expectedUser)

		handler.CreateUserHandler(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var res model.User
		err := json.Unmarshal(rr.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, res)

		mockStore.AssertCalled(t, "Create", mock.AnythingOfType("model.User"))
	})
}
