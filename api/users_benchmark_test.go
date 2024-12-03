package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/huy125/financial-data-web/api"
	model "github.com/huy125/financial-data-web/api/store/models"
	"github.com/stretchr/testify/require"
)

func BenchmarkGetUserHandler(b *testing.B) {
	id := uuid.New()

	store := &storeMock{}
	wantUserModel := &model.User{
		ID:        id,
		Email:     "test@example.com",
		Firstname: "Bob",
		Lastname:  "Smith",
		CreatedAt: time.Date(2024, 11, 24, 21, 58, 0o0, 0o0, time.UTC),
		UpdatedAt: time.Now(),
	}
	store.On("Find", id).Return(wantUserModel, nil)

	srv := api.New(testAPIKey, store)

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
