package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hamba/cmd/v2/observe"
	"github.com/huy125/finscope/api"
	"github.com/huy125/finscope/store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
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
	idToken := createIDToken(b)
	authMock.On("ExtractTokenFromRequest", mock.AnythingOfType("*http.Request")).
		Return("valid-token")
	authMock.On("VerifyAccessToken", mock.AnythingOfType("*context.timerCtx"), &oauth2.Token{AccessToken: "valid-token"}).
		Return(idToken, nil)

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
