package api_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/coreos/go-oidc/v3/oidc/oidctest"
	"github.com/google/uuid"
	"github.com/hamba/cmd/v2/observe"
	"github.com/huy125/finscope/api"
	"github.com/huy125/finscope/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

const (
	testAPIKey   = "testAPIKey"
	testFilePath = "testFilePath"
)

func TestServer_CreateUserHandler(t *testing.T) {
	t.Parallel()

	validBody := `{"email": "test@example.com", "firstname": "Alice", "lastname": "Smith"}`
	now := time.Now()
	tests := []struct {
		name string

		sendBody       string
		wantCreateUser *store.CreateUser
		wantUserModel  *store.User

		returnErr error

		wantStatus int
		wantResult []byte
	}{
		{
			name: "creates user successfully",

			sendBody: validBody,

			wantCreateUser: &store.CreateUser{Email: "test@example.com", Firstname: "Alice", Lastname: "Smith"},
			wantUserModel: &store.User{
				Model: store.Model{
					ID:        uuid.MustParse("ab678e01-00ee-4e4c-acfc-6dc0b68fee20"),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Email:     "test@example.com",
				Firstname: "Alice",
				Lastname:  "Smith",
			},
			returnErr: nil,

			wantStatus: http.StatusCreated,
			wantResult: []byte(`
				{
					"id": "ab678e01-00ee-4e4c-acfc-6dc0b68fee20",
					"email": "test@example.com",
					"firstname": "Alice",
					"lastname": "Smith"
				}`,
			),
		},
		{
			name: "handles internal server error",

			sendBody: validBody,

			wantCreateUser: &store.CreateUser{Email: "test@example.com", Firstname: "Alice", Lastname: "Smith"},
			wantUserModel:  nil,
			returnErr:      errors.New("internal error"),

			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "handles bad request error",

			sendBody:       "invalid request",
			wantCreateUser: nil,
			wantUserModel:  nil,

			wantStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			storeMock := &storeMock{}
			if test.wantCreateUser != nil {
				storeMock.On("CreateUser", test.wantCreateUser).Return(test.wantUserModel, test.returnErr)
			}

			authMock := &authenticatorMock{}
			idToken := createIDToken(t)

			authMock.On("ExtractTokenFromRequest", mock.AnythingOfType("*http.Request")).
				Return("valid-token")
			authMock.On("VerifyAccessToken", mock.AnythingOfType("*context.timerCtx"), &oauth2.Token{AccessToken: "valid-token"}).
				Return(idToken, nil)

			obsvr := observe.NewFake()
			srv := api.New(testAPIKey, testFilePath, storeMock, authMock, obsvr)

			httpSrv := httptest.NewServer(srv)
			t.Cleanup(func() { httpSrv.Close() })

			reqURL := httpSrv.URL + "/users"
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBufferString(test.sendBody))

			require.NoError(t, err)

			rr := httptest.NewRecorder()

			srv.ServeHTTP(rr, req)

			assert.Equal(t, test.wantStatus, rr.Code)

			if test.wantResult != nil {
				res, err := io.ReadAll(rr.Body)
				require.NoError(t, err)

				assert.JSONEq(t, string(test.wantResult), string(res))
			}

			storeMock.AssertExpectations(t)
		})
	}
}

func TestServer_UpdateUserHandler(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	validBody := `{"id": "` + id.String() + `", "email": "test@example.com", "firstname": "Bob", "lastname": "Smith"}`
	tests := []struct {
		name string

		sendBody string

		wantUpdateUser *store.UpdateUser
		wantUserModel  *store.User
		returnErr      error

		expectedStatusCode int
		expectedResponse   []byte
	}{
		{
			name:     "updates user successfully",
			sendBody: validBody,

			wantUpdateUser: &store.UpdateUser{
				ID:         id,
				CreateUser: store.CreateUser{Email: "test@example.com", Firstname: "Bob", Lastname: "Smith"},
			},
			wantUserModel: &store.User{
				Model: store.Model{
					ID:        id,
					CreatedAt: time.Date(2024, 11, 24, 21, 58, 0o0, 0o0, time.UTC),
					UpdatedAt: time.Now(),
				},
				Email:     "test@example.com",
				Firstname: "Bob",
				Lastname:  "Smith",
			},
			returnErr: nil,

			expectedStatusCode: http.StatusOK,
			expectedResponse:   []byte(validBody),
		},
		{
			name:     "handles user not found error",
			sendBody: validBody,

			wantUpdateUser: &store.UpdateUser{
				ID:         id,
				CreateUser: store.CreateUser{Email: "test@example.com", Firstname: "Bob", Lastname: "Smith"},
			},
			wantUserModel: nil,
			returnErr:     store.ErrNotFound,

			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:     "handles internal server error",
			sendBody: validBody,

			wantUpdateUser: &store.UpdateUser{
				ID:         id,
				CreateUser: store.CreateUser{Email: "test@example.com", Firstname: "Bob", Lastname: "Smith"},
			},
			wantUserModel: nil,
			returnErr:     errors.New("internal error"),

			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:     "handles bad request error",
			sendBody: "invalid request",

			wantUpdateUser: nil,
			wantUserModel:  nil,

			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			storeMock := &storeMock{}
			if test.wantUpdateUser != nil {
				storeMock.On("UpdateUser", test.wantUpdateUser).Return(test.wantUserModel, test.returnErr)
			}

			authMock := &authenticatorMock{}
			idToken := createIDToken(t)

			authMock.On("ExtractTokenFromRequest", mock.AnythingOfType("*http.Request")).
				Return("valid-token")
			authMock.On("VerifyAccessToken", mock.AnythingOfType("*context.timerCtx"), &oauth2.Token{AccessToken: "valid-token"}).
				Return(idToken, nil)

			obsvr := observe.NewFake()
			srv := api.New(testAPIKey, testFilePath, storeMock, authMock, obsvr)

			httpSrv := httptest.NewServer(srv)
			t.Cleanup(func() { httpSrv.Close() })

			reqURL := httpSrv.URL + "/users/" + id.String()

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewBufferString(test.sendBody))
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			srv.ServeHTTP(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)

			if rr.Code == http.StatusOK {
				res, err := io.ReadAll(rr.Body)
				require.NoError(t, err)

				assert.JSONEq(t, string(test.expectedResponse), string(res))
			}

			storeMock.AssertExpectations(t)
		})
	}
}

func TestServer_GetUserHandler(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	tests := []struct {
		name string

		wantUserModel *store.User
		returnErr     error

		wantStatus int
		wantResult []byte
	}{
		{
			name: "returns user successfully",

			wantUserModel: &store.User{
				Model: store.Model{
					ID:        id,
					CreatedAt: time.Date(2024, 11, 24, 21, 58, 0o0, 0o0, time.UTC),
					UpdatedAt: time.Now(),
				},
				Email:     "test@example.com",
				Firstname: "Bob",
				Lastname:  "Smith",
			},
			returnErr: nil,

			wantStatus: http.StatusOK,
			wantResult: []byte(`
				{
					"id": "` + id.String() + `",
					"email": "test@example.com",
					"firstname": "Bob",
					"lastname": "Smith"
				}`,
			),
		},
		{
			name: "handles user not found error",

			wantUserModel: nil,
			returnErr:     store.ErrNotFound,

			wantStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			storeMock := &storeMock{}
			storeMock.On("FindUser", id).Return(test.wantUserModel, test.returnErr)

			authMock := &authenticatorMock{}
			idToken := createIDToken(t)
			authMock.On("ExtractTokenFromRequest", mock.AnythingOfType("*http.Request")).
				Return("valid-token")
			authMock.On("VerifyAccessToken", mock.AnythingOfType("*context.timerCtx"), &oauth2.Token{AccessToken: "valid-token"}).
				Return(idToken, nil)

			obsvr := observe.NewFake()
			srv := api.New(testAPIKey, testFilePath, storeMock, authMock, obsvr)

			httpSrv := httptest.NewServer(srv)
			t.Cleanup(func() { httpSrv.Close() })

			reqURL := httpSrv.URL + "/users/" + id.String()

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			srv.ServeHTTP(rr, req)

			assert.Equal(t, test.wantStatus, rr.Code)

			if rr.Code == http.StatusOK {
				res, err := io.ReadAll(rr.Body)
				require.NoError(t, err)

				assert.JSONEq(t, string(test.wantResult), string(res))
			}

			storeMock.AssertExpectations(t)
		})
	}
}

type storeMock struct {
	mock.Mock
}

func (m *storeMock) CreateUser(_ context.Context, user *store.CreateUser) (*store.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *storeMock) ListUsers(_ context.Context, limit, offset int) ([]store.User, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]store.User), args.Error(1)
}

func (m *storeMock) FindUser(_ context.Context, id uuid.UUID) (*store.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*store.User), args.Error(1)
}

func (m *storeMock) UpdateUser(_ context.Context, user *store.UpdateUser) (*store.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *storeMock) FindStockBySymbol(_ context.Context, symbol string) (*store.Stock, error) {
	args := m.Called(symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*store.Stock), args.Error(1)
}

func (m *storeMock) ListMetrics(_ context.Context, limit, offset int) ([]store.Metric, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]store.Metric), args.Error(1)
}

func (m *storeMock) CreateStockMetric(
	_ context.Context,
	storeID, metricID uuid.UUID,
	value float64,
) (*store.StockMetric, error) {
	args := m.Called(storeID, metricID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.StockMetric), args.Error(1)
}

func (m *storeMock) FindLatestStockMetrics(_ context.Context, stockID uuid.UUID) ([]store.LatestStockMetric, error) {
	args := m.Called(stockID)

	return args.Get(0).([]store.LatestStockMetric), args.Error(1)
}

func (m *storeMock) CreateAnalysis(
	_ context.Context,
	userID, stockID uuid.UUID,
	score float64,
) (*store.Analysis, error) {
	args := m.Called(userID, stockID, score)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Analysis), args.Error(1)
}

func (m *storeMock) CreateRecommendation(
	_ context.Context,
	analysisID uuid.UUID,
	action store.Action,
	confidenceLevel float64,
	reason string,
) (*store.Recommendation, error) {
	args := m.Called(analysisID, action, confidenceLevel, reason)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.Recommendation), args.Error(1)
}

type authenticatorMock struct {
	mock.Mock
}

func (m *authenticatorMock) ExtractTokenFromRequest(r *http.Request) string {
	args := m.Called(r)
	return args.String(0)
}

func (m *authenticatorMock) VerifyAccessToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*oidc.IDToken), args.Error(1)
}

func (m *authenticatorMock) GenerateState() (string, error) {
	args := m.Called()
	return args.String(), nil
}

func (m *authenticatorMock) GetBaseURL() (string, error) {
	args := m.Called()
	return args.String(), nil
}

func (m *authenticatorMock) VerifyState(s string) bool {
	args := m.Called(s)
	return args.Bool(0)
}

func (m *authenticatorMock) VerifyToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*oidc.IDToken), args.Error(1)
}

func (m *authenticatorMock) RevokeToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *authenticatorMock) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(*oauth2.Token), args.Error(1)
}

func (m *authenticatorMock) GetClientID() string {
	args := m.Called()
	return args.String(0)
}

func (m *authenticatorMock) GetClientOrigin() string {
	args := m.Called()
	return args.String(0)
}

func createIDToken(tb testing.TB) *oidc.IDToken {
	tb.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		tb.Fatalf("creating server: %v", err)
	}

	s := &oidctest.Server{
		PublicKeys: []oidctest.PublicKey{
			{
				PublicKey: priv.Public(),
				KeyID:     "my-key-id",
				Algorithm: oidc.RS256,
			},
		},
	}

	srv := httptest.NewServer(s)
	defer tb.Cleanup(srv.Close)
	s.SetIssuer(srv.URL)

	now := time.Now()
	rawClaims := `{
			"iss": "` + srv.URL + `",
			"aud": "my-client-id",
			"sub": "foo",
			"exp": ` + strconv.FormatInt(now.Add(time.Hour).Unix(), 10) + `,
			"email": "foo@example.com",
			"email_verified": true
		}`
	token := oidctest.SignIDToken(priv, "my-key-id", oidc.RS256, rawClaims)

	ctx := context.Background()
	p, err := oidc.NewProvider(ctx, srv.URL)
	if err != nil {
		tb.Fatalf("new provider: %v", err)
	}
	config := &oidc.Config{
		ClientID: "my-client-id",
		Now:      func() time.Time { return now },
	}
	v := p.VerifierContext(ctx, config)

	idToken, err := v.Verify(ctx, token)
	if err != nil {
		tb.Fatalf("verify token: %v", err)
	}

	return idToken
}
