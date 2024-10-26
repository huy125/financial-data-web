package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huy125/financial-data-web/api"
	model "github.com/huy125/financial-data-web/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type storeMock struct {
	mock.Mock
}

func (m *storeMock) Create(user model.User) model.User {
	args := m.Called(user)
	return args.Get(0).(model.User)
}

func (m *storeMock) List() []model.User {
	args := m.Called()
	return args.Get(0).([]model.User)
}

func TestServer_CreateUserHandler(t *testing.T) {
	tests := []struct {
		name string

		sendBody           string
		wantUser           model.User
		expectedStatusCode int
		expectedUser       model.User
	}{
		{
			name: "creates user",

			sendBody: `{"email":"test@example.com", "hash":"hashedpassword"}`,

			wantUser:           model.User{Username: "test@example.com", Hash: "hashedpassword"},
			expectedStatusCode: http.StatusCreated,
			expectedUser: model.User{
				ID:       1,
				Username: "test@example.com",
				Hash:     "hashedpassword",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store := &storeMock{}
			if test.wantUser != (model.User{}) {
				store.On("Create", test.wantUser).Return(test.expectedUser)
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

			var res model.User

			err := json.Unmarshal(rr.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Equal(t, test.expectedUser, res)

			store.AssertExpectations(t)
		})
	}
}
