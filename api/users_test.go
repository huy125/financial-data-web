package api_test

import (
	"bytes"
	"context"
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

func (m *storeMock) Create(ctx context.Context, user model.User) error {
	m.Called(user)
	return nil
}

func (m *storeMock) List(ctx context.Context, limit, offset int) ([]model.User, error) {
	return []model.User{
		{
			ID:       1,
			Username: "test@xyz.com",
			Hash:     "hashedpassword1",
		},
		{
			ID:       2,
			Username: "example@abc.com",
			Hash:     "hashedpassword2",
		},
	}, nil
}

func TestServer_CreateUserHandler(t *testing.T) {
	tests := []struct {
		name string

		sendBody           string
		wantUser           model.User
		expectedStatusCode int
		expectedResponse      any
	}{
		{
			name: "creates user successfully",

			sendBody: `{"email":"test@example.com", "hash":"hashedpassword"}`,

			wantUser:           model.User{Username: "test@example.com", Hash: "hashedpassword"},
			expectedStatusCode: http.StatusCreated,
			expectedResponse: "User created successfully",
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
