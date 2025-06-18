package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hamba/cmd/v2/observe"
	"github.com/huy125/financial-data-web/api"
	"github.com/huy125/financial-data-web/store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func BenchmarkGetUserHandler(b *testing.B) {
	id := uuid.New()

	storeMock := &storeMock{}
	wantUserModel := &store.User{
		Model: store.Model{
			ID:        id,
			CreatedAt: time.Date(2024, 11, 24, 21, 58, 0o0, 0o0, time.UTC),
			UpdatedAt: time.Now(),
		},
		Email:     "test@example.com",
		Firstname: "Bob",
		Lastname:  "Smith",
	}
	storeMock.On("FindUser", id).Return(wantUserModel, nil)

	authMock := &authenticatorMock{}
	authMock.On("RequireAuth", mock.AnythingOfType("http.HandlerFunc")).
		Return(func(h http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				h(w, r)
			}
		})

	obsvr := observe.NewFake()
	srv := api.New(testAPIKey, testFilePath, storeMock, authMock, obsvr)

	httpSrv := httptest.NewServer(srv)
	b.Cleanup(func() { httpSrv.Close() })

	reqURL := httpSrv.URL + "/users/" + id.String()

	ctx, cancel := context.WithTimeout(b.Context(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	require.NoError(b, err)

	rr := httptest.NewRecorder()

	for range b.N {
		srv.ServeHTTP(rr, req)
		require.Equal(b, http.StatusOK, rr.Code)
	}
}
