For developers
API
Info endpoint
The info endpoint is used to fetch information about the exchange and specific users. The different request bodies result in different corresponding response body schemas.

Pagination
Responses that take a time range will only return 500 elements or distinct blocks of data. To query larger ranges, use the last returned timestamp as the next startTime for pagination.

Perpetuals vs Spot
The endpoints in this section as well as websocket subscriptions work for both Perpetuals and Spot. For perpetuals coin is the name returned in the meta response. For Spot, coin should be PURR/USDC for PURR, and @{index} e.g. @1 for all other spot tokens where index is the index of the spot pair in the universe field of the spotMeta response. For example, the spot index for HYPE on mainnet is @107 because the token index of HYPE is 150 and the spot pair @107 has tokens [150, 0]. Note that some assets may be remapped on user interfaces. For example, BTC/USDC on app.hyperliquid.xyz corresponds to UBTC/USDC on mainnet HyperCore. The L1 name on the token details page can be used to detect remappings.

User address
To query the account data associated with a master or sub-account, you must pass in the actual address of that account. A common pitfall is to use an agent wallet's address which leads to an empty result.

Retrieve mids for all coins
POST https://api.hyperliquid.xyz/info

Note that if the book is empty, the last trade price will be used as a fallback

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

"allMids"

dex

String

Perp dex name. Defaults to the empty string which represents the first perp dex. Spot mids are only included with the first perp dex..

200: OK Successful Response
Copy
{
    "APE": "4.33245",
    "ARB": "1.21695"
}
Retrieve a user's open orders
POST https://api.hyperliquid.xyz/info

See a user's open orders

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

"openOrders"

user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

dex

String

Perp dex name. Defaults to the empty string which represents the first perp dex. Spot open orders are only included with the first perp dex.

200: OK Successful R
Copy
[
    {
        "coin": "BTC",
        "limitPx": "29792.0",
        "oid": 91490942,
        "side": "A",
        "sz": "0.0",
        "timestamp": 1681247412573
    }
]
Retrieve a user's open orders with additional frontend info
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

"frontendOpenOrders"

user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

dex

String

Perp dex name. Defaults to the empty string which represents the first perp dex. Spot open orders are only included with the first perp dex.

200: OK
Copy
[
    {
        "coin": "BTC",
        "isPositionTpsl": false,
        "isTrigger": false,
        "limitPx": "29792.0",
        "oid": 91490942,
        "orderType": "Limit",
        "origSz": "5.0",
        "reduceOnly": false,
        "side": "A",
        "sz": "5.0",
        "timestamp": 1681247412573,
        "triggerCondition": "N/A",
        "triggerPx": "0.0",
    }
]
Retrieve a user's fills
POST https://api.hyperliquid.xyz/info

Returns at most 2000 most recent fills

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

"userFills"

user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

aggregateByTime

bool

When true, partial fills are combined when a crossing order gets filled by multiple different resting orders. Resting orders filled by multiple crossing orders will not be aggregated.

200: OK
Copy
[
    // Perp fill
    {
        "closedPnl": "0.0",
        "coin": "AVAX",
        "crossed": false,
        "dir": "Open Long",
        "hash": "0xa166e3fa63c25663024b03f2e0da011a00307e4017465df020210d3d432e7cb8",
        "oid": 90542681,
        "px": "18.435",
        "side": "B",
        "startPosition": "26.86",
        "sz": "93.53",
        "time": 1681222254710,
        "fee": "0.01",
        "feeToken": "USDC",
        "builderFee": "0.01", // this is optional and will not be present if 0
        "tid": 118906512037719
    },
    // Spot fill - note the difference in the "coin" format. Refer to 
    // https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/asset-ids
    // for more information on how spot asset IDs work
    {
        "coin": "@107",
        "px": "18.62041381",
        "sz": "43.84",
        "side": "A",
        "time": 1735969713869,
        "startPosition": "10659.65434798",
        "dir": "Sell",
        "closedPnl": "8722.988077",
        "hash": "0x2222138cc516e3fe746c0411dd733f02e60086f43205af2ae37c93f6a792430b",
        "oid": 59071663721,
        "crossed": true,
        "fee": "0.304521",
        "tid": 907359904431134,
        "feeToken": "USDC"
    }
]
Retrieve a user's fills by time
POST https://api.hyperliquid.xyz/info

Returns at most 2000 fills per response and only the 10000 most recent fills are available

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

userFillsByTime

user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

startTime*

int

Start time in milliseconds, inclusive

endTime

int

End time in milliseconds, inclusive. Defaults to current time.

aggregateByTime

bool

When true, partial fills are combined when a crossing order gets filled by multiple different resting orders. Resting orders filled by multiple crossing orders will not be aggregated.

200: OK Number of fills is limited to 2000
Copy
[
    // Perp fill
    {
        "closedPnl": "0.0",
        "coin": "AVAX",
        "crossed": false,
        "dir": "Open Long",
        "hash": "0xa166e3fa63c25663024b03f2e0da011a00307e4017465df020210d3d432e7cb8",
        "oid": 90542681,
        "px": "18.435",
        "side": "B",
        "startPosition": "26.86",
        "sz": "93.53",
        "time": 1681222254710,
        "fee": "0.01",
        "feeToken": "USDC",
        "builderFee": "0.01", // this is optional and will not be present if 0
        "tid": 118906512037719
    },
    // Spot fill - note the difference in the "coin" format. Refer to 
    // https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/asset-ids
    // for more information on how spot asset IDs work
    {
        "coin": "@107",
        "px": "18.62041381",
        "sz": "43.84",
        "side": "A",
        "time": 1735969713869,
        "startPosition": "10659.65434798",
        "dir": "Sell",
        "closedPnl": "8722.988077",
        "hash": "0x2222138cc516e3fe746c0411dd733f02e60086f43205af2ae37c93f6a792430b",
        "oid": 59071663721,
        "crossed": true,
        "fee": "0.304521",
        "tid": 907359904431134,
        "feeToken": "USDC"
    }
]
Query user rate limits
POST https://api.hyperliquid.xyz/info

Request Body
Name
Type
Description
user

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000

type

String

userRateLimit

200: OK A successful response
Copy
{
  "cumVlm": "2854574.593578",
  "nRequestsUsed": 2890,
  "nRequestsCap": 2864574
}
Query order status by oid or cloid
POST https://api.hyperliquid.xyz/info

Request Body
Name
Type
Description
user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

type*

String

"orderStatus"

oid*

uint64 or string

Either u64 representing the order id or 16-byte hex string representing the client order id

The <status> string returned has the following possible values:

Order status
Explanation
open

Placed successfully

filled

Filled

canceled

Canceled by user

triggered

Trigger order triggered

rejected

Rejected at time of placement

marginCanceled

Canceled because insufficient margin to fill

vaultWithdrawalCanceled

Vaults only. Canceled due to a user's withdrawal from vault 

openInterestCapCanceled

Canceled due to order being too aggressive when open interest was at cap

selfTradeCanceled

Canceled due to self-trade prevention

reduceOnlyCanceled

Canceled reduced-only order that does not reduce position

siblingFilledCanceled

TP/SL only. Canceled due to sibling ordering being filled

delistedCanceled

Canceled due to asset delisting

liquidatedCanceled

Canceled due to liquidation

scheduledCancel

API only. Canceled due to exceeding scheduled cancel deadline (dead man's switch)

tickRejected

Rejected due to invalid tick price

minTradeNtlRejected

Rejected due to order notional below minimum

perpMarginRejected

Rejected due to insufficient margin

reduceOnlyRejected

Rejected due to reduce only

badAloPxRejected

Rejected due to post-only immediate match

iocCancelRejected

Rejected due to IOC not able to match

badTriggerPxRejected

Rejected due to invalid TP/SL price

marketOrderNoLiquidityRejected

Rejected due to lack of liquidity for market order

positionIncreaseAtOpenInterestCapRejected

Rejected due to open interest cap

positionFlipAtOpenInterestCapRejected

Rejected due to open interest cap

tooAggressiveAtOpenInterestCapRejected

Rejected due to price too aggressive at open interest cap

openInterestIncreaseRejected

Rejected due to open interest cap

insufficientSpotBalanceRejected

Rejected due to insufficient spot balance

oracleRejected

Rejected due to price too far from oracle

perpMaxPositionRejected

Rejected due to exceeding margin tier limit at current leverage

200: OK A successful response
200: OK Missing Order
Copy
{
  "status": "order",
  "order": {
    "order": {
      "coin": "ETH",
      "side": "A",
      "limitPx": "2412.7",
      "sz": "0.0",
      "oid": 1,
      "timestamp": 1724361546645,
      "triggerCondition": "N/A",
      "isTrigger": false,
      "triggerPx": "0.0",
      "children": [],
      "isPositionTpsl": false,
      "reduceOnly": true,
      "orderType": "Market",
      "origSz": "0.0076",
      "tif": "FrontendMarket",
      "cloid": null
    },
    "status": <status>,
    "statusTimestamp": 1724361546645
  }
}
L2 book snapshot
POST https://api.hyperliquid.xyz/info

Returns at most 20 levels per side

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

"l2Book"

coin*

String

coin

nSigFigs

Number

Optional field to aggregate levels to nSigFigs significant figures. Valid values are 2, 3, 4, 5, and null, which means full precision

mantissa

Number

Optional field to aggregate levels. This field is only allowed if nSigFigs is 5. Accepts values of 1, 2 or 5.

Response

200: OK
Copy
{
  "coin": "BTC",
  "time": 1754450974231,
  "levels": [
    [
      {
        "px": "113377.0",
        "sz": "7.6699",
        "n": 17 // number of levels
      },
      {
        "px": "113376.0",
        "sz": "4.13714",
        "n": 8
      },
    ],
    [
      {
        "px": "113397.0",
        "sz": "0.11543",
        "n": 3
      }
    ]
  ]
}
Candle snapshot
POST https://api.hyperliquid.xyz/info

Only the most recent 5000 candles are available

Supported intervals: "1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "8h", "12h", "1d", "3d", "1w", "1M"

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

"candleSnapshot"

req*

Object

{"coin": <coin>, "interval": "15m", "startTime": <epoch millis>, "endTime": <epoch millis>}

Response

200: OK
Copy
[
  {
    "T": 1681924499999,
    "c": "29258.0",
    "h": "29309.0",
    "i": "15m",
    "l": "29250.0",
    "n": 189,
    "o": "29295.0",
    "s": "BTC",
    "t": 1681923600000,
    "v": "0.98639"
  }
]
Check builder fee approval
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

"maxBuilderFee"

user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

builder*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

Response

200: OK
Copy
1 // maximum fee approved in tenths of a basis point i.e. 1 means 0.001%
Retrieve a user's historical orders
POST https://api.hyperliquid.xyz/info

Returns at most 2000 most recent historical orders

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

"historicalOrders"

user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK
Copy
[
  {
    "order": {
      "coin": "ETH",
      "side": "A",
      "limitPx": "2412.7",
      "sz": "0.0",
      "oid": 1,
      "timestamp": 1724361546645,
      "triggerCondition": "N/A",
      "isTrigger": false,
      "triggerPx": "0.0",
      "children": [],
      "isPositionTpsl": false,
      "reduceOnly": true,
      "orderType": "Market",
      "origSz": "0.0076",
      "tif": "FrontendMarket",
      "cloid": null
    },
    "status": "filled" | "open" | "canceled" | "triggered" | "rejected" | "marginCanceled",
    "statusTimestamp": 1724361546645
  }
]
Retrieve a user's TWAP slice fills
POST https://api.hyperliquid.xyz/info

Returns at most 2000 most recent TWAP slice fills

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

"userTwapSliceFills"

user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK
Copy
[
    {
        "fill": {
            "closedPnl": "0.0",
            "coin": "AVAX",
            "crossed": true,
            "dir": "Open Long",
            "hash": "0x0000000000000000000000000000000000000000000000000000000000000000", // TWAP fills have a hash of 0
            "oid": 90542681,
            "px": "18.435",
            "side": "B",
            "startPosition": "26.86",
            "sz": "93.53",
            "time": 1681222254710,
            "fee": "0.01",
            "feeToken": "USDC",
            "tid": 118906512037719
        },
        "twapId": 3156
    }
]
Retrieve a user's subaccounts
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

"subAccounts"

user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK
Copy
[
  {
    "name": "Test",
    "subAccountUser": "0x035605fc2f24d65300227189025e90a0d947f16c",
    "master": "0x8c967e73e6b15087c42a10d344cff4c96d877f1d",
    "clearinghouseState": {
      "marginSummary": {
        "accountValue": "29.78001",
        "totalNtlPos": "0.0",
        "totalRawUsd": "29.78001",
        "totalMarginUsed": "0.0"
      },
      "crossMarginSummary": {
        "accountValue": "29.78001",
        "totalNtlPos": "0.0",
        "totalRawUsd": "29.78001",
        "totalMarginUsed": "0.0"
      },
      "crossMaintenanceMarginUsed": "0.0",
      "withdrawable": "29.78001",
      "assetPositions": [],
      "time": 1733968369395
    },
    "spotState": {
      "balances": [
        {
          "coin": "USDC",
          "token": 0,
          "total": "0.22",
          "hold": "0.0",
          "entryNtl": "0.0"
        }
      ]
    }
  }
]
Retrieve details for a vault
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

"vaultDetails"

vaultAddress*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

user

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK
Copy
{
  "name": "Test",
  "vaultAddress": "0xdfc24b077bc1425ad1dea75bcb6f8158e10df303",
  "leader": "0x677d831aef5328190852e24f13c46cac05f984e7",
  "description": "This community-owned vault provides liquidity to Hyperliquid through multiple market making strategies, performs liquidations, and accrues platform fees.",
  "portfolio": [
    [
      "day",
      {
        "accountValueHistory": [
          [
            1734397526634,
            "329265410.90790099"
          ]
        ],
        "pnlHistory": [
          [
            1734397526634,
            "0.0"
          ],
        ],
        "vlm": "0.0"
      }
    ],
    [
      "week" | "month" | "allTime" | "perpDay" | "perpWeek" | "perpMonth" | "perpAllTime",
      {...}
    ]
  ],
  "apr": 0.36387129259090006,
  "followerState": null,
  "leaderFraction": 0.0007904828725729887,
  "leaderCommission": 0,
  "followers": [
    {
      "user": "0x005844b2ffb2e122cf4244be7dbcb4f84924907c",
      "vaultEquity": "714491.71026243",
      "pnl": "3203.43026143",
      "allTimePnl": "79843.74476743",
      "daysFollowing": 388,
      "vaultEntryTime": 1700926145201,
      "lockupUntil": 1734824439201
    }
  ],
  "maxDistributable": 94856870.164485,
  "maxWithdrawable": 742557.680863,
  "isClosed": false,
  "relationship": {
    "type": "parent",
    "data": {
      "childAddresses": [
        "0x010461c14e146ac35fe42271bdc1134ee31c703a",
        "0x2e3d94f0562703b25c83308a05046ddaf9a8dd14",
        "0x31ca8395cf837de08b24da3f660e77761dfb974b"
      ]
    }
  },
  "allowDeposits": true,
  "alwaysCloseOnWithdraw": false
}  
Retrieve a user's vault deposits
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

"userVaultEquities"

user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK
Copy
[
  {
    "vaultAddress": "0xdfc24b077bc1425ad1dea75bcb6f8158e10df303",
    "equity": "742500.082809",
  }
]
Query a user's role
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

"userRole"

user*

String

Address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

User
Agent
Vault
Subaccount
Missing
Copy
{"role":"user"} # "missing", "user", "agent", "vault", or "subAccount"
Query a user's portfolio
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

"portfolio"

user*

String

hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK
Copy
[
  [
    "day",
    {
      "accountValueHistory": [
        [
          1741886630493,
          "0.0"
        ],
        [
          1741895270493,
          "0.0"
        ],
        ...
      ],
      "pnlHistory": [
        [
          1741886630493,
          "0.0"
        ],
        [
          1741895270493,
          "0.0"
        ],
        ...
      ],
      "vlm": "0.0"
    }
  ],
  ["week", { ... }],
  ["month", { ... }],
  ["allTime", { ... }],
  ["perpDay", { ... }],
  ["perpWeek", { ... }],
  ["perpMonth", { ... }],
  ["perpAllTime", { ... }]
]
Query a user's referral information
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

"referral"

user*

String

hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK
Copy
{
    "referredBy": {
        "referrer": "0x5ac99df645f3414876c816caa18b2d234024b487",
        "code": "TESTNET"
    },
    "cumVlm": "149428030.6628420055",
    "unclaimedRewards": "11.047361",
    "claimedRewards": "22.743781",
    "builderRewards": "0.027802",
    "referrerState": {
        "stage": "ready",
        "data": {
            "code": "TEST",
            "referralStates": [
                {
                    "cumVlm": "960652.017122",
                    "cumRewardedFeesSinceReferred": "196.838825",
                    "cumFeesRewardedToReferrer": "19.683748",
                    "timeJoined": 1679425029416,
                    "user": "0x11af2b93dcb3568b7bf2b6bd6182d260a9495728"
                },
                {
                    "cumVlm": "438278.672653",
                    "cumRewardedFeesSinceReferred": "97.628107",
                    "cumFeesRewardedToReferrer": "9.762562",
                    "timeJoined": 1679423947882,
                    "user": "0x3f69d170055913103a034a418953b8695e4e42fa"
                }
            ]
        }
    },
    "rewardHistory": []
}
Note that rewardHistory is for legacy rewards. Claimed rewards are now returned in nonFundingLedgerUpdate

Query a user's fees
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

"userFees"

user*

String

hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK
Copy
{
  "dailyUserVlm": [
    {
      "date": "2025-05-23",
      "userCross": "0.0",
      "userAdd": "0.0",
      "exchange": "2852367.0770729999"
    },
    ...
  ],
  "feeSchedule": {
    "cross": "0.00045",
    "add": "0.00015",
    "spotCross": "0.0007",
    "spotAdd": "0.0004",
    "tiers": {
      "vip": [
        {
          "ntlCutoff": "5000000.0",
          "cross": "0.0004",
          "add": "0.00012",
          "spotCross": "0.0006",
          "spotAdd": "0.0003"
        },
        ...
      ],
      "mm": [
        {
          "makerFractionCutoff": "0.005",
          "add": "-0.00001"
        },
        ...
      ]
    },
    "referralDiscount": "0.04",
    "stakingDiscountTiers": [
      {
        "bpsOfMaxSupply": "0.0",
        "discount": "0.0"
      },
      {
        "bpsOfMaxSupply": "0.0001",
        "discount": "0.05"
      },
      ...
    ]
  },
  "userCrossRate": "0.000315",
  "userAddRate": "0.000105",
  "userSpotCrossRate": "0.00049",
  "userSpotAddRate": "0.00028",
  "activeReferralDiscount": "0.0",
  "trial": null,
  "feeTrialReward": "0.0",
  "nextTrialAvailableTimestamp": null,
  "stakingLink": {
    "type": "tradingUser",
    "stakingUser": "0x54c049d9c7d3c92c2462bf3d28e083f3d6805061"
  },
  "activeStakingDiscount": {
    "bpsOfMaxSupply": "4.7577998927",
    "discount": "0.3"
  }
}
Query a user's staking delegations
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

"delegations"

user*

String

hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK
Copy
[
    {
        "validator":"0x5ac99df645f3414876c816caa18b2d234024b487",
        "amount":"12060.16529862",
        "lockedUntilTimestamp":1735466781353
    },
    ...
]
Query a user's staking summary
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

"delegatorSummary"

user*

String

hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK
Copy
{
    "delegated": "12060.16529862",
    "undelegated": "0.0",
    "totalPendingWithdrawal": "0.0",
    "nPendingWithdrawals": 0
}
Query a user's staking history
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

"delegatorHistory"

user*

String

hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK
Copy
[
    {
        "time": 1735380381353,
        "hash": "0x55492465cb523f90815a041a226ba90147008d4b221a24ae8dc35a0dbede4ea4",
        "delta": {
            "delegate": {
                "validator": "0x5ac99df645f3414876c816caa18b2d234024b487",
                "amount": "10000.0",
                "isUndelegate": false
            }
        }
    },
    ...
]
Query a user's staking rewards
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

"delegatorRewards"

user*

String

hexadecimal format; e.g. 0x0000000000000000000000000000000000000000.

200: OK
Copy
[
    {
        "time": 1736726400073,
        "source": "delegation",
        "totalAmount": "0.73117184"
    },
    {
        "time": 1736726400073,
        "source": "commission",
        "totalAmount": "130.76445876"
    },
    ...
]