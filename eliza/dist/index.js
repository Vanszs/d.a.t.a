var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/index.ts
var index_exports = {};
__export(index_exports, {
  onchainDataPlugin: () => onchainDataPlugin
});
module.exports = __toCommonJS(index_exports);

// src/actions/fetchTransaction.ts
var import_core = require("@elizaos/core");
var fetchTransactionAction = {
  name: "fetch_transactions",
  description: "Fetch Ethereum transactions based on various criteria",
  similes: [
    "get transactions",
    "show transfers",
    "display eth transactions",
    "find transactions",
    "search transfers",
    "check transactions",
    "view transfers",
    "list transactions",
    "recent transactions",
    "transaction history",
    "transfer records",
    "eth movements",
    "wallet activity",
    "transaction lookup",
    "transfer search"
  ],
  examples: [
    [
      {
        user: "user",
        content: {
          text: "Show me the latest 10 Ethereum transactions",
          action: "FETCH_TRANSACTIONS"
        }
      }
    ],
    [
      {
        user: "user",
        content: {
          text: "Get transactions for address 0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
          action: "FETCH_TRANSACTIONS"
        }
      }
    ],
    [
      {
        user: "user",
        content: {
          text: "Show me transactions above 10 ETH from last week",
          action: "FETCH_TRANSACTIONS"
        }
      }
    ],
    [
      {
        user: "user",
        content: {
          text: "Find failed transactions with highest gas fees",
          action: "FETCH_TRANSACTIONS"
        }
      }
    ],
    [
      {
        user: "user",
        content: {
          text: "Show me transactions between blocks 1000000 and 1000100",
          action: "FETCH_TRANSACTIONS"
        }
      }
    ]
  ],
  validate: async (runtime) => {
    return true;
  },
  handler: async (runtime, message, state, _options, callback) => {
    try {
      import_core.elizaLogger.log("$$$$ Fetching Ethereum transactions...");
      import_core.elizaLogger.log("$$$$message", message);
      if (callback) {
        const mockText = "This is the transaction details of ethereum";
        const mockQuery = "SELECT * FROM ethereum_transactions";
        const mockParams = {
          limit: 10,
          orderBy: "timestamp",
          orderDirection: "DESC"
        };
        callback({
          // text: `Here's the SQL query to retrieve Ethereum transactions:\n${sqlQuery}\nThis query will return the specified transactions, including details like transaction hash, block number, sender/receiver addresses, value, and gas used. Let me know if you'd like further analysis or specific details about any of these transactions!`,
          // content: {
          //     success: true,
          //     query: sqlQuery,
          //     params: queryParams,
          // },
          text: mockText,
          content: {
            success: true,
            query: mockQuery,
            params: mockParams
          }
        });
      }
      return true;
    } catch (error) {
      import_core.elizaLogger.error("Error in fetch transaction action:", error);
      if (callback) {
        callback({
          text: `Error fetching transactions: ${error.message}`,
          content: { error: error.message }
        });
      }
      return false;
    }
  }
};

// src/providers/ethereum/database.ts
var import_core2 = require("@elizaos/core");
var DatabaseProvider = class {
  chain;
  API_URL = "https://dev-interface.carv.io/ai-agent-backend/sql_query";
  // fake data
  MOCK_RESPONSE = {
    code: 0,
    msg: "Success",
    data: {
      column_infos: [
        "hash",
        "nonce",
        "transaction_index",
        "from_address",
        "to_address",
        "value",
        "gas",
        "gas_price",
        "input",
        "receipt_cumulative_gas_used",
        "receipt_gas_used",
        "receipt_contract_address",
        "receipt_root",
        "receipt_status",
        "block_timestamp",
        "block_number",
        "block_hash",
        "max_fee_per_gas",
        "max_priority_fee_per_gas",
        "transaction_type",
        "receipt_effective_gas_price",
        "date"
      ],
      rows: [
        {
          items: [
            "0xb9f2c4dd816305a29471f7e843f33b8a4f52c24bfac58dba5f6dd703bdcc347d",
            "131",
            "0",
            "0xb7b3690efa6b3f08d4ec289ff655c4b7bb15ee39",
            "0x32be343b94f860124dc4fee278fdcbd38c102d88",
            "5.06132703E18",
            "21000",
            "58587049895",
            "0x",
            "21000",
            "21000",
            "",
            "",
            "1",
            "2015-10-18 09:01:42.000",
            "401609",
            "0xab8cf7f52769cb62a8a970347a116368c2ae08581d412573dd09a8a4af99b4cd",
            "0",
            "0",
            "0",
            "58587049895",
            "2015-10-18"
          ]
        },
        {
          items: [
            "0xbca5c575aa36f6164dbd98bf7c1008791d615cb0f255c19db7c6b45b06367ba0",
            "9573",
            "0",
            "0x2a65aca4d5fc5b5c859090a6c34d164135398226",
            "0xff3a70d8d5692dd05d71175fa29e8565f7450f57",
            "2.7443251E18",
            "90000",
            "50000000000",
            "0x",
            "21000",
            "21000",
            "",
            "",
            "1",
            "2015-10-18 12:21:43.000",
            "402253",
            "0x290dfc39bec9918fca28c8cebe2beaa19f9303f3b432d62d1e3409b1195556d6",
            "0",
            "0",
            "0",
            "50000000000",
            "2015-10-18"
          ]
        }
      ]
    }
  };
  constructor(chain) {
    this.chain = chain;
  }
  extractSQLQuery(preResponse) {
    try {
      let jsonData = preResponse;
      if (typeof preResponse === "string") {
        try {
          jsonData = JSON.parse(preResponse);
        } catch (e) {
          import_core2.elizaLogger.error(
            "Failed to parse preResponse as JSON:",
            e
          );
          return null;
        }
      }
      const findSQLQuery = (obj) => {
        if (!obj) return null;
        if (typeof obj === "string") {
          const sqlPattern = /^\s*(SELECT|WITH)\s+[\s\S]+?(?:;|$)/i;
          const commentPattern = /--.*$|\/\*[\s\S]*?\*\//gm;
          const cleanStr = obj.replace(commentPattern, "").trim();
          if (sqlPattern.test(cleanStr)) {
            const unsafeKeywords = [
              "drop",
              "delete",
              "update",
              "insert",
              "alter",
              "create"
            ];
            const isUnsafe = unsafeKeywords.some(
              (keyword) => cleanStr.toLowerCase().includes(keyword)
            );
            if (!isUnsafe) {
              return cleanStr;
            }
          }
          return null;
        }
        if (Array.isArray(obj)) {
          for (const item of obj) {
            const result = findSQLQuery(item);
            if (result) return result;
          }
          return null;
        }
        if (typeof obj === "object") {
          for (const key of Object.keys(obj)) {
            if (key.toLowerCase() === "query" && obj.sql) {
              const result = findSQLQuery(obj[key]);
              if (result) return result;
            }
          }
          for (const key of Object.keys(obj)) {
            const result = findSQLQuery(obj[key]);
            if (result) return result;
          }
        }
        return null;
      };
      const sqlQuery = findSQLQuery(jsonData);
      if (!sqlQuery) {
        import_core2.elizaLogger.warn("No valid SQL query found in preResponse");
        return null;
      }
      return sqlQuery;
    } catch (error) {
      import_core2.elizaLogger.error("Error in extractSQLQuery:", error);
      return null;
    }
  }
  async sendSqlQuery(sql, mock = false) {
    if (mock) {
      import_core2.elizaLogger.log("Using mock data for SQL query");
      return this.MOCK_RESPONSE;
    }
    try {
      const response = await fetch(this.API_URL, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          sql_content: sql
        })
      });
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data = await response.json();
      return data;
    } catch (error) {
      import_core2.elizaLogger.error("Error sending SQL query to API:", error);
      throw error;
    }
  }
  // Transform API response data
  transformApiResponse(apiResponse) {
    const { column_infos, rows } = apiResponse.data;
    return rows.map((row) => {
      const rowData = {};
      row.items.forEach((value, index) => {
        const columnName = column_infos[index];
        rowData[columnName] = value;
      });
      return rowData;
    });
  }
  // Execute query
  async executeQuery(sql) {
    try {
      if (!sql || sql.length > 5e3) {
        throw new Error("Invalid SQL query length");
      }
      const queryType = sql.toLowerCase().includes("token_transfers") ? "token" : sql.toLowerCase().includes("count") ? "aggregate" : "transaction";
      const apiResponse = await this.sendSqlQuery(sql);
      if (apiResponse.code !== 0) {
        throw new Error(`API Error: ${apiResponse.msg}`);
      }
      const transformedData = this.transformApiResponse(apiResponse);
      const queryResult = {
        success: true,
        data: transformedData,
        metadata: {
          total: transformedData.length,
          queryTime: (/* @__PURE__ */ new Date()).toISOString(),
          queryType,
          executionTime: 0,
          cached: false
        }
      };
      return queryResult;
    } catch (error) {
      import_core2.elizaLogger.error("Query execution failed:", error);
      return {
        success: false,
        data: [],
        metadata: {
          total: 0,
          queryTime: (/* @__PURE__ */ new Date()).toISOString(),
          queryType: "unknown",
          executionTime: 0,
          cached: false
        },
        error: {
          code: error.code || "EXECUTION_ERROR",
          message: error.message || "Unknown error occurred",
          details: error
        }
      };
    }
  }
  async query(sql) {
    return this.executeQuery(sql);
  }
  getDatabaseSchema() {
    return `
        CREATE EXTERNAL TABLE transactions(
            hash string,
            nonce bigint,
            block_hash string,
            block_number bigint,
            block_timestamp timestamp,
            date string,
            transaction_index bigint,
            from_address string,
            to_address string,
            value double,
            gas bigint,
            gas_price bigint,
            input string,
            max_fee_per_gas bigint,
            max_priority_fee_per_gas bigint,
            transaction_type bigint
        ) PARTITIONED BY (date string)
        ROW FORMAT SERDE 'org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe'
        STORED AS INPUTFORMAT 'org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat'
        OUTPUTFORMAT 'org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat';

        CREATE EXTERNAL TABLE token_transfers(
            token_address string,
            from_address string,
            to_address string,
            value double,
            transaction_hash string,
            log_index bigint,
            block_timestamp timestamp,
            date string,
            block_number bigint,
            block_hash string
        ) PARTITIONED BY (date string)
        ROW FORMAT SERDE 'org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe'
        STORED AS INPUTFORMAT 'org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat'
        OUTPUTFORMAT 'org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat';
        `;
  }
  getQueryExamples() {
    return `
        Common Query Examples:

        1. Find Most Active Addresses in Last 7 Days:
        WITH address_activity AS (
            SELECT
                from_address AS address,
                COUNT(*) AS tx_count
            FROM
                eth.transactions
            WHERE date_parse(date, '%Y-%m-%d') >= date_add('day', -7, current_date)
            GROUP BY
                from_address
            UNION ALL
            SELECT
                to_address AS address,
                COUNT(*) AS tx_count
            FROM
                eth.transactions
            WHERE
                date_parse(date, '%Y-%m-%d') >= date_add('day', -7, current_date)
            GROUP BY
                to_address
        )
        SELECT
            address,
            SUM(tx_count) AS total_transactions
        FROM
            address_activity
        GROUP BY
            address
        ORDER BY
            total_transactions DESC
        LIMIT 10;

        2. Analyze Address Transaction Statistics (Last 30 Days):
        WITH recent_transactions AS (
            SELECT
                from_address,
                to_address,
                value,
                block_timestamp,
                CASE
                    WHEN from_address = :address THEN 'outgoing'
                    WHEN to_address = :address THEN 'incoming'
                    ELSE 'other'
                END AS transaction_type
            FROM eth.transactions
            WHERE date >= date_format(date_add('day', -30, current_date), '%Y-%m-%d')
                AND (from_address = :address OR to_address = :address)
        )
        SELECT
            transaction_type,
            COUNT(*) AS transaction_count,
            SUM(CASE WHEN transaction_type = 'outgoing' THEN value ELSE 0 END) AS total_outgoing_value,
            SUM(CASE WHEN transaction_type = 'incoming' THEN value ELSE 0 END) AS total_incoming_value
        FROM recent_transactions
        GROUP BY transaction_type;

        3. Token Transfer Analysis:
        WITH filtered_transactions AS (
            SELECT
                token_address,
                from_address,
                to_address,
                value,
                block_timestamp
            FROM eth.token_transfers
            WHERE token_address = :token_address
                AND date >= :start_date
        )
        SELECT
            COUNT(*) AS transaction_count,
            SUM(value) AS total_transaction_value,
            MAX(value) AS max_transaction_value,
            MIN(value) AS min_transaction_value,
            MAX_BY(from_address, value) AS max_value_from_address,
            MAX_BY(to_address, value) AS max_value_to_address,
            MIN_BY(from_address, value) AS min_value_from_address,
            MIN_BY(to_address, value) AS min_value_to_address
        FROM filtered_transactions;

        Note: Replace :address, :token_address, and :start_date with actual values when querying.
        `;
  }
  getQueryTemplate() {
    return `
        # Database Schema
        {{databaseSchema}}

        # Query Examples
        {{queryExamples}}

        # User's Query
        {{userQuery}}

        # Query Guidelines:
        1. Time Range Requirements:
           - ALWAYS include time range limitations in queries
           - Default to last 3 months if no specific time range is mentioned
           - Use date_parse(date, '%Y-%m-%d') >= date_add('month', -3, current_date) for default time range
           - Adjust time range based on user's specific requirements

        2. Query Optimization:
           - Include appropriate LIMIT clauses
           - Use proper indexing columns (date, address, block_number)
           - Consider partitioning by date
           - Add WHERE clauses for efficient filtering

        3. Response Format Requirements:
           You MUST respond in the following JSON format:
           {
             "sql": {
               "query": "your SQL query string",
               "explanation": "brief explanation of the query",
               "timeRange": "specified time range in the query"
             },
             "analysis": {
               "overview": {
                 "totalTransactions": "number",
                 "timeSpan": "time period covered",
                 "keyMetrics": ["list of important metrics"]
               },
               "patterns": {
                 "transactionPatterns": ["identified patterns"],
                 "addressBehavior": ["address analysis"],
                 "temporalTrends": ["time-based trends"]
               },
               "statistics": {
                 "averages": {},
                 "distributions": {},
                 "anomalies": []
               },
               "insights": ["key insights from the data"],
               "recommendations": ["suggested actions or areas for further investigation"]
             }
           }

        4. Analysis Requirements:
           - Focus on recent data patterns
           - Identify trends and anomalies
           - Provide statistical analysis
           - Include risk assessment
           - Suggest further investigations

        Example Response:
        {
          "sql": {
            "query": "WITH recent_txs AS (SELECT * FROM eth.transactions WHERE date_parse(date, '%Y-%m-%d') >= date_add('month', -3, current_date))...",
            "explanation": "Query fetches last 3 months of transactions with aggregated metrics",
            "timeRange": "Last 3 months"
          },
          "analysis": {
            "overview": {
              "totalTransactions": 1000000,
              "timeSpan": "2024-01-01 to 2024-03-12",
              "keyMetrics": ["Average daily transactions: 11000", "Peak day: 2024-02-15"]
            },
            "patterns": {
              "transactionPatterns": ["High volume during Asian trading hours", "Weekend dips in activity"],
              "addressBehavior": ["5 addresses responsible for 30% of volume", "Increasing DEX activity"],
              "temporalTrends": ["Growing transaction volume", "Decreasing gas costs"]
            },
            "statistics": {
              "averages": {
                "dailyTransactions": 11000,
                "gasPrice": "25 gwei"
              },
              "distributions": {
                "valueRanges": ["0-1 ETH: 60%", "1-10 ETH: 30%", ">10 ETH: 10%"]
              },
              "anomalies": ["Unusual spike in gas prices on 2024-02-01"]
            },
            "insights": [
              "Growing DeFi activity indicated by smart contract interactions",
              "Whale addresses showing increased accumulation"
            ],
            "recommendations": [
              "Monitor growing gas usage trend",
              "Track new active addresses for potential market signals"
            ]
          }
        }
        `;
  }
  getAnalysisInstruction() {
    return `
            1. Data Overview:
                - Analyze the overall pattern in the query results
                - Identify key metrics and their significance
                - Note any unusual or interesting patterns

            2. Transaction Analysis:
                - Examine transaction values and their distribution
                - Analyze gas usage patterns
                - Evaluate transaction frequency and timing
                - Identify significant transactions or patterns

            3. Address Behavior:
                - Analyze address interactions
                - Identify frequent participants
                - Evaluate transaction patterns for specific addresses
                - Note any suspicious or interesting behavior

            4. Temporal Patterns:
                - Analyze time-based patterns
                - Identify peak activity periods
                - Note any temporal anomalies
                - Consider seasonal or cyclical patterns

            5. Token Analysis (if applicable):
                - Examine token transfer patterns
                - Analyze token holder behavior
                - Evaluate token concentration
                - Note significant token movements

            6. Statistical Insights:
                - Provide relevant statistical measures
                - Compare with typical blockchain metrics
                - Highlight significant deviations
                - Consider historical context

            7. Risk Assessment:
                - Identify potential suspicious activities
                - Note any unusual patterns
                - Flag potential security concerns
                - Consider regulatory implications

            Please provide a comprehensive analysis of the Ethereum blockchain data based on these ethereum information.
            Focus on significant patterns, anomalies, and insights that would be valuable for understanding the blockchain activity.
            Use technical blockchain terminology and provide specific examples from the data to support your analysis.

            Note: This analysis is based on simulated data for demonstration purposes.
        `;
  }
};
var databaseProvider = (runtime) => {
  const chain = "ethereum-mainnet";
  return new DatabaseProvider(chain);
};
var ethereumDataProvider = {
  get: async (runtime, message, state) => {
    import_core2.elizaLogger.log("%%%% Pis Retrieving from ethereum data provider...");
    try {
      const provider = databaseProvider(runtime);
      const schema = provider.getDatabaseSchema();
      const examples = provider.getQueryExamples();
      const template = provider.getQueryTemplate();
      if (!state) {
        state = await runtime.composeState(message);
      } else {
        state = await runtime.updateRecentMessageState(state);
      }
      import_core2.elizaLogger.log("%%%%&& Pis Context:", message.content.text);
      const buildContext = template.replace("{{databaseSchema}}", schema).replace("{{queryExamples}}", examples).replace("{{userQuery}}", message.content.text || "");
      const context = JSON.stringify({
        user: runtime.agentId,
        content: buildContext,
        action: "fetch_transactions"
      });
      import_core2.elizaLogger.log("%%%% Pis Generated database context");
      const preResponse = await (0, import_core2.generateMessageResponse)({
        runtime,
        context,
        modelClass: import_core2.ModelClass.LARGE
      });
      const userMessage = {
        agentId: runtime.agentId,
        roomId: message.roomId,
        userId: message.userId,
        content: message.content
      };
      const preResponseMessage = {
        id: (0, import_core2.stringToUuid)(message.id + "-" + runtime.agentId),
        ...userMessage,
        userId: runtime.agentId,
        content: preResponse,
        embedding: (0, import_core2.getEmbeddingZeroVector)(),
        createdAt: Date.now()
      };
      await runtime.messageManager.createMemory(preResponseMessage);
      await runtime.updateRecentMessageState(state);
      import_core2.elizaLogger.log("**** Pis preResponse", preResponse);
      const sqlQuery = provider.extractSQLQuery(preResponse);
      if (sqlQuery) {
        import_core2.elizaLogger.log("%%%% Found SQL query:", sqlQuery);
        const analysisInstruction = provider.getAnalysisInstruction();
        try {
          const queryResult = await provider.query(sqlQuery);
          import_core2.elizaLogger.log("%%%% Pis queryResult", queryResult);
          return `
                    # query by user
                    ${message.content.text}

                    # query result
                    ${JSON.stringify(queryResult, null, 2)}

                    # Analysis Instructions
                    ${analysisInstruction}
                    `;
        } catch (error) {
          import_core2.elizaLogger.error("Error executing query:", error);
          return context;
        }
      } else {
        import_core2.elizaLogger.log("%%%% Pis no SQL query found");
      }
      return context;
    } catch (error) {
      import_core2.elizaLogger.error("Error in ethereum data provider:", error);
      return null;
    }
  }
};

// src/index.ts
var onchainDataPlugin = {
  name: "onchain data plugin",
  description: "Enables onchain data fetching",
  actions: [fetchTransactionAction],
  providers: [ethereumDataProvider],
  evaluators: [],
  // separate examples will be added for services and clients
  // services: [new DataService()],
  services: [],
  clients: []
};
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  onchainDataPlugin
});
