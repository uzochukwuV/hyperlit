For developers
API
Exchange endpoint
The exchange endpoint is used to interact with and trade on the Hyperliquid chain. See the Python SDK for code to generate signatures for these requests.

Asset
Many of the requests take asset as an input. For perpetuals this is the index in the universe field returned by themeta response. For spot assets, use 10000 + index where index is the corresponding index in spotMeta.universe. For example, when submitting an order for PURR/USDC, the asset that should be used is 10000 because its asset index in the spot metadata is 0.

Subaccounts and vaults
Subaccounts and vaults do not have private keys. To perform actions on behalf of a subaccount or vault signing should be done by the master account and the vaultAddress field should be set to the address of the subaccount or vault. The basic_vault.py example in the Python SDK demonstrates this.

Expires After
Some actions support an optional field expiresAfter which is a timestamp in milliseconds after which the action will be rejected. User-signed actions such as Core USDC transfer do not support the expiresAfter field. Note that actions consume 5x the usual address-based rate limit when canceled due to a stale expiresAfter field. 

See the Python SDK for details on how to incorporate this field when signing. 

Place an order
POST https://api.hyperliquid.xyz/exchange

See Python SDK for full featured examples on the fields of the order request.

For limit orders, TIF (time-in-force) sets the behavior of the order upon first hitting the book.

ALO (add liquidity only, i.e. "post only") will be canceled instead of immediately matching.

IOC (immediate or cancel) will have the unfilled part canceled instead of resting.

GTC (good til canceled) orders have no special behavior.

Client Order ID (cloid) is an optional 128 bit hex string, e.g. 0x1234567890abcdef1234567890abcdef

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
action*

Object

{

  "type": "order",
  "orders": [{

    "a": Number,

    "b": Boolean,

    "p": String,

    "s": String,

    "r": Boolean,

    "t": {

      "limit": {

        "tif": "Alo" | "Ioc" | "Gtc" 

      } or

      "trigger": {

         "isMarket": Boolean,

         "triggerPx": String,

         "tpsl": "tp" | "sl"

       }

    },

    "c": Cloid (optional)

  }],

  "grouping": "na" | "normalTpsl" | "positionTpsl",

  "builder": Optional({"b": "address", "f": Number})

}

Meaning of keys:
a is asset
b is isBuy
p is price
s is size
r is reduceOnly
t is type
c is cloid (client order id)

Meaning of keys in optional builder argument:
b is the address the should receive the additional fee
f is the size of the fee in tenths of a basis point e.g. if f is 10, 1bp of the order notional  will be charged to the user and sent to the builder

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

vaultAddress

String

If trading on behalf of a vault or subaccount, its Onchain address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000

expiresAfter

Number

Timestamp in milliseconds

200: OK Successful Response (resting)
200: OK Error Response
200: OK Successful Response (filled)
Copy
{
   "status":"ok",
   "response":{
      "type":"order",
      "data":{
         "statuses":[
            {
               "resting":{
                  "oid":77738308
               }
            }
         ]
      }
   }
}
Cancel order(s)
POST https://api.hyperliquid.xyz/exchange

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
action*

Object

{

  "type": "cancel",

  "cancels": [

    {

      "a": Number,

      "o": Number

    }

  ]

}

Meaning of keys:
a is asset
o is oid (order id)

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

vaultAddress

String

If trading on behalf of a vault or subaccount, its address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000

expiresAfter

Number

Timestamp in milliseconds

200: OK Successful Response
200: OK Error Response
Copy
{
   "status":"ok",
   "response":{
      "type":"cancel",
      "data":{
         "statuses":[
            "success"
         ]
      }
   }
}
Cancel order(s) by cloid
POST https://api.hyperliquid.xyz/exchange 

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
action*

Object

{

  "type": "cancelByCloid",

  "cancels": [

    {

      "asset": Number,

      "cloid": String

    }

  ]

}

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

vaultAddress

String

If trading on behalf of a vault or subaccount, its address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000

expiresAfter

Number

Timestamp in milliseconds

200: OK Successful Response
200: OK Error Response
Schedule cancel (dead man's switch)
POST https://api.hyperliquid.xyz/exchange 

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
action*

Object

{

  "type": "scheduleCancel",

  "time": number (optional)

}

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

vaultAddress

String

If trading on behalf of a vault or subaccount, its address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000

expiresAfter

Number

Timestamp in milliseconds

Schedule a cancel-all operation at a future time. Not including time will remove the scheduled cancel operation. The time must be at least 5 seconds after the current time. Once the time comes, all open orders will be canceled and a trigger count will be incremented. The max number of triggers per day is 10. This trigger count is reset at 00:00 UTC.

Modify an order
POST https://api.hyperliquid.xyz/exchange  

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
action*

Object

{

  "type": "modify",

  "oid": Number | Cloid,

  "order": {

    "a": Number,

    "b": Boolean,

    "p": String,

    "s": String,

    "r": Boolean,

    "t": {

      "limit": {

        "tif": "Alo" | "Ioc" | "Gtc" 

      } or

      "trigger": {

         "isMarket": Boolean,

         "triggerPx": String,

         "tpsl": "tp" | "sl"

       }

    },

    "c": Cloid (optional)

  }

}

Meaning of keys:
a is asset
b is isBuy
p is price
s is size
r is reduceOnly
t is type
c is cloid (client order id)

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

vaultAddress

String

If trading on behalf of a vault or subaccount, its Onchain address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000

expiresAfter

Number

Timestamp in milliseconds

200: OK Successful Response
200: OK Error Response
Modify multiple orders
POST https://api.hyperliquid.xyz/exchange

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
action*

Object

{

  "type": "batchModify",

  "modifies": [{

    "oid": Number | Cloid,

    "order": {

      "a": Number,

      "b": Boolean,

      "p": String,

      "s": String,

      "r": Boolean,

      "t": {

        "limit": {

          "tif": "Alo" | "Ioc" | "Gtc" 

        } or

        "trigger": {

           "isMarket": Boolean,

           "triggerPx": String,

           "tpsl": "tp" | "sl"

         }

      },

      "c": Cloid (optional)

    }

  }]

}

Meaning of keys:
a is asset
b is isBuy
p is price
s is size
r is reduceOnly
t is type
c is cloid (client order id)

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

vaultAddress

String

If trading on behalf of a vault or subaccount, its Onchain address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000

expiresAfter

Number

Timestamp in milliseconds

Update leverage
POST https://api.hyperliquid.xyz/exchange

Update cross or isolated leverage on a coin. 

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
action*

Object

{

  "type": "updateLeverage",

  "asset": index of coin,

  "isCross": true or false if updating cross-leverage,

  "leverage": integer representing new leverage, subject to leverage constraints on that coin

}

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

vaultAddress

String

If trading on behalf of a vault or subaccount, its Onchain address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000

expiresAfter

Number

Timestamp in milliseconds

200: OK Successful response
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Update isolated margin
POST https://api.hyperliquid.xyz/exchange

Add or remove margin from isolated position

Note that to target a specific leverage instead of a USDC value of margin change, there is an alternate action {"type": "topUpIsolatedOnlyMargin", "asset": <asset>, "leverage": <float string>}

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
action*

Object

{

  "type": "updateIsolatedMargin",

  "asset": index of coin,

  "isBuy": true, (this parameter won't have any effect until hedge mode is introduced)

  "ntli": int representing amount to add or remove with 6 decimals, e.g. 1000000 for 1 usd,

}

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

vaultAddress

String

If trading on behalf of a vault or subaccount, its Onchain address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000

expiresAfter

Number

Timestamp in milliseconds

200: OK Successful response
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Core USDC transfer
POST https://api.hyperliquid.xyz/exchange

Send usd to another address. This transfer does not touch the EVM bridge. The signature format is human readable for wallet interfaces.

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
action*

Object

{

  "type": "usdSend",

  "hyperliquidChain": "Mainnet" (on testnet use "Testnet" instead),
  "signatureChainId": the id of the chain used when signing in hexadecimal format; e.g. "0xa4b1" for Arbitrum,

  "destination": address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000,

   "amount": amount of usd to send as a string, e.g. "1" for 1 usd,

     "time": current timestamp in milliseconds as a Number, should match nonce

}

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

200: OK Successful Response
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Core spot transfer
POST https://api.hyperliquid.xyz/exchange

Send spot assets to another address. This transfer does not touch the EVM bridge. The signature format is human readable for wallet interfaces.

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
action*

Object

{

  "type": "spotSend",

  "hyperliquidChain": "Mainnet" (on testnet use "Testnet" instead),
  "signatureChainId": the id of the chain used when signing in hexadecimal format; e.g. "0xa4b1" for Arbitrum,

  "destination": address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000,
  "token": tokenName:tokenId; e.g. "PURR:0xc4bf3f870c0e9465323c0b6ed28096c2",

   "amount": amount of token to send as a string, e.g. "0.01",

     "time": current timestamp in milliseconds as a Number, should match nonce

}

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

200: OK Successful Response
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Copy
Example sign typed data for generating the signature:
{
  "types": {
    "HyperliquidTransaction:SpotSend": [
      {
        "name": "hyperliquidChain",
        "type": "string"
      },
      {
        "name": "destination",
        "type": "string"
      },
      {
        "name": "token",
        "type": "string"
      },
      {
        "name": "amount",
        "type": "string"
      },
      {
        "name": "time",
        "type": "uint64"
      }
    ]
  },
  "primaryType": "HyperliquidTransaction:SpotSend",
  "domain": {
    "name": "HyperliquidSignTransaction",
    "version": "1",
    "chainId": 42161,
    "verifyingContract": "0x0000000000000000000000000000000000000000"
  },
  "message": {
    "destination": "0x0000000000000000000000000000000000000000",
    "token": "PURR:0xc1fb593aeffbeb02f85e0308e9956a90",
    "amount": "0.1",
    "time": 1716531066415,
    "hyperliquidChain": "Mainnet"
  }
}
Initiate a withdrawal request
POST https://api.hyperliquid.xyz/exchange

This method is used to initiate the withdrawal flow. After making this request, the L1 validators will sign and send the withdrawal request to the bridge contract. There is a $1 fee for withdrawing at the time of this writing and withdrawals take approximately 5 minutes to finalize.

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
action*

Object

{
  "type": "withdraw3",

  "hyperliquidChain": "Mainnet" (on testnet use "Testnet" instead),
  "signatureChainId": the id of the chain used when signing in hexadecimal format; e.g. "0xa4b1" for Arbitrum,

  "amount": amount of usd to send as a string, e.g. "1" for 1 usd,

  "time": current timestamp in milliseconds as a Number, should match nonce,

  "destination": address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000

}

nonce*

Number

Recommended to use the current timestamp in milliseconds, must match the nonce in the action Object above

signature*

Object

200: OK
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Transfer from Spot account to Perp account (and vice versa)
POST https://api.hyperliquid.xyz/exchange

This method is used to transfer USDC from the user's spot wallet to perp wallet and vice versa.

Headers

Name
Value
Content-Type*

"application/json"

Body

Name
Type
Description
action*

Object

{

  "type": "usdClassTransfer",

  "hyperliquidChain": "Mainnet" (on testnet use "Testnet" instead),
  "signatureChainId": the id of the chain used when signing in hexadecimal format; e.g. "0xa4b1" for Arbitrum,

 "amount": amount of usd to transfer as a string, e.g. "1" for 1 usd. If you want to use this action for a subaccount, you can include subaccount: address after the amount, e.g. "1" subaccount:0x0000000000000000000000000000000000000000,

  "toPerp": true if (spot -> perp) else false,

"nonce": current timestamp in milliseconds as a Number, must match nonce in outer request body

}

nonce*

Number

Recommended to use the current timestamp in milliseconds, must match the nonce in the action Object above

signature*

Object

Response

200: OK
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Send Asset (testnet only)
POST https://api.hyperliquid.xyz/exchange

This generalized method is used to transfer tokens between different perp DEXs, spot balance, users, and/or sub-accounts. Use "" to specify the default USDC perp DEX and "spot" to specify spot. Only the collateral token can be transferred to or from a perp DEX.

Headers
Name
Value
Content-Type*

application/json

Body
Name
Type
Description
action*

Object

{

  "type": "sendAsset",

  "hyperliquidChain": "Mainnet" (on testnet use "Testnet" instead),

  "signatureChainId": the id of the chain used when signing in hexadecimal format; e.g. "0xa4b1" for Arbitrum,

  "destination": address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000,

  "sourceDex": name of perp dex to transfer from,

  "destinationDex": name of the perp dex to transfer to,

  "token": tokenName:tokenId; e.g. "PURR:0xc4bf3f870c0e9465323c0b6ed28096c2",

  "amount": amount of token to send as a string; e.g. "0.01",

  "fromSubAccount": address in 42-character hexadecimal format or empty string if not from a subaccount,

  "nonce": current timestamp in milliseconds as a Number, should match nonce

}

nonce*

Number

Recommended to use the current timestamp in milliseconds, must match the nonce in the action Object above

signature*

Object

Response
200: OK
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Deposit into staking
POST https://api.hyperliquid.xyz/exchange

This method is used to transfer native token from the user's spot account into staking for delegating to validators. 

Headers
Name
Value
Content-Type*

application/json

Body
Name
Type
Description
action*

Object

{

  "type": "cDeposit",

  "hyperliquidChain": "Mainnet" (on testnet use "Testnet" instead),
  "signatureChainId": the id of the chain used when signing in hexadecimal format; e.g. "0xa4b1" for Arbitrum,

 "wei": amount of wei to transfer as a number,

"nonce": current timestamp in milliseconds as a Number, must match nonce in outer request body

}

nonce*

Number

Recommended to use the current timestamp in milliseconds, must match the nonce in the action Object above

signature*

Object

Response
200: OK
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Withdraw from staking
POST https://api.hyperliquid.xyz/exchange

This method is used to transfer native token from staking into the user's spot account. Note that transfers from staking to spot account go through a 7 day unstaking queue.

Headers
Name
Value
Content-Type*

application/json

Body
Name
Type
Description
action*

Object

{

  "type": "cWithdraw",

  "hyperliquidChain": "Mainnet" (on testnet use "Testnet" instead),
  "signatureChainId": the id of the chain used when signing in hexadecimal format; e.g. "0xa4b1" for Arbitrum,

 "wei": amount of wei to transfer as a number,

"nonce": current timestamp in milliseconds as a Number, must match nonce in outer request body

}

nonce*

Number

Recommended to use the current timestamp in milliseconds, must match the nonce in the action Object above

signature*

Object

Response
200: OK
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Delegate or undelegate stake from validator
POST https://api.hyperliquid.xyz/exchange

Delegate or undelegate native tokens to or from a validator. Note that delegations to a particular validator have a lockup duration of 1 day.

Headers
Name
Value
Content-Type*

application/json

Body
Name
Type
Description
action*

Object

{

  "type": "tokenDelegate",

  "hyperliquidChain": "Mainnet" (on testnet use "Testnet" instead),
  "signatureChainId": the id of the chain used when signing in hexadecimal format; e.g. "0xa4b1" for Arbitrum,

  "validator": address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000,
"isUndelegate": boolean,

"wei": number,

"nonce": current timestamp in milliseconds as a Number, must match nonce in outer request body

}

nonce*

number

Recommended to use the current timestamp in milliseconds

signature*

Object

Response
200: OK
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Deposit or withdraw from a vault
POST https://api.hyperliquid.xyz/exchange

Add or remove funds from a vault.

Headers

Name
Value
Content-Type*

application/json

Body

Name
Type
Description
action*

Object

{

  "type": "vaultTransfer",

  "vaultAddress": address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000,
"isDeposit": boolean,

"usd": number

}

nonce*

number

Recommended to use the current timestamp in milliseconds

signature*

Object

expiresAfter

Number

Timestamp in milliseconds

Response

200
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Approve an API wallet
POST https://api.hyperliquid.xyz/exchange

Approves an API Wallet (also sometimes referred to as an Agent Wallet). See here for more details.

Headers

Name
Value
Content-Type*

application/json

Body

Name
Type
Description
action*

Object

{
  "type": "approveAgent",

  "hyperliquidChain": "Mainnet" (on testnet use "Testnet" instead),
  "signatureChainId": the id of the chain used when signing in hexadecimal format; e.g. "0xa4b1" for Arbitrum,

  "agentAddress": address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000,

"agentName": Optional name for the API wallet. An account can have 1 unnamed approved wallet and up to 3 named ones. And additional 2 named agents are allowed per subaccount,

  "nonce": current timestamp in milliseconds as a Number, must match nonce in outer request body

}

nonce*

number

Recommended to use the current timestamp in milliseconds

signature*

Object

Response

200
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Approve a builder fee
POST https://api.hyperliquid.xyz/exchange

Approve a maximum fee rate for a builder.

Headers

Name
Value
Content-Type*

application/json

Body

Name
Type
Description
action*

Object

{
  "type": "approveBuilderFee",

  "hyperliquidChain": "Mainnet" (on testnet use "Testnet" instead),
  "signatureChainId": the id of the chain used when signing in hexadecimal format; e.g. "0xa4b1" for Arbitrum,

  "maxFeeRate": the maximum allowed builder fee rate as a percent string; e.g. "0.001%",

  "builder": address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000,

  "nonce": current timestamp in milliseconds as a Number, must match nonce in outer request body

}

nonce*

number

Recommended to use the current timestamp in milliseconds

signature*

Object

Response

200
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Place a TWAP order
POST https://api.hyperliquid.xyz/exchange

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
action*

Object

{

  "type": "twapOrder",
  "twap": {

    "a": Number,

    "b": Boolean,

    "s": String,

    "r": Boolean,

    "m": Number,

    "t": Boolean

  }

  }

Meaning of keys:
a is asset
b is isBuy
s is size
r is reduceOnly

m is minutes
t is randomize

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

vaultAddress

String

If trading on behalf of a vault or subaccount, its Onchain address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000

expiresAfter

Number

Timestamp in milliseconds

200: OK Successful Response
200: OK Error Response
Copy
{
   "status":"ok",
   "response":{
      "type":"twapOrder",
      "data":{
         "status": {
            "running":{
               "twapId":77738308
            }
         }
      }
   }
}
Cancel a TWAP order
POST https://api.hyperliquid.xyz/exchange

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
action*

Object

{

  "type": "twapCancel",

   "a": Number,

   "t": Number

}

Meaning of keys:
a is asset
t is twap_id

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

vaultAddress

String

If trading on behalf of a vault or subaccount, its address in 42-character hexadecimal format; e.g. 0x0000000000000000000000000000000000000000

expiresAfter

Number

Timestamp in milliseconds

200: OK Successful Response
200: OK Error Response
Copy
{
   "status":"ok",
   "response":{
      "type":"twapCancel",
      "data":{
         "status": "success"
      }
   }
}
Reserve Additional Actions
POST https://api.hyperliquid.xyz/exchange 

Instead of trading to increase the address based rate limits, this action allows reserving additional actions for 0.0005 USDC per request. The cost is paid from the Perps balance. 

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
action*

Object

{

  "type": "reserveRequestWeight",

   "weight": Number

}

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

expiresAfter

Number

Timestamp in milliseconds

200: OK Successful Response
Copy
{'status': 'ok', 'response': {'type': 'default'}}
Invalidate Pending Nonce (noop)
POST https://api.hyperliquid.xyz/exchange 

This action does not do anything (no operation), but causes the nonce to be marked as used. This can be a more effective way to cancel in-flight orders than the cancel action.

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
action*

Object

{

  "type": "noop"

}

nonce*

Number

Recommended to use the current timestamp in milliseconds

signature*

Object

expiresAfter

Number
    
Timestamp in milliseconds

200: OK Successful Response
Copy
{'status': 'ok', 'response': {'type': 'default'}}