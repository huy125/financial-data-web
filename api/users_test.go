package api_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hamba/cmd/v2/observe"
	"github.com/huy125/financial-data-web/api"
	"github.com/huy125/financial-data-web/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
			authMock.On("RequireAuth", mock.AnythingOfType("http.HandlerFunc")).
				Return(func(h http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, r *http.Request) {
						h(w, r)
					}
				})

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
			authMock.On("RequireAuth", mock.AnythingOfType("http.HandlerFunc")).
				Return(func(h http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, r *http.Request) {
						h(w, r)
					}
				})

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
			authMock.On("RequireAuth", mock.AnythingOfType("http.HandlerFunc")).
				Return(func(h http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, r *http.Request) {
						h(w, r)
					}
				})

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
	args := m.Called(userID, stockID)
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

func (m *authenticatorMock) LoginHandler(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func (m *authenticatorMock) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func (m *authenticatorMock) RequireAuth(handle http.HandlerFunc) http.HandlerFunc {
	return handle
}

func (m *authenticatorMock) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}
