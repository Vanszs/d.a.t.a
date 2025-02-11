package actions_test

import (
	"time"
)

const (
	defaultTimeout = 120 * time.Second
	maxRetries     = 2
)

// func setupTestProvider(t *testing.T) actions.DatabaseProvider {
// 	apiURL := os.Getenv("DATA_API_KEY")
// 	if apiURL == "" {
// 		t.Skip("DATA_API_KEY not set")
// 	}
// 	authToken := os.Getenv("DATA_AUTH_TOKEN")
// 	if authToken == "" {
// 		t.Skip("DATA_AUTH_TOKEN not set")
// 	}

// 	return actions.NewDatabaseProvider(apiURL, authToken, "ethereum-mainnet", nil)
// }

// // MockLLMClient is a mock implementation of llm.Client
// type MockLLMClient struct {
// 	mock.Mock
// }

// func (m *MockLLMClient) CreateCompletion(ctx context.Context, request llm.CompletionRequest) (string, error) {
// 	args := m.Called(ctx, request)
// 	return args.String(0), args.Error(1)
// }

// // APIResponseRow represents a row in the API response
// type APIResponseRow struct {
// 	Items []interface{} `json:"items"`
// }

// // APIResponseData represents the data in the API response
// type APIResponseData struct {
// 	ColumnInfos []string         `json:"column_infos"`
// 	Rows        []APIResponseRow `json:"rows"`
// }

// // TestAPIResponse represents the test API response
// type TestAPIResponse struct {
// 	Code int             `json:"code"`
// 	Msg  string          `json:"msg"`
// 	Data APIResponseData `json:"data"`
// }

// func TestBuildSQLQuery(t *testing.T) {
// 	provider := helper.SetupTestProvider(t)
// 	providerImpl, ok := provider.(*actions.DatabaseProviderImpl)
// 	require.True(t, ok, "provider should be of type *DatabaseProviderImpl")

// 	validAddress := "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
// 	tests := []struct {
// 		name     string
// 		params   actions.FetchTransactionParams
// 		expected string
// 		wantErr  bool
// 	}{
// 		{
// 			name:     "Empty params",
// 			params:   actions.FetchTransactionParams{},
// 			expected: "SELECT * FROM eth.transactions WHERE 1=1 AND date >= date_format(date_add('day', -90, current_date), '%Y-%m-%d') LIMIT 100",
// 			wantErr:  false,
// 		},
// 		{
// 			name: "With address",
// 			params: actions.FetchTransactionParams{
// 				Address: helper.StrPtr(validAddress),
// 			},
// 			expected: fmt.Sprintf("SELECT * FROM eth.transactions WHERE 1=1 AND date >= date_format(date_add('day', -90, current_date), '%%Y-%%m-%%d') AND (from_address = '%s' OR to_address = '%s') LIMIT 100", validAddress, validAddress),
// 			wantErr:  false,
// 		},
// 		{
// 			name: "With date range",
// 			params: actions.FetchTransactionParams{
// 				StartDate: helper.StrPtr("2024-01-01"),
// 				EndDate:   helper.StrPtr("2024-02-01"),
// 			},
// 			expected: "SELECT * FROM eth.transactions WHERE 1=1 AND date_parse(date, '%Y-%m-%d') >= date_parse('2024-01-01', '%Y-%m-%d') AND date_parse(date, '%Y-%m-%d') <= date_parse('2024-02-01', '%Y-%m-%d') LIMIT 100",
// 			wantErr:  false,
// 		},
// 		{
// 			name: "With value range",
// 			params: actions.FetchTransactionParams{
// 				MinValue: helper.StrPtr("1.5"),
// 				MaxValue: helper.StrPtr("10.0"),
// 			},
// 			expected: "SELECT * FROM eth.transactions WHERE 1=1 AND date >= date_format(date_add('day', -90, current_date), '%Y-%m-%d') AND value >= 1.5 AND value <= 10.0 LIMIT 100",
// 			wantErr:  false,
// 		},
// 		{
// 			name: "With order and limit",
// 			params: actions.FetchTransactionParams{
// 				OrderBy:        helper.StrPtr("value"),
// 				OrderDirection: helper.StrPtr("ASC"),
// 				Limit:          helper.IntPtr(10),
// 			},
// 			expected: "SELECT * FROM eth.transactions WHERE 1=1 AND date >= date_format(date_add('day', -90, current_date), '%Y-%m-%d') ORDER BY value ASC LIMIT 10",
// 			wantErr:  false,
// 		},
// 		{
// 			name: "Invalid order by field",
// 			params: actions.FetchTransactionParams{
// 				OrderBy: helper.StrPtr("invalid_field"),
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Invalid order direction",
// 			params: actions.FetchTransactionParams{
// 				OrderBy:        helper.StrPtr("value"),
// 				OrderDirection: helper.StrPtr("INVALID"),
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Invalid address format",
// 			params: actions.FetchTransactionParams{
// 				Address: helper.StrPtr("invalid_address"),
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Invalid date format",
// 			params: actions.FetchTransactionParams{
// 				StartDate: helper.StrPtr("invalid_date"),
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "All parameters",
// 			params: actions.FetchTransactionParams{
// 				Address:        helper.StrPtr(validAddress),
// 				StartDate:      helper.StrPtr("2024-01-01"),
// 				EndDate:        helper.StrPtr("2024-02-01"),
// 				MinValue:       helper.StrPtr("1.5"),
// 				MaxValue:       helper.StrPtr("10.0"),
// 				OrderBy:        helper.StrPtr("value"),
// 				OrderDirection: helper.StrPtr("ASC"),
// 				Limit:          helper.IntPtr(10),
// 			},
// 			expected: fmt.Sprintf("SELECT * FROM eth.transactions WHERE 1=1 AND date_parse(date, '%%Y-%%m-%%d') >= date_parse('2024-01-01', '%%Y-%%m-%%d') AND date_parse(date, '%%Y-%%m-%%d') <= date_parse('2024-02-01', '%%Y-%%m-%%d') AND (from_address = '%s' OR to_address = '%s') AND value >= 1.5 AND value <= 10.0 ORDER BY value ASC LIMIT 10", validAddress, validAddress),
// 			wantErr:  false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := providerImpl.BuildSQLQuery(tt.params)
// 			if tt.wantErr {
// 				assert.Error(t, err)
// 				return
// 			}
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expected, got)
// 		})
// 	}
// }

// func TestExecuteQuery(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration test in short mode")
// 	}

// 	provider := setupTestProvider(t)
// 	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
// 	defer cancel()

// 	// Test simple query
// 	t.Run("Simple query", func(t *testing.T) {
// 		sql := `SELECT * FROM eth.transactions WHERE date >= date_format(date_add('day', -90, current_date), '%Y-%m-%d') ORDER BY block_timestamp DESC LIMIT 3;`

// 		t.Logf("Context: %+v", ctx)
// 		t.Logf("SQL Query: %s", sql)

// 		result, err := provider.ExecuteQuery(ctx, sql)
// 		if err != nil {
// 			t.Logf("Query execution error: %v", err)
// 			if result != nil {
// 				t.Logf("Partial result: %+v", result)
// 			}
// 			t.FailNow()
// 		}
// 		require.NoError(t, err)
// 		require.NotNil(t, result)
// 		require.True(t, result.Success)
// 		require.NotEmpty(t, result.Data)

// 		// Log the result details
// 		t.Logf("Result metadata: %+v", result.Metadata)
// 		t.Logf("Result data count: %d", len(result.Data))

// 		// Verify the returned data structure
// 		data := result.Data[0].(map[string]interface{})
// 		require.Contains(t, data, "hash")
// 		require.Contains(t, data, "block_number")
// 		require.Contains(t, data, "from_address")
// 		require.Contains(t, data, "to_address")
// 		require.Contains(t, data, "value")
// 	})

// 	// Test complex query
// 	t.Run("Complex query", func(t *testing.T) {
// 		sql := `
// 		WITH daily_stats AS (
// 			SELECT
// 				date_format(block_timestamp, '%Y-%m-%d') as day,
// 				COUNT(*) as tx_count,
// 				AVG(CAST(gas_price AS DOUBLE)) as avg_gas_price,
// 				SUM(CAST(value AS DOUBLE)) as total_value
// 			FROM eth.transactions
// 			WHERE date >= date_format(date_add('day', -7, current_date), '%Y-%m-%d')
// 			GROUP BY date_format(block_timestamp, '%Y-%m-%d')
// 			ORDER BY day DESC
// 		)
// 		SELECT * FROM daily_stats
// 		LIMIT 1`

// 		result, err := provider.ExecuteQuery(ctx, sql)
// 		if err != nil {
// 			t.Logf("Query execution error: %v", err)
// 			t.FailNow()
// 		}
// 		require.NoError(t, err)
// 		require.NotNil(t, result)
// 		require.True(t, result.Success)
// 		require.NotEmpty(t, result.Data)

// 		// Verify the returned data structure
// 		data := result.Data[0].(map[string]interface{})
// 		require.Contains(t, data, "day")
// 		require.Contains(t, data, "tx_count")
// 		require.Contains(t, data, "avg_gas_price")
// 		require.Contains(t, data, "total_value")
// 	})

// 	// Test error cases
// 	t.Run("Invalid SQL", func(t *testing.T) {
// 		sql := "INVALID SQL"
// 		result, err := provider.ExecuteQuery(ctx, sql)
// 		require.Error(t, err)
// 		require.Nil(t, result)
// 		require.Contains(t, err.Error(), "MALFORMED_QUERY")
// 	})

// 	t.Run("Empty SQL", func(t *testing.T) {
// 		sql := ""
// 		result, err := provider.ExecuteQuery(ctx, sql)
// 		require.Error(t, err)
// 		require.Nil(t, result)
// 		require.Contains(t, err.Error(), "invalid SQL query length")
// 	})
// }

// func TestAnalyzeQuery(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration test in short mode")
// 	}

// 	llmConfig := &llm.LLMConfig{
// 		Provider: "deepseek",
// 		APIKey:   os.Getenv("DEEPSEEK_API_KEY"),
// 		BaseURL:  os.Getenv("DEEPSEEK_API_URL"),
// 		Model:    "deepseek-chat",
// 	}

// 	if llmConfig.APIKey == "" {
// 		t.Skip("Skipping test: DEEPSEEK_API_KEY not set")
// 	}

// 	llmClient := llm.NewClient(llmConfig)

// 	provider := actions.NewDatabaseProvider("", "", "ethereum-mainnet", llmClient)

// 	result := &actions.TransactionQueryResult{
// 		Success: true,
// 		Data: []interface{}{
// 			map[string]interface{}{
// 				"block_number": "21783739",
// 				"from_address": "0xde2a3b8017453b4068a7b2a692f2456dabdb7a25",
// 				"hash":         "0x8d39af7135d2791609892a655a68b241d94b2591ed6b990b3fbfd2e927bb55c8",
// 				"to_address":   "0x51c72848c68a965f66fa7a88855f9f7784502a7f",
// 				"value":        "0.0",
// 			},
// 			map[string]interface{}{
// 				"block_number": "21783739",
// 				"from_address": "0x40786472eddbf02e9ce122adef40bda90c20302c",
// 				"hash":         "0x013c65e67a58b17068ae6eb9bba2470ab33a26333beaa102128544dd0e12a8f8",
// 				"to_address":   "0x111111125421ca6dc452d289314280a0f8842a65",
// 				"value":        "0.0",
// 			},
// 			map[string]interface{}{
// 				"block_number": "21783739",
// 				"from_address": "0xde2a3b8017453b4068a7b2a692f2456dabdb7a25",
// 				"hash":         "0x50025504d92bc8a64545c6530cec49fd6a0b2a9f9f7e54cd91794c42df8b35ba",
// 				"to_address":   "0x51c72848c68a965f66fa7a88855f9f7784502a7f",
// 				"value":        "0.0",
// 			},
// 		},
// 		Metadata: struct {
// 			Total         int    `json:"total"`
// 			QueryTime     string `json:"queryTime"`
// 			QueryType     string `json:"queryType"`
// 			ExecutionTime int    `json:"executionTime"`
// 			Cached        bool   `json:"cached"`
// 			QueryDetails  *struct {
// 				Params          actions.FetchTransactionParams `json:"params"`
// 				Query           string                         `json:"query"`
// 				ParamValidation []string                       `json:"paramValidation,omitempty"`
// 			} `json:"queryDetails,omitempty"`
// 			BlockStats       *actions.BlockStats       `json:"blockStats,omitempty"`
// 			TransactionStats *actions.TransactionStats `json:"transactionStats,omitempty"`
// 		}{
// 			Total:     3,
// 			QueryTime: time.Now().Format(time.RFC3339),
// 			QueryType: "transaction",
// 		},
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
// 	defer cancel()

// 	analysis, err := provider.AnalyzeQuery(ctx, result)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, analysis)

// 	// Verify analysis contains key sections
// 	require.Contains(t, analysis, "Transaction Overview")
// 	require.Contains(t, analysis, "Value Analysis")
// 	require.Contains(t, analysis, "Gas and Network Analysis")
// 	require.Contains(t, analysis, "Address Activity")
// 	require.Contains(t, analysis, "Technical Insights")
// 	require.Contains(t, analysis, "Risk and Security")
// }

// func TestTransformAPIResponse(t *testing.T) {
// 	provider := helper.SetupTestProvider(t)
// 	providerImpl, ok := provider.(*actions.DatabaseProviderImpl)
// 	require.True(t, ok, "provider should be of type *DatabaseProviderImpl")

// 	tests := []struct {
// 		name     string
// 		input    *actions.APIResponse
// 		expected []interface{}
// 	}{
// 		{
// 			name: "Empty response",
// 			input: &actions.APIResponse{
// 				Data: struct {
// 					ColumnInfos []string `json:"column_infos"`
// 					Rows        []struct {
// 						Items []interface{} `json:"items"`
// 					} `json:"rows"`
// 				}{
// 					ColumnInfos: []string{},
// 					Rows:        nil,
// 				},
// 			},
// 			expected: []interface{}{},
// 		},
// 		{
// 			name: "Single row",
// 			input: &actions.APIResponse{
// 				Data: struct {
// 					ColumnInfos []string `json:"column_infos"`
// 					Rows        []struct {
// 						Items []interface{} `json:"items"`
// 					} `json:"rows"`
// 				}{
// 					ColumnInfos: []string{"hash", "value"},
// 					Rows: []struct {
// 						Items []interface{} `json:"items"`
// 					}{
// 						{
// 							Items: []interface{}{"0x123", 1.5},
// 						},
// 					},
// 				},
// 			},
// 			expected: []interface{}{
// 				map[string]interface{}{
// 					"hash":  "0x123",
// 					"value": 1.5,
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got := providerImpl.TransformAPIResponse(tt.input)
// 			assert.Equal(t, tt.expected, got)
// 		})
// 	}
// }

// // Helper functions for creating pointers to values
// func strPtr(s string) *string {
// 	return &s
// }

// func intPtr(i int) *int {
// 	return &i
// }
