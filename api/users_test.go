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
	"github.com/huy125/financial-data-web/api"
	"github.com/huy125/financial-data-web/api/dto"
	"github.com/huy125/financial-data-web/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testAPIKey = "testAPIKey"

type storeMock struct {
	mock.Mock
}

func (m *storeMock) Create(_ context.Context, user *store.User) (*store.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *storeMock) List(_ context.Context, limit, offset int) ([]store.User, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]store.User), args.Error(1)
}

func (m *storeMock) Find(_ context.Context, id uuid.UUID) (*store.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*store.User), args.Error(1)
}

func (m *storeMock) Update(_ context.Context, user *store.User) (*store.User, error) {
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

func TestServer_CreateUserHandler(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []struct {
		name string

		sendBody      string
		wantUserDto   *dto.UserDto
		wantUserModel *store.User

		returnErr error

		wantStatus int
		wantResult []byte
	}{
		{
			name: "creates user successfully",

			sendBody: `{"email": "test@example.com", "firstname": "Alice", "lastname": "Smith"}`,

			wantUserDto: &dto.UserDto{Email: "test@example.com", Firstname: "Alice", Lastname: "Smith"},
			wantUserModel: &store.User{
				ID:        uuid.MustParse("ab678e01-00ee-4e4c-acfc-6dc0b68fee20"),
				Email:     "test@example.com",
				Firstname: "Alice",
				Lastname:  "Smith",
				CreatedAt: &now,
				UpdatedAt: &now,
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

			sendBody: `{"email": "test@example.com", "firstname": "Alice", "lastname": "Smith"}`,

			wantUserDto:   &dto.UserDto{Email: "test@example.com", Firstname: "Alice", Lastname: "Smith"},
			wantUserModel: nil,
			returnErr:     errors.New("internal error"),

			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "handles bad request error",

			sendBody:      "invalid request",
			wantUserDto:   nil,
			wantUserModel: nil,

			wantStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			storeMock := &storeMock{}
			if test.wantUserDto != nil {
				storeMock.On("Create",
					mock.MatchedBy(func(u *store.User) bool {
						return u.Email == test.wantUserDto.Email &&
							u.Firstname == test.wantUserDto.Firstname &&
							u.Lastname == test.wantUserDto.Lastname
					}),
				).Return(test.wantUserModel, test.returnErr)
			}

			srv := api.New(testAPIKey, storeMock)

			httpSrv := httptest.NewServer(srv)
			t.Cleanup(func() { httpSrv.Close() })

			reqURL := httpSrv.URL + "/users"

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
	createdAt := time.Date(2024, 11, 24, 21, 58, 0o0, 0o0, time.UTC)
	updatedAt := time.Now()
	tests := []struct {
		name string

		sendBody string

		wantUserDto   *dto.UserDto
		wantUserModel *store.User
		returnErr     error

		expectedStatusCode int
		expectedResponse   []byte
	}{
		{
			name:     "updates user successfully",
			sendBody: `{"id": "` + id.String() + `", "email": "test@example.com", "firstname": "Bob", "lastname": "Smith"}`,

			wantUserDto: &dto.UserDto{ID: id.String(), Email: "test@example.com", Firstname: "Bob", Lastname: "Smith"},
			wantUserModel: &store.User{
				ID:        id,
				Email:     "test@example.com",
				Firstname: "Bob",
				Lastname:  "Smith",
				CreatedAt: &createdAt,
				UpdatedAt: &updatedAt,
			},
			returnErr: nil,

			expectedStatusCode: http.StatusOK,
			expectedResponse: []byte(`
				{
					"id": "` + id.String() + `",
					"email": "test@example.com",
					"firstname": "Bob",
					"lastname": "Smith"
				}`,
			),
		},
		{
			name:     "handles user not found error",
			sendBody: `{"id": "` + id.String() + `", "email": "test@example.com", "firstname": "Bob", "lastname": "Smith"}`,

			wantUserDto:   &dto.UserDto{ID: id.String(), Email: "test@example.com", Firstname: "Bob", Lastname: "Smith"},
			wantUserModel: nil,
			returnErr:     store.ErrNotFound,

			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:     "handles internal server error",
			sendBody: `{"id": "` + id.String() + `", "email": "test@example.com", "firstname": "Bob", "lastname": "Smith"}`,

			wantUserDto:   &dto.UserDto{ID: id.String(), Email: "test@example.com", Firstname: "Bob", Lastname: "Smith"},
			wantUserModel: nil,
			returnErr:     errors.New("internal error"),

			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:     "handles bad request error",
			sendBody: "invalid request",

			wantUserDto:   nil,
			wantUserModel: nil,

			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			storeMock := &storeMock{}
			if test.wantUserDto != nil {
				storeMock.On("Update",
					mock.MatchedBy(func(u *store.User) bool {
						return u.ID.String() == test.wantUserDto.ID &&
							u.Email == test.wantUserDto.Email &&
							u.Firstname == test.wantUserDto.Firstname &&
							u.Lastname == test.wantUserDto.Lastname
					}),
				).Return(test.wantUserModel, test.returnErr)
			}

			srv := api.New(testAPIKey, storeMock)

			httpSrv := httptest.NewServer(srv)
			t.Cleanup(func() { httpSrv.Close() })

			reqURL := httpSrv.URL + "/users/" + id.String()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
	createdAt := time.Date(2024, 11, 24, 21, 58, 0o0, 0o0, time.UTC)
	updatedAt := time.Now()
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
				ID:        id,
				Email:     "test@example.com",
				Firstname: "Bob",
				Lastname:  "Smith",
				CreatedAt: &createdAt,
				UpdatedAt: &updatedAt,
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
			storeMock.On("Find", id).Return(test.wantUserModel, test.returnErr)

			srv := api.New(testAPIKey, storeMock)

			httpSrv := httptest.NewServer(srv)
			t.Cleanup(func() { httpSrv.Close() })

			reqURL := httpSrv.URL + "/users/" + id.String()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
