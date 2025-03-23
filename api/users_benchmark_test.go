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

	obsvr := observe.NewFake()
	srv := api.New(testAPIKey, storeMock, obsvr)

	httpSrv := httptest.NewServer(srv)
	b.Cleanup(func() { httpSrv.Close() })

	reqURL := httpSrv.URL + "/users/" + id.String()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	require.NoError(b, err)

	rr := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(rr, req)
		require.Equal(b, http.StatusOK, rr.Code)
	}
}
