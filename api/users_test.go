package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/huy125/financial-data-web/api"
	model "github.com/huy125/financial-data-web/api/store/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type storeMock struct {
	mock.Mock
}

func (m *storeMock) Create(_ context.Context, user model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *storeMock) List(_ context.Context, limit, offset int) ([]model.User, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *storeMock) Find(_ context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if user, ok := args.Get(0).(*model.User); ok {
		return user, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *storeMock) Update(_ context.Context, id uuid.UUID, userUpdate model.UserUpdate) error {
	args := m.Called(id, userUpdate)
	return args.Error(0)
}

func TestServer_CreateUserHandler(t *testing.T) {
	tests := []struct {
		name string

		sendBody           string
		wantUser           model.User
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			name: "creates user successfully",

			sendBody: `{"email":"test@example.com", "hash":"hashedpassword"}`,

			wantUser:           model.User{Username: "test@example.com", Hash: "hashedpassword"},
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   "User created successfully",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store := &storeMock{}
			if test.wantUser != (model.User{}) {
				store.On("Create", test.wantUser).Return(nil)
			}

			apiKey := "testApiKey"
			srv := api.New(apiKey, store)

			httpSrv := httptest.NewServer(srv)
			t.Cleanup(func() { httpSrv.Close() })

			reqURL := httpSrv.URL + "/users"

			var req *http.Request
			req, _ = http.NewRequest(http.MethodPost, reqURL, bytes.NewBufferString(test.sendBody))

			rr := httptest.NewRecorder()

			srv.ServeHTTP(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)

			var res string
			err := json.Unmarshal(rr.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, test.expectedResponse, res)

			store.AssertExpectations(t)
		})
	}
}

func TestServer_UpdateUserHandler(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name string

		uri      string
		sendBody string

		existUser      *model.User
		wantUser     model.UserUpdate
		shouldMockFind bool

		expectedStatusCode int
		expectedResponse   []byte
	}{
		{
			name:     "updates user successfully",
			uri:      "/users/" + id.String(),
			sendBody: `{"username":"test1@example.com"}`,

			existUser:      &model.User{ID: id, Username: "test@example.com", Hash: "hashpassword"},
			wantUser:     model.UserUpdate{Username: "test1@example.com"},
			shouldMockFind: true,

			expectedStatusCode: http.StatusOK,
			expectedResponse:   []byte(`{"message":"User updated successfully"}`),
		},
		{
			name:     "handles user not found",
			uri:      "/users/" + id.String(),
			sendBody: `{"username":"test1@example.com"}`,

			existUser:      nil,
			wantUser:     model.UserUpdate{},
			shouldMockFind: true,

			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:     "handles invalid id",
			uri:      "/users/invalid-id",
			sendBody: `{"username":"test1@example.com"}`,

			existUser:      nil,
			wantUser:     model.UserUpdate{},
			shouldMockFind: false,

			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store := &storeMock{}
			if test.shouldMockFind {
				store.On("Find", id).Return(test.existUser, nil)
			}

			if test.wantUser != (model.UserUpdate{}) {
				store.On("Update", id, test.wantUser).Return(nil)
			}

			apiKey := "testApiKey"
			srv := api.New(apiKey, store)

			httpSrv := httptest.NewServer(srv)
			t.Cleanup(func() { httpSrv.Close() })

			reqURL := httpSrv.URL + test.uri

			req, err := http.NewRequest(http.MethodPatch, reqURL, bytes.NewBufferString(test.sendBody))
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			srv.ServeHTTP(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)

			if rr.Code == http.StatusOK {
				res, err := io.ReadAll(rr.Body)
				require.NoError(t, err)

				assert.JSONEq(t, string(test.expectedResponse), string(res))
			}

			store.AssertExpectations(t)
		})
	}
}
