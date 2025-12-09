//nolint:err113,funlen,gochecknoglobals,forcetypeassert,varnamelen
package schema

import (
	"context"
	"errors"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSchemaProvider is a mock implementation of components.SchemaProvider.
type MockSchemaProvider struct {
	mock.Mock
}

// Boilerplate set up, actual mock logic is in each test case.
func (m *MockSchemaProvider) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	args := m.Called(ctx, objects)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*common.ListObjectMetadataResult), args.Error(1)
}

func (m *MockSchemaProvider) SchemaSource() string {
	args := m.Called()

	return args.String(0)
}

// Predefined results for test inputs and outputs.
var userAndOrderSuccess = &common.ListObjectMetadataResult{
	Result: map[string]common.ObjectMetadata{
		"user": {
			DisplayName: "User",
			Fields:      map[string]common.FieldMetadata{"firstName": {DisplayName: "First Name"}},
			FieldsMap:   map[string]string{"firstName": "First Name"},
		},
		"order": {
			DisplayName: "Order",
			Fields:      map[string]common.FieldMetadata{"total": {DisplayName: "Total"}},
			FieldsMap:   map[string]string{"total": "Total"},
		},
	},
	Errors: map[string]error{},
}

var userOnlySuccess = &common.ListObjectMetadataResult{
	Result: map[string]common.ObjectMetadata{
		"user": {
			DisplayName: "User",
			Fields:      map[string]common.FieldMetadata{"firstName": {DisplayName: "First Name"}},
			FieldsMap:   map[string]string{"firstName": "First Name"},
		},
	},
	Errors: map[string]error{
		"order":   errors.New("order not found"),
		"product": errors.New("product not found"),
	},
}

var orderOnlySuccess = &common.ListObjectMetadataResult{
	Result: map[string]common.ObjectMetadata{
		"order": {
			DisplayName: "Order",
			Fields:      map[string]common.FieldMetadata{"total": {DisplayName: "Total"}},
			FieldsMap:   map[string]string{"total": "Total"},
		},
	},
	Errors: map[string]error{
		"product": errors.New("product not found"),
	},
}

var userAndOrderPartialSuccess = &common.ListObjectMetadataResult{
	Result: map[string]common.ObjectMetadata{
		"user": {
			DisplayName: "User",
			Fields:      map[string]common.FieldMetadata{"firstName": {DisplayName: "First Name"}},
			FieldsMap:   map[string]string{"firstName": "First Name"},
		},
		"order": {
			DisplayName: "Order",
			Fields:      map[string]common.FieldMetadata{"total": {DisplayName: "Total"}},
			FieldsMap:   map[string]string{"total": "Total"},
		},
	},
	Errors: map[string]error{
		"product": errors.New("product not found"),
	},
}

func TestCompositeSchemaProvider_ListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		objects        []string
		setupMocks     func(*MockSchemaProvider, *MockSchemaProvider)
		expectedResult *common.ListObjectMetadataResult
		expectedError  error
	}{
		{
			name:    "empty objects list",
			objects: []string{},
			setupMocks: func(mock1, mock2 *MockSchemaProvider) {
				// No calls expected for empty objects
			},
			expectedResult: &common.ListObjectMetadataResult{
				Result: make(map[string]common.ObjectMetadata),
				Errors: make(map[string]error),
			},
			expectedError: nil,
		},
		{
			name:    "first provider successfully gets metadata for all objects",
			objects: []string{"user", "order"},
			setupMocks: func(mock1, mock2 *MockSchemaProvider) {
				mock1.On("SchemaSource").Return("MockProvider1")
				// Because objects are first converted to a set then back to a slice,
				// the order is not guaranteed. We need to mock both possible orders.
				mock1.On("ListObjectMetadata", mock.Anything, []string{"user", "order"}).Return(
					userAndOrderSuccess,
					nil)
				mock1.On("ListObjectMetadata", mock.Anything, []string{"order", "user"}).Return(
					userAndOrderSuccess,
					nil)
			},
			expectedResult: userAndOrderSuccess,
			expectedError:  nil,
		},
		{
			name:    "first provider fails, second provider succeeds",
			objects: []string{"user", "order"},
			setupMocks: func(mock1, mock2 *MockSchemaProvider) {
				mock1.On("SchemaSource").Return("MockProvider1")
				mock1.On("ListObjectMetadata", mock.Anything, mock.Anything).Return(
					nil, errors.New("provider 1 failed"))

				mock2.On("SchemaSource").Return("MockProvider2")
				mock2.On("ListObjectMetadata", mock.Anything, []string{"user", "order"}).Return(
					userAndOrderSuccess, nil)
				mock2.On("ListObjectMetadata", mock.Anything, []string{"order", "user"}).Return(
					userAndOrderSuccess, nil)
			},
			expectedResult: userAndOrderSuccess,
			expectedError:  nil,
		},
		{
			name:    "both providers have partial success",
			objects: []string{"user", "order", "product"},
			setupMocks: func(mock1, mock2 *MockSchemaProvider) {
				mock1.On("SchemaSource").Return("MockProvider1")
				mock1.On("ListObjectMetadata", mock.Anything, []string{"order", "product", "user"}).Return(userOnlySuccess, nil)
				mock1.On("ListObjectMetadata", mock.Anything, []string{"order", "user", "product"}).Return(userOnlySuccess, nil)
				mock1.On("ListObjectMetadata", mock.Anything, []string{"product", "order", "user"}).Return(userOnlySuccess, nil)
				mock1.On("ListObjectMetadata", mock.Anything, []string{"product", "user", "order"}).Return(userOnlySuccess, nil)
				mock1.On("ListObjectMetadata", mock.Anything, []string{"user", "order", "product"}).Return(userOnlySuccess, nil)
				mock1.On("ListObjectMetadata", mock.Anything, []string{"user", "product", "order"}).Return(userOnlySuccess, nil)

				mock2.On("SchemaSource").Return("MockProvider2")
				mock2.On("ListObjectMetadata", mock.Anything, []string{"order", "product"}).Return(orderOnlySuccess, nil)
				mock2.On("ListObjectMetadata", mock.Anything, []string{"product", "order"}).Return(orderOnlySuccess, nil)
			},
			expectedResult: userAndOrderPartialSuccess,
			expectedError:  nil,
		},
		{
			name:    "all providers fail",
			objects: []string{"user", "order"},
			setupMocks: func(mock1, mock2 *MockSchemaProvider) {
				mock1.On("SchemaSource").Return("MockProvider1")
				mock1.On("ListObjectMetadata", mock.Anything, mock.Anything).Return(
					nil, errors.New("provider 1 failed"))

				mock2.On("SchemaSource").Return("MockProvider2")
				mock2.On("ListObjectMetadata", mock.Anything, mock.Anything).Return(
					nil, errors.New("provider 2 failed"))
			},
			expectedResult: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				// The second provider's errors win,
				// since it's the last provider.
				Errors: map[string]error{
					"user":  errors.New("provider 2 failed"),
					"order": errors.New("provider 2 failed"),
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create mock providers
			mockProvider1 := &MockSchemaProvider{}
			mockProvider2 := &MockSchemaProvider{}

			// Setup mocks
			tt.setupMocks(mockProvider1, mockProvider2)

			// Create composite provider
			compositeProvider := NewCompositeSchemaProvider(mockProvider1, mockProvider2)

			// Execute test
			result, err := compositeProvider.ListObjectMetadata(t.Context(), tt.objects)

			// Assertions
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
