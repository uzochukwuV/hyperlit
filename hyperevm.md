For developers
HyperEVM
Dual-block architecture
The total HyperEVM throughput is split between small blocks produced at a fast rate and large blocks produced at a slower rate. 

The primary motivation behind the dual-block architecture is to decouple block speed and block size when allocating throughput improvements. Users want faster blocks for lower time to confirmation. Builders want larger blocks to include larger transactions such as more complex contract deployments. Instead of a forced tradeoff, the dual-block system will allow simultaneous improvement along both axes. 

The HyperEVM "mempool" is still onchain state with respect to the umbrella L1 execution, but is split into two independent mempools that source transactions for the two block types. The two block types are interleaved with a unique increasing sequence of EVM block numbers. The onchain mempool implementation accepts only the next 8 nonces for each address. Transactions older than 1 day old in the mempool are pruned. 

The initial configuration is set conservatively, and throughput is expected to increase over successive technical upgrades. Fast block duration is set to 1 seconds with a 2M gas limit. Slow block duration is set to 1 minute with a 30M gas limit. 

More precisely, in the definitions above, block duration of x means that the first L1 block for each value   of l1_block_time % x produces an EVM block. 

Developers can deploy larger contracts as follows:

Submit action {"type": "evmUserModify", "usingBigBlocks": true} to direct HyperEVM transactions to big blocks instead of small blocks. Note that this user state flag is set on the HyperCore user level, and must be unset again to target small blocks. Like any action, this requires an existing Core user to send. Like any EOA, the deployer address can be converted to a Core user by receiving a Core asset such as USDC.

Optionally use the JSON-RPC method bigBlockGasPrice in place of gasPrice to estimate base gas fee on the next big block.

For developers
HyperEVM
Interacting with HyperCore
Read precompiles
The testnet EVM provides read precompiles that allows querying HyperCore information. The precompile addresses start at  0x0000000000000000000000000000000000000800 and have methods for querying information such as perps positions, spot balances, vault equity, staking delegations, oracle prices, and the L1 block number.

The values are guaranteed to match the latest HyperCore state at the time the EVM block is constructed.

Attached is a Solidity file L1Read.sol describing the read precompiles. As an example, this call queries the third perp oracle price on testnet:

Copy
cast call 0x0000000000000000000000000000000000000807 0x0000000000000000000000000000000000000000000000000000000000000003 --rpc-url https://rpc.hyperliquid-testnet.xyz/evm
To convert to floating point numbers, divide the returned price by 10^(6 - szDecimals)for perps and 10^(8 - base asset szDecimals) for spot.

Precompiles called on invalid inputs such as invalid assets or vault address will return an error and consume all gas passed into the precompile call frame. Precompiles have a gas cost of 2000 + 65 * output_len.

CoreWriter contract
A system contract is available at 0x3333333333333333333333333333333333333333 for sending transactions from the HyperEVM to HyperCore. It burns ~25,000 gas before emitting a log to be processed by HyperCore as an action. In practice the gas usage for a basic call will be ~47000. A solidity file CoreWriter.sol for the write system contract is attached.

Action encoding details
Byte 1: Encoding version

Currently, only version 1 is supported, but enables future upgrades while maintaining backward compatibility.

Bytes 2-4: Action ID

These three bytes, when decoded as a big-endian unsigned integer, represent the unique identifier for the action.

Remaining bytes: Action encoding

The rest of the bytes constitue the action-specific data. It is always the raw ABI encoding of a sequence of Solidity types

To prevent any potential latency advantages for using HyperEVM to bypass the L1 mempool, order actions and vault transfers sent from CoreWriter are delayed onchain for a few seconds. This has no noticeable effect on UX because the end user has to wait for at least one small block confirmation. These onchain-delayed actions appear twice in the L1 explorer: first as an enqueuing and second as a HyperCore execution.

Action ID
Action
Fields
Solidity Type
Notes
1

Limit order

(asset, isBuy, limitPx, sz, reduceOnly, encodedTif, cloid)

(uint32, bool, uint64, uint64, bool, uint8, uint128)

Tif encoding: 1 for Alo , 2 for Gtc , 3 for Ioc . Cloid encoding: 0 means no cloid, otherwise uses the number as the cloid. limitPx and sz should be sent as 10^8 * the human readable value

2

Vault transfer

(vault, isDeposit, usd)

(address, bool, uint64)

3

Token delegate

(validator, wei, isUndelegate)

(address, uint64, bool)

4

Staking deposit

wei

uint64

5

Staking withdraw

wei

uint64

6

Spot send

(destination, token, wei)

(address, uint64, uint64)

7

USD class transfer

(ntl, toPerp)

(uint64, bool)

8

Finalize EVM Contract

(token, encodedFinalizeEvmContractVariant, createNonce)

(uint64, uint8, uint64)

encodedFinalizeEvmContractVariant 1 for Create, 2 for FirstStorageSlot , 3 for CustomStorageSlot . If Create variant, then createNonce input argument is used.

9

Add API wallet

(API wallet address, API wallet name)

(address, string)

If the API wallet name is empty then this becomes the main API wallet / agent

10

Cancel order by oid

(asset, oid)

(uint32, uint64)

11

Cancel order by cloid

(asset, cloid)

(uint32, uint128)

Below is an example contract that would send an action on behalf of its own contract address on HyperCore, which also demonstrates one way to construct the encoded action in Solidity.

Copy
contract CoreWriterCaller {
    function sendUsdClassTransfer(uint64 ntl, bool toPerp) external {
        bytes memory encodedAction = abi.encode(ntl, toPerp);
        bytes memory data = new bytes(4 + encodedAction.length);
        data[0] = 0x01;
        data[1] = 0x00;
        data[2] = 0x00;
        data[3] = 0x07;
        for (uint256 i = 0; i < encodedAction.length; i++) {
            data[4 + i] = encodedAction[i];
        }
        CoreWriter(0x3333333333333333333333333333333333333333).sendRawAction(data);
    }
}

Happy building. Any feedback is appreciated.`
For developers
HyperEVM
HyperCore <> HyperEVM transfers
Introduction
Spot assets can be sent between HyperCore and the HyperEVM. In the context of these transfers, spot assets on HyperCore are called Core spot while ones on the EVM are called EVM spot. The spot deployer can link their Core spot asset to any ERC20 contract deployed to the EVM. The Core spot asset and ERC20 token can be deployed in either order.

As the native token on HyperCore, HYPE also links to the native HyperEVM balance rather than an ERC20 contract.

System Addresses
Every token has a system address on the Core, which is the address with first byte 0x20 and the remaining bytes all zeros, except for the token index encoded in big-endian format. For example, for token index 200, the system address would be 0x20000000000000000000000000000000000000c8 .

The exception is HYPE, which has a system address of 0x2222222222222222222222222222222222222222 .

Transferring HYPE
HYPE is a special case as the native gas token on the HyperEVM. HYPE is received on the EVM side of a transfer as the native gas token instead of an ERC20 token. To transfer back to HyperCore, HYPE can be sent as a transaction value. The EVM transfer address 0x222..2 is a system contract that emits event Received(address indexed user, uint256 amount) as its payable receive() function. Here user is msg.sender, so this implementation enables both smart contracts and EOAs to transfer HYPE back to HyperCore. Note that there is a small gas cost to emitting this log on the EVM side.

Transferring between Core and EVM
Only once a token is linked, it can be converted between HyperCore and HyperEVM spot using a spotSend action (or via the frontend) and on the EVM by using an ERC20 transfer.

Transferring tokens from HyperCore to HyperEVM can be done using a spotSend action (or via the frontend) with the corresponding system address as the destination. The tokens are credited by a system transaction that calls transfer(recipient, amount) on the linked contract as the system address, where recipient is the sender of the spotSend action. 

Transferring tokens from HyperEVM to HyperCore can be done using an ERC20 transfer with the corresponding system address as the destination. The tokens are credited to the Core based on the emitted Transfer(address from, address to, uint256 value) from the linked contract.

Do not blindly assume accurate fungibility between Core and EVM spot. See Caveats for more details.

Gas costs
A transfer from HyperEVM to HyperCore costs similar gas to the equivalent transfer of the ERC20 token or HYPE to any other address on the HyperEVM that has an existing balance.

A transfer from HyperCore to HyperEVM costs 200k gas at the base gas price of the next HyperEVM block.

Linking Core and EVM Spot Assets
In order for transfers between Core spot and EVM spot to work the token's system address must have the total non-system balance on the other side. For example, to deploy an ERC20 contract for an existing Core spot asset, the system contract should have the entirety of the EVM spot supply equal to the max Core spot supply.

Once this is done the spot deployer needs to send a spot deploy action to link the token to the EVM:

Copy
/**
 * @param token - The token index to link
 * @param address - The address of the ERC20 contract on the evm.
 * @param evmExtraWeiDecimals - The difference in Wei decimals between Core and EVM spot. E.g. Core PURR has 5 weiDecimals but EVM PURR has 18, so this would be 13. evmExtraWeiDecimals should be in the range [-2, 18] inclusive
 */
interface RequestEvmContract {
  type: “requestEvmContract”;
  token: number;
  Address: address;
  evmExtraWeiDecimals: number;
}
After sending this action, HyperCore will store the pending EVM address to be linked. The deployer of the EVM contract must then verify their intention to link to the HyperCore token in one of two ways:

If the EVM contract was deployed from an EOA, the EVM user can send an action using the nonce that was used to deploy the EVM contract.

If the EVM contract was deployed by another contract (e.g. create2 via a multisig), the contract's first storage slot or slot at keccak256("HyperCore deployer") must store the address of a finalizer user.

To finalize the link, the finalizer user sends the following action (note that this not nested in a spot deploy action). In the "create" case, the EVM deployer sends the action. In the "firstStorageSlot" or "customStorageSlot" case, the finalizer must match the value in the corresponding slot.

Copy
/**
 * @param input - One of the EVM deployer options above
 */
interface FinalizeEvmContract {
  type: “finalizeEvmContract”;
  token: number;
  input: {"create": {"nonce": number}} | "firstStorageSlot" | "customStorageSlot"};
}
Caveats
There are currently no checks that the system address has sufficient supply or that the contract is a valid ERC20, so be careful when sending funds.

In particular, the linked contract may have arbitrary bytecode, so it's prudent to verify that its implementation is correct. There are no guarantees about what the transfer call does on the EVM, so make sure to verify the source code and total balance of the linked EVM contract. 

If the EVM contract has extra Wei decimals, then if the relevant log emitted has a value that is not round (does not end in extraEvmWeiDecimals zeros), the non-round amount is burned (guaranteed to be <1 Wei). This is true for both HYPE and any other spot tokens.

Mainnet PURR
Mainnet PURR is deployed as an ERC20 contract at 0x9b498C3c8A0b8CD8BA1D9851d40D186F1872b44E with the following code. It will be linked to PURR on HyperCore once linking is enabled on mainnet.

Copy
// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";

contract Purr is ERC20Permit {
    constructor() ERC20("Purr", "PURR") ERC20Permit("Purr") {
        address initialHolder = 0x2000000000000000000000000000000000000001;
        uint256 initialBalance = 600000000;

        _mint(initialHolder, initialBalance * 10 ** decimals());
    }
}
Final Notes
Attached is a sample script for deploying an ERC20 token to the EVM and linking it to a Core spot token.
For developers
HyperEVM
Wrapped HYPE
A canonical system contract for wrapped HYPE is deployed at 0x555...5. The contract is immutable, with the same source code as wrapped ETH on Ethereum, apart from the token name and symbol. 

The source code for WHYPE is provided below. Note that this is based on the WETH contract on Ethereum mainnet and other EVM chains.

Copy
pragma solidity >=0.4.22 <0.6;

contract WHYPE9 {
  string public name = "Wrapped HYPE";
  string public symbol = "WHYPE";
  uint8 public decimals = 18;

  event Approval(address indexed src, address indexed guy, uint wad);
  event Transfer(address indexed src, address indexed dst, uint wad);
  event Deposit(address indexed dst, uint wad);
  event Withdrawal(address indexed src, uint wad);

  mapping(address => uint) public balanceOf;
  mapping(address => mapping(address => uint)) public allowance;

  function() external payable {
    deposit();
  }

  function deposit() public payable {
    balanceOf[msg.sender] += msg.value;
    emit Deposit(msg.sender, msg.value);
  }

  function withdraw(uint wad) public {
    require(balanceOf[msg.sender] >= wad);
    balanceOf[msg.sender] -= wad;
    msg.sender.transfer(wad);
    emit Withdrawal(msg.sender, wad);
  }

  function totalSupply() public view returns (uint) {
    return address(this).balance;
  }

  function approve(address guy, uint wad) public returns (bool) {
    allowance[msg.sender][guy] = wad;
    emit Approval(msg.sender, guy, wad);
    return true;
  }

  function transfer(address dst, uint wad) public returns (bool) {
    return transferFrom(msg.sender, dst, wad);
  }

  function transferFrom(address src, address dst, uint wad) public returns (bool) {
    require(balanceOf[src] >= wad);

    if (src != msg.sender && allowance[src][msg.sender] != uint(-1)) {
      require(allowance[src][msg.sender] >= wad);
      allowance[src][msg.sender] -= wad;
    }

    balanceOf[src] -= wad;
    balanceOf[dst] += wad;

    emit Transfer(src, dst, wad);

    return true;
  }
}

For developers
HyperEVM
JSON-RPC
The following RPC endpoints are available

net_version

web3_clientVersion

eth_blockNumber

eth_call

only the latest block is supported

eth_chainId

eth_estimateGas

only the latest block is supported

eth_feeHistory

eth_gasPrice

returns the base fee for the next small block

eth_getBalance

only the latest block is supported

eth_getBlockByHash

eth_getBlockByNumber

eth_getBlockReceipts

eth_getBlockTransactionCountByHash

eth_getBlockTransactionCountByNumber

eth_getCode

only the latest block is supported

eth_getLogs

up to 4 topics

up to 50 blocks in query range

eth_getStorageAt

only the latest block is supported

eth_getTransactionByBlockHashAndIndex

eth_getTransactionByBlockNumberAndIndex

eth_getTransactionByHash

eth_getTransactionCount

only the latest block is supported

eth_getTransactionReceipt

eth_maxPriorityFeePerGas

always returns zero currently

eth_syncing

always returns false

The following custom endpoints are available

eth_bigBlockGasPrice

returns the base fee for the next big block

eth_usingBigBlocks

returns whether the address is using big blocks

eth_getSystemTxsByBlockHash  and eth_getSystemTxsByBlockNumber

similar to the "getTransaction" analogs but returns the system transactions that originate from HyperCore

Unsupported requests

Requests that require historical state are not supported at this time on the default RPC implementation. However, independent archive node implementations are available for use, and the GitHub repository has examples on how to get started indexing historical data locally. Note that read precompiles are only recorded for the calls actually made on each block. Hypothetical read precompile results could be obtained from a full L1 replay.

Rate limits: IP based rate limits are the same as the API server. 