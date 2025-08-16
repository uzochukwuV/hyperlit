For developers
API
Info endpoint
Perpetuals
The section documents the info endpoints that are specific to perpetuals. See Rate limits section for rate limiting logic and weights.

Retrieve all perpetual dexs
POST https://api.hyperliquid.xyz/info

Headers
Name
Type
Description
Content-Type*

String

"application/json"

Request Body
Name
Type
Description
type*

String

"perpDexs"

200: OK Successful Response
Copy
[
  null,
  {
    "name": "test",
    "full_name": "test dex",
    "deployer": "0x5e89b26d8d66da9888c835c9bfcc2aa51813e152",
    "oracle_updater": null
  }
]
Retrieve perpetuals metadata (universe and margin tables)
POST https://api.hyperliquid.xyz/info

Headers
Name
Type
Description
Content-Type*

String

"application/json"

Request Body
Name
Type
Description
type*

String

"meta"

dex

String

Perp dex name. Defaults to the empty string which represents the first perp dex.

200: OK Successful Response
Copy
{
    "universe": [
        {
            "name": "BTC",
            "szDecimals": 5,
            "maxLeverage": 50
        },
        {
            "name": "ETH",
            "szDecimals": 4,
            "maxLeverage": 50
        },
        {
            "name": "HPOS",
            "szDecimals": 0,
            "maxLeverage": 3,
            "onlyIsolated": true
        },
        {
            "name": "LOOM",
            "szDecimals": 1,
            "maxLeverage": 3,
            "onlyIsolated": true,
            "isDelisted": true
        }
    ],
    "marginTables": [
        [
            50,
            {
                "description": "",
                "marginTiers": [
                    {
                        "lowerBound": "0.0",
                        "maxLeverage": 50
                    }
                ]
            }
        ],
        [
            51,
            {
                "description": "tiered 10x",
                "marginTiers": [
                    {
                        "lowerBound": "0.0",
                        "maxLeverage": 10
                    },
                    {
                        "lowerBound": "3000000.0",
                        "maxLeverage": 5
                    }
                ]
            }
        ]
    ]
}
Retrieve perpetuals asset contexts (includes mark price, current funding, open interest, etc.)
POST https://api.hyperliquid.xyz/info

Headers
Name
Type
Description
Content-Type*

String

"application/json"

Request Body
Name
Type
Description
type*

String

"metaAndAssetCtxs"

200: OK Successful Response
Copy
[
{
     "universe": [
        {
            "name": "BTC",
            "szDecimals": 5,
            "maxLeverage": 50
        },
        {
            "name": "ETH",
            "szDecimals": 4,
            "maxLeverage": 50
        },
        {
            "name": "HPOS",
            "szDecimals": 0,
            "maxLeverage": 3,
            "onlyIsolated": true
        }
    ]
},
[
    {
        "dayNtlVlm":"1169046.29406",
         "funding":"0.0000125",
         "impactPxs":[
            "14.3047",
            "14.3444"
         ],
         "markPx":"14.3161",
         "midPx":"14.314",
         "openInterest":"688.11",
         "oraclePx":"14.32",
         "premium":"0.00031774",
         "prevDayPx":"15.322"
    },
    {
         "dayNtlVlm":"1426126.295175",
         "funding":"0.0000125",
         "impactPxs":[
            "6.0386",
            "6.0562"
         ],
         "markPx":"6.0436",
         "midPx":"6.0431",
         "openInterest":"1882.55",
         "oraclePx":"6.0457",
         "premium":"0.00028119",
         "prevDayPx":"6.3611"
      },
      {
         "dayNtlVlm":"809774.565507",
         "funding":"0.0000125",
         "impactPxs":[
            "8.4505",
            "8.4722"
         ],
         "markPx":"8.4542",
         "midPx":"8.4557",
         "openInterest":"2912.05",
         "oraclePx":"8.4585",
         "premium":"0.00033694",
         "prevDayPx":"8.8097"
      }
]
]
Retrieve user's perpetuals account summary
POST https://api.hyperliquid.xyz/info

See a user's open positions and margin summary for perpetuals trading

Headers
Name
Type
Description
Content-Type*

"application/json"

Request Body
Name
Type
Description
type*

String

"clearinghouseState"

user*

String

Onchain address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

dex

String

Perp dex name. Defaults to the empty string which represents the first perp dex.

200: OK Successful Response
Copy
{
  "assetPositions": [
    {
      "position": {
        "coin": "ETH",
        "cumFunding": {
          "allTime": "514.085417",
          "sinceChange": "0.0",
          "sinceOpen": "0.0"
        },
        "entryPx": "2986.3",
        "leverage": {
          "rawUsd": "-95.059824",
          "type": "isolated",
          "value": 20
        },
        "liquidationPx": "2866.26936529",
        "marginUsed": "4.967826",
        "maxLeverage": 50,
        "positionValue": "100.02765",
        "returnOnEquity": "-0.0026789",
        "szi": "0.0335",
        "unrealizedPnl": "-0.0134"
      },
      "type": "oneWay"
    }
  ],
  "crossMaintenanceMarginUsed": "0.0",
  "crossMarginSummary": {
    "accountValue": "13104.514502",
    "totalMarginUsed": "0.0",
    "totalNtlPos": "0.0",
    "totalRawUsd": "13104.514502"
  },
  "marginSummary": {
    "accountValue": "13109.482328",
    "totalMarginUsed": "4.967826",
    "totalNtlPos": "100.02765",
    "totalRawUsd": "13009.454678"
  },
  "time": 1708622398623,
  "withdrawable": "13104.514502"
}
Retrieve a user's funding history or non-funding ledger updates
POST https://api.hyperliquid.xyz/info

Note: Non-funding ledger updates include deposits, transfers, and withdrawals.

Headers
Name
Type
Description
Content-Type*

String

"application/json"

Request Body
Name
Type
Description
type*

String

"userFunding" or "userNonFundingLedgerUpdates"

user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

startTime*

int

Start time in milliseconds, inclusive

endTime

int

End time in milliseconds, inclusive. Defaults to current time.

200: OK Successful Response
Copy
[
    {
        "delta": {
            "coin":"ETH",
            "fundingRate":"0.0000417",
            "szi":"49.1477",
            "type":"funding",
            "usdc":"-3.625312"
        },
        "hash":"0xa166e3fa63c25663024b03f2e0da011a00307e4017465df020210d3d432e7cb8",
        "time":1681222254710
    },
    ...
]
Retrieve historical funding rates
POST https://api.hyperliquid.xyz/info

Headers
Name
Type
Description
Content-Type*

String

"application/json"

Request Body
Name
Type
Description
type*

String

"fundingHistory"

coin*

String

Coin, e.g. "ETH"

startTime*

int

Start time in milliseconds, inclusive

endTime

int

End time in milliseconds, inclusive. Defaults to current time.

200: OK
Copy
[
    {
        "coin":"ETH",
        "fundingRate": "-0.00022196",
        "premium": "-0.00052196",
        "time":1683849600076
    }
]
Retrieve predicted funding rates for different venues
POST https://api.hyperliquid.xyz/info

Headers
Name
Type
Description
Content-Type*

String

"application/json"

Request Body
Name
Type
Description
type*

String

"predictedFundings"

200: OK Successful Response
Copy
[
  [
    "AVAX",
    [
      [
        "BinPerp",
        {
          "fundingRate": "0.0001",
          "nextFundingTime": 1733961600000
        }
      ],
      [
        "HlPerp",
        {
          "fundingRate": "0.0000125",
          "nextFundingTime": 1733958000000
        }
      ],
      [
        "BybitPerp",
        {
          "fundingRate": "0.0001",
          "nextFundingTime": 1733961600000
        }
      ]
    ]
  ],...
]
Query perps at open interest caps
POST https://api.hyperliquid.xyz/info

Headers
Name
Type
Description
Content-Type*

String

"application/json"

Request Body
Name
Type
Description
type*

String

"perpsAtOpenInterestCap"

200: OK Successful Response
Copy
["BADGER","CANTO","FTM","LOOM","PURR"]
Retrieve information about the Perp Deploy Auction
POST https://api.hyperliquid.xyz/info

Headers
Name
Type
Description
Content-Type*

String

"application/json"

Request Body
Name
Type
Description
type*

String

"perpDeployAuctionStatus"

200: OK Successful Response
Copy
{
  "startTimeSeconds": 1747656000,
  "durationSeconds": 111600,
  "startGas": "500.0",
  "currentGas": "500.0",
  "endGas": null
}
Retrieve User's Active Asset Data
POST https://api.hyperliquid.xyz/info

Headers
Name
Type
Description
Content-Type*

String

"application/json"

Request Body
Name
Type
Description
type*

String

"activeAssetData"

user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

coin*

String

Coin, e.g. "ETH". See here for more details.

200: OK
Copy
{
  "user": "0xb65822a30bbaaa68942d6f4c43d78704faeabbbb",
  "coin": "APT",
  "leverage": {
    "type": "cross",
    "value": 3
  },
  "maxTradeSzs": ["24836370.4400000013", "24836370.4400000013"],
  "availableToTrade": ["37019438.0284740031", "37019438.0284740031"],
  "markPx": "4.4716"
}