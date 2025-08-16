For developers
API
Websocket
WebSocket endpoints are available for real-time data streaming and as an alternative to HTTP request sending on the Hyperliquid exchange. The WebSocket URLs by network are:

Mainnet: wss://api.hyperliquid.xyz/ws 

Testnet: wss://api.hyperliquid-testnet.xyz/ws.

Connecting
To connect to the WebSocket API, you must establish a WebSocket connection to the respective URL based on your desired network. Once connected, you can start sending subscription messages to receive real-time data updates.

Example from command line:

Copy
$ wscat -c  wss://api.hyperliquid.xyz/ws
Connected (press CTRL+C to quit)
>  { "method": "subscribe", "subscription": { "type": "trades", "coin": "SOL" } }
< {"channel":"subscriptionResponse","data":{"method":"subscribe","subscription":{"type":"trades","coin":"SOL"}}}
Note: this doc uses Typescript for defining many of the message types. If you prefer to use Python, you can check out the equivalent types in the python SDK here and example connection code here.

Subscriptions
This page describes subscribing to data streams using the WebSocket API.

Subscription messages
To subscribe to specific data feeds, you need to send a subscription message. The subscription message format is as follows:

Copy
{
  "method": "subscribe",
  "subscription": { ... }
}
The subscription ack provides a snapshot of previous data for time series data (e.g. user fills). These snapshot messages are tagged with isSnapshot: true and can be ignored if the previous messages were already processed.

The subscription object contains the details of the specific feed you want to subscribe to. Choose from the following subscription types and provide the corresponding properties:

allMids:

Subscription message: { "type": "allMids", "dex": "<dex>" }

Data format: AllMids 

The dex field represents the perp dex to source mids from.

Note that the dex field is optional. If not provided, then the first perp dex is used. Spot mids are only included with the first perp dex.

notification:

Subscription message: { "type": "notification", "user": "<address>" }

Data format: Notification

webData2

Subscription message: { "type": "webData2", "user": "<address>" }

Data format: WebData2

candle:

Subscription message: { "type": "candle", "coin": "<coin_symbol>", "interval": "<candle_interval>" }

 Supported intervals: "1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "8h", "12h", "1d", "3d", "1w", "1M"

Data format: Candle[]

l2Book:

Subscription message: { "type": "l2Book", "coin": "<coin_symbol>" }

Optional parameters: nSigFigs: int, mantissa: int

Data format: WsBook

trades:

Subscription message: { "type": "trades", "coin": "<coin_symbol>" }

Data format: WsTrade[]

orderUpdates:

Subscription message: { "type": "orderUpdates", "user": "<address>" }

Data format: WsOrder[]

userEvents: 

Subscription message: { "type": "userEvents", "user": "<address>" }

Data format: WsUserEvent

userFills: 

Subscription message: { "type": "userFills", "user": "<address>" }

Optional parameter:  aggregateByTime: bool 

Data format: WsUserFills

userFundings: 

Subscription message: { "type": "userFundings", "user": "<address>" }

Data format: WsUserFundings

userNonFundingLedgerUpdates: 

Subscription message: { "type": "userNonFundingLedgerUpdates", "user": "<address>" }

Data format: WsUserNonFundingLedgerUpdates

activeAssetCtx: 

Subscription message: { "type": "activeAssetCtx", "coin": "coin_symbol>" }

Data format: WsActiveAssetCtx or WsActiveSpotAssetCtx 

activeAssetData: (only supports Perps)

Subscription message: { "type": "activeAssetData", "user": "<address>", "coin": "coin_symbol>" }

Data format: WsActiveAssetData

userTwapSliceFills: 

Subscription message: { "type": "userTwapSliceFills", "user": "<address>" }

Data format: WsUserTwapSliceFills

userTwapHistory: 

Subscription message: { "type": "userTwapHistory", "user": "<address>" }

Data format: WsUserTwapHistory

bbo :

Subscription message: { "type": "bbo", "coin": "<coin>" }

Data format: WsBbo

Data formats
The server will respond to successful subscriptions with a message containing the channel property set to "subscriptionResponse", along with the data field providing the original subscription. The server will then start sending messages with the channel property set to the corresponding subscription type e.g. "allMids" and the data field providing the subscribed data.

The data field format depends on the subscription type:

AllMids: All mid prices.

Format: AllMids { mids: Record<string, string> }

Notification: A notification message.

Format: Notification { notification: string }

WebData2: Aggregate information about a user, used primarily for the frontend.

Format: WebData2

WsTrade[]: An array of trade updates.

Format: WsTrade[]

WsBook: Order book snapshot updates.

Format: WsBook { coin: string; levels: [Array<WsLevel>, Array<WsLevel>]; time: number; }

WsOrder: User order updates.

Format: WsOrder[]

WsUserEvent: User events that are not order updates

Format: WsUserEvent { "fills": [WsFill] | "funding": WsUserFunding | "liquidation": WsLiquidation | "nonUserCancel": [WsNonUserCancel] }

WsUserFills : Fills snapshot followed by streaming fills

WsUserFundings : Funding payments snapshot followed by funding payments on the hour

WsUserNonFundingLedgerUpdates: Ledger updates not including funding payments: withdrawals, deposits, transfers, and liquidations

WsBbo : Bbo updates that are sent only if the bbo changes on a block

For the streaming user endpoints such as WsUserFills,WsUserFundings the first message has isSnapshot: true and the following streaming updates have isSnapshot: false. 

Data type definitions
Here are the definitions of the data types used in the WebSocket API:

Copy
interface WsTrade {
  coin: string;
  side: string;
  px: string;
  sz: string;
  hash: string;
  time: number;
  // tid is 50-bit hash of (buyer_oid, seller_oid). 
  // For a globally unique trade id, use (block_time, coin, tid)
  tid: number;  
  users: [string, string] // [buyer, seller]
}

// Snapshot feed, pushed on each block that is at least 0.5 since last push
interface WsBook {
  coin: string;
  levels: [Array<WsLevel>, Array<WsLevel>];
  time: number;
}

interface WsBbo {
  coin: string;
  time: number;
  bbo: [WsLevel | null, WsLevel | null];
}

interface WsLevel {
  px: string; // price
  sz: string; // size
  n: number; // number of orders
}

interface Notification {
  notification: string;
}

interface AllMids {
  mids: Record<string, string>;
}

interface Candle {
  t: number; // open millis
  T: number; // close millis
  s: string; // coin
  i: string; // interval
  o: number; // open price
  c: number; // close price
  h: number; // high price
  l: number; // low price
  v: number; // volume (base unit)
  n: number; // number of trades
}

type WsUserEvent = {"fills": WsFill[]} | {"funding": WsUserFunding} | {"liquidation": WsLiquidation} | {"nonUserCancel" :WsNonUserCancel[]};

interface WsUserFills {
  isSnapshot?: boolean;
  user: string;
  fills: Array<WsFill>;
}

interface WsFill {
  coin: string;
  px: string; // price
  sz: string; // size
  side: string;
  time: number;
  startPosition: string;
  dir: string; // used for frontend display
  closedPnl: string;
  hash: string; // L1 transaction hash
  oid: number; // order id
  crossed: boolean; // whether order crossed the spread (was taker)
  fee: string; // negative means rebate
  tid: number; // unique trade id
  liquidation?: FillLiquidation;
  feeToken: string; // the token the fee was paid in
  builderFee?: string; // amount paid to builder, also included in fee
}

interface FillLiquidation {
  liquidatedUser?: string;
  markPx: number;
  method: "market" | "backstop";
}

interface WsUserFunding {
  time: number;
  coin: string;
  usdc: string;
  szi: string;
  fundingRate: string;
}

interface WsLiquidation {
  lid: number;
  liquidator: string;
  liquidated_user: string;
  liquidated_ntl_pos: string;
  liquidated_account_value: string;
}

interface WsNonUserCancel {
  coin: String;
  oid: number;
}

interface WsOrder {
  order: WsBasicOrder;
  status: string; // See https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/info-endpoint#query-order-status-by-oid-or-cloid for a list of possible values
  statusTimestamp: number;
}

interface WsBasicOrder {
  coin: string;
  side: string;
  limitPx: string;
  sz: string;
  oid: number;
  timestamp: number;
  origSz: string;
  cloid: string | undefined;
}

interface WsActiveAssetCtx {
  coin: string;
  ctx: PerpsAssetCtx;
}

interface WsActiveSpotAssetCtx {
  coin: string;
  ctx: SpotAssetCtx;
}

type SharedAssetCtx = {
  dayNtlVlm: number;
  prevDayPx: number;
  markPx: number;
  midPx?: number;
};

type PerpsAssetCtx = SharedAssetCtx & {
  funding: number;
  openInterest: number;
  oraclePx: number;
};

type SpotAssetCtx = SharedAssetCtx & {
  circulatingSupply: number;
};

interface WsActiveAssetData {
  user: string;
  coin: string;
  leverage: Leverage;
  maxTradeSzs: [number, number];
  availableToTrade: [number, number];
}

interface WsTwapSliceFill {
  fill: WsFill;
  twapId: number;
}

interface WsUserTwapSliceFills {
  isSnapshot?: boolean;
  user: string;
  twapSliceFills: Array<WsTwapSliceFill>;
}

interface TwapState {
  coin: string;
  user: string;
  side: string;
  sz: number;
  executedSz: number;
  executedNtl: number;
  minutes: number;
  reduceOnly: boolean;
  randomize: boolean;
  timestamp: number;
}

type TwapStatus = "activated" | "terminated" | "finished" | "error";
interface WsTwapHistory {
  state: TwapState;
  status: {
    status: TwapStatus;
    description: string;
  };
  time: number;
}

interface WsUserTwapHistory {
  isSnapshot?: boolean;
  user: string;
  history: Array<WsTwapHistory>;
}
Please note that the above data types are in TypeScript format, and their usage corresponds to the respective subscription types.

Examples
Here are a few examples of subscribing to different feeds using the subscription messages:

Subscribe to all mid prices:

Copy
{ "method": "subscribe", "subscription": { "type": "allMids" } }
Subscribe to notifications for a specific user:

Copy
{ "method": "subscribe", "subscription": { "type": "notification", "user": "<address>" } }
Subscribe to web data for a specific user:

Copy
{ "method": "subscribe", "subscription": { "type": "webData", "user": "<address>" } }
Subscribe to candle updates for a specific coin and interval:

Copy
{ "method": "subscribe", "subscription": { "type": "candle", "coin": "<coin_symbol>", "interval": "<candle_interval>" } }
Subscribe to order book updates for a specific coin:

Copy
{ "method": "subscribe", "subscription": { "type": "l2Book", "coin": "<coin_symbol>" } }
Subscribe to trades for a specific coin:

Copy
{ "method": "subscribe", "subscription": { "type": "trades", "coin": "<coin_symbol>" } }
Unsubscribing from WebSocket feeds
To unsubscribe from a specific data feed on the Hyperliquid WebSocket API, you need to send an unsubscribe message with the following format:

Copy
{
  "method": "unsubscribe",
  "subscription": { ... }
}
The subscription object should match the original subscription message that was sent when subscribing to the feed. This allows the server to identify the specific feed you want to unsubscribe from. By sending this unsubscribe message, you inform the server to stop sending further updates for the specified feed.

Please note that unsubscribing from a specific feed does not affect other subscriptions you may have active at that time. To unsubscribe from multiple feeds, you can send multiple unsubscribe messages, each with the appropriate subscription details.

Previous
Websocket
Next
For developers
API
Websocket
Post requests
This page describes posting requests using the WebSocket API.

Request format
The WebSocket API supports posting requests that you can normally post through the HTTP API. These requests are either info requests or signed actions. For examples of info request payloads, please refer to the Info endpoint section. For examples of signed action payloads, please refer to the Exchange endpoint section.

To send such a payload for either type via the WebSocket API, you must wrap it as such:

Copy
{
  "method": "post",
  "id": <number>,
  "request": {
    "type": "info" | "action",
    "payload": { ... }
  }
}
Note: The method and id fields are mandatory. It is recommended that you use a unique id for every post request you send in order to track outstanding requests through the channel.

Note: explorer requests are not supported via WebSocket.

Response format
The server will respond to post requests with either a success or an error. For errors, a String is returned mirroring the HTTP status code and description that would have been returned if the request were sent through HTTP.

Copy
{
  "channel": "post",
  "data": {
    "id": <number>,
    "response": {
      "type": "info" | "action" | "error",
      "payload": { ... }
    }
  }
}
Examples
Here are a few examples of subscribing to different feeds using the subscription messages:

Sending an L2Book info request:

Copy
{
  "method": "post",
  "id": 123,
  "request": {
    "type": "info",
    "payload": {
      "type": "l2Book",
      "coin": "ETH",
      "nSigFigs": 5,
      "mantissa": null
    }
  }
}
Sample response:

Copy
{
  "channel": "post",
  "data": {
    "id": <number>,
    "response": {
      "type": "info",
      "payload": {
        "type": "l2Book",
        "data": {
          "coin": "ETH",
          "time": <number>,
          "levels": [
            [{"px":"3007.1","sz":"2.7954","n":1}],
            [{"px":"3040.1","sz":"3.9499","n":1}]
          ]
        }
      }
    }
  }
}
Sending an order signed action request:

Copy
{
  "method": "post",
  "id": 256,
  "request": {
    "type": "action",
    "payload": {
      "action": {
        "type": "order",
        "orders": [{"a": 4, "b": true, "p": "1100", "s": "0.2", "r": false, "t": {"limit": {"tif": "Gtc"}}}],
        "grouping": "na"
      },
      "nonce": 1713825891591,
      "signature": {
        "r": "...",
        "s": "...",
        "v": "..."
      },
      "vaultAddress": "0x12...3"
    }
  }
}
Sample response:

Copy
{
  "channel": "post",
  "data": {
    "id": 256,
    "response": {
      "type":"action",
      "payload": {
        "status": "ok",
        "response": {
          "type": "order",
          "data": {
            "statuses": [
              {
                "resting": {
                  "oid": 88383,
                }
              }
            ]
          }
        }
      }
    }
  }
}Timeouts and heartbeats
This page describes the measures to keep WebSocket connections alive.

The server will close any connection if it hasn't sent a message to it in the last 60 seconds. If you are subscribing to a channel that doesn't receive messages every 60 seconds, you can send heartbeat messages to keep your connection alive. The format for these messages are:

Copy
{ "method": "ping" }
The server will respond with:

Copy
{ "channel": "pong" }
