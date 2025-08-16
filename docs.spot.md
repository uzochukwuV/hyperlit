For developers
API
Info endpoint
Spot
The section documents the info endpoints that are specific to spot.

Retrieve spot metadata
POST https://api.hyperliquid.xyz/info

Headers

Name
Value
Content-Type*

"application/json"

Body

Name
Type
Description
type*

String

"spotMeta"

Response

200: OK Successful Response
Copy
{
    "tokens": [
        {
            "name": "USDC",
            "szDecimals": 8,
            "weiDecimals" 8,
            "index": 0,
            "tokenId": "0x6d1e7cde53ba9467b783cb7c530ce054",
            "isCanonical": true,
            "evmContract":null,
            "fullName":null
        },
        {
            "name": "PURR",
            "szDecimals": 0,
            "weiDecimals": 5,
            "index": 1,
            "tokenId": "0xc1fb593aeffbeb02f85e0308e9956a90",
            "isCanonical": true,
            "evmContract":null,
            "fullName":null
        },
        {
            "name": "HFUN",
            "szDecimals": 2,
            "weiDecimals": 8,
            "index": 2,
            "tokenId": "0xbaf265ef389da684513d98d68edf4eae",
            "isCanonical": false,
            "evmContract":null,
            "fullName":null
        },
    ],
    "universe": [
        {
            "name": "PURR/USDC",
            "tokens": [1, 0],
            "index": 0,
            "isCanonical": true
        },
        {
            "tokens": [2, 0],
            "name": "@1",
            "index": 1,
            "isCanonical": false
        },
    ]
}
Retrieve spot asset contexts
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

"spotMetaAndAssetCtxs"

200: OK Successful Response
Copy
[
{
    "tokens": [
        {
            "name": "USDC",
            "szDecimals": 8,
            "weiDecimals" 8,
            "index": 0,
            "tokenId": "0x6d1e7cde53ba9467b783cb7c530ce054",
            "isCanonical": true,
            "evmContract":null,
            "fullName":null
        },
        {
            "name": "PURR",
            "szDecimals": 0,
            "weiDecimals": 5,
            "index": 1,
            "tokenId": "0xc1fb593aeffbeb02f85e0308e9956a90",
            "isCanonical": true,
            "evmContract":null,
            "fullName":null
        }
    ],
    "universe": [
        {
            "name": "PURR/USDC",
            "tokens": [1, 0],
            "index": 0,
            "isCanonical": true
        }
    ]
},
[
    {
        "dayNtlVlm":"8906.0",
        "markPx":"0.14",
        "midPx":"0.209265",
        "prevDayPx":"0.20432"
    }
]
]
Retrieve a user's token balances
POST https://api.hyperliquid.xyz/info

See a user's token balances

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

"spotClearinghouseState"

user*

String

Onchain address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK Successful Response
Copy
{
    "balances": [
        {
            "coin": "USDC",
            "token": 0,
            "hold": "0.0",
            "total": "14.625485",
            "entryNtl": "0.0"
        },
        {
            "coin": "PURR",
            "token": 1,
            "hold": "0",
            "total": "2000",
            "entryNtl": "1234.56",
        }
    ]
}
Retrieve information about the Spot Deploy Auction
POST https://api.hyperliquid.xyz/info

Headers

Name
Value
Content-Type*

"application/json"

Body

Name
Type
Description
type*

String

"spotDeployState"

user*

String

Onchain address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

Response

200: OK Successful Response
Copy
{
  "states": [
    {
      "token": 150,
      "spec" : {
        "name": "HYPE",
        "szDecimals": 2,
        "weiDecimals": 8,
      },
      "fullName": "Hyperliquid",
      "spots": [107],
      "maxSupply": 1000000000,
      "hyperliquidityGenesisBalance": "120000",
      "totalGenesisBalanceWei": "100000000000000000",
      "userGenesisBalances": [
        ("0xdddddddddddddddddddddddddddddddddddddddd", "428,062,211")...
      ],
      "existingTokenGenesisBalances": [
        (1, "0")...
      ]
    }
  ],
  "gasAuction": {
    "startTimeSeconds": 1733929200,
    "durationSeconds": 111600,
    "startGas": "181305.90046",
    "currentGas": null,
    "endGas": "181291.247358"
  }
}
Retrieve information about a token
POST https://api.hyperliquid.xyz/info

Headers

Name
Value
Content-Type*

"application/json"

Body

Name
Type
Description
type*

String

"tokenDetails"

tokenId*

String

Onchain id in 34-character hexadecimal format; e.g. 0x00000000000000000000000000000000.

Response

200: OK Successful Response
Copy
{
  "name": "TEST",
  "maxSupply": "1852229076.12716007",
  "totalSupply": "851681534.05516005",
  "circulatingSupply": "851681534.05516005",
  "szDecimals": 0,
  "weiDecimals": 5,
  "midPx": "3.2049",
  "markPx": "3.2025",
  "prevDayPx": "3.2025",
  "genesis": {
    "userBalances": [
      [
        "0x0000000000000000000000000000000000000001",
        "1000000000.0"
      ],
      [
        "0xffffffffffffffffffffffffffffffffffffffff",
        "1000000000.0"
      ]
    ],
    "existingTokenBalances": []
  },
  "deployer": "0x0000000000000000000000000000000000000001",
  "deployGas": "100.0",
  "deployTime": "2024-06-05T10:50:59.434",
  "seededUsdc": "0.0",
  "nonCirculatingUserBalances": [],
  "futureEmissions": "0.0"
}
