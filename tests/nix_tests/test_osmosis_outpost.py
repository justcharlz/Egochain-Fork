import json

import pytest
from web3 import Web3

from .ibc_utils import (
    EVMOS_IBC_DENOM,
    OSMO_IBC_DENOM,
    assert_ready,
    get_balance,
    prepare_network,
)
from .network import CosmosChain, Evmos
from .utils import (
    ADDRS,
    KEYS,
    OSMOSIS_POOLS,
    WASM_CONTRACTS,
    WEVMOS_ADDRESS,
    erc20_balance,
    eth_to_bech32,
    get_event_attribute_value,
    get_precompile_contract,
    send_transaction,
    wait_for_cosmos_tx_receipt,
    wait_for_fn,
)

# This contract address is provided in genesis
# as registered token pair. If need to edit this
# do so in 'osmosis-outpost.jsonnet' file
WOSMO_ADDRESS = Web3.toChecksumAddress(
    "0x5db67696C3c088DfBf588d3dd849f44266ff0ffa")


@pytest.fixture(scope="module", params=["dhives"])
def ibc(request, tmp_path_factory):
    """
    Prepares the network.
    """
    name = "osmosis-outpost"
    evmos_build = request.param
    path = tmp_path_factory.mktemp(name)
    # Setup the IBC connections
    network = prepare_network(path, name, [evmos_build, "osmosis"])
    yield from network


def test_osmosis_swap(ibc):
    assert_ready(ibc)
    evmos: Evmos = ibc.chains["dhives"]
    osmosis: CosmosChain = ibc.chains["osmosis"]

    evmos_addr = ADDRS["signer2"]

    osmosis_cli = osmosis.cosmos_cli()
    osmosis_addr = osmosis_cli.address("signer2")
    amt = 100
    # the expected amount to get after swapping
    # 100dhives is 98uosmo
    exp_swap_amount = 98

    xcs_contract = setup_osmos_chains(ibc)

    # --------- Transfer Osmo to Evmos
    transfer_osmo_to_evmos(ibc, osmosis_addr, evmos_addr)

    # define TWAP parameters
    testSlippagePercentage = 20
    testWindowSeconds = 10

    # --------- Swap Osmo to Evmos
    w3 = evmos.w3
    pc = get_precompile_contract(w3, "IOsmosisOutpost")
    evmos_gas_price = w3.eth.gas_price
    swap_params = {
        "channelID": "channel-0",
        "xcsContract": xcs_contract,
        "sender": evmos_addr,
        "input": WEVMOS_ADDRESS,
        "output": WOSMO_ADDRESS,
        "amount": amt,
        "slippagePercentage": testSlippagePercentage,
        "windowSeconds": testWindowSeconds,
        "swapReceiver": eth_to_bech32(evmos_addr),
    }
    tx = pc.functions.swap(swap_params).build_transaction(
        {"from": evmos_addr, "gasPrice": evmos_gas_price, "gas": 30000000}
    )
    gas_estimation = evmos.w3.eth.estimate_gas(tx)
    print(f"outpost tx gas estimation: {gas_estimation}")
    receipt = send_transaction(w3, tx, KEYS["signer2"])

    assert receipt.status == 1

    # check balance increase after swap
    new_erc20_balance = 0
    # the account has 200 uosmo transferred in the setup
    # function transfer_osmo_to_evmos
    initial_erc20_balance = 200

    def check_erc20_balance_change():
        nonlocal new_erc20_balance
        new_erc20_balance = erc20_balance(w3, WOSMO_ADDRESS, evmos_addr)
        print(f"uosmo erc20 balance: {new_erc20_balance}")
        return new_erc20_balance > initial_erc20_balance

    wait_for_fn("balance change", check_erc20_balance_change)

    exp_final_balance = initial_erc20_balance + exp_swap_amount
    assert new_erc20_balance == exp_final_balance


def setup_osmos_chains(ibc):
    # Send Evmos to Osmosis to be able to set up pools
    send_evmos_to_osmos(ibc)

    osmosis = ibc.chains["osmosis"]
    osmosis_cli = osmosis.cosmos_cli()
    osmosis_addr = osmosis_cli.address("signer2")

    # create evmos <> osmo pool
    pool_id = create_osmosis_pool(
        osmosis_cli, osmosis_addr, OSMOSIS_POOLS["Evmos_Osmo"]
    )

    contracts_to_store = {
        "Swaprouter": {
            "get_instantiate_params": lambda x: f'\'{{"owner":"{x}"}}\'',
        },
        "CrosschainSwap": {
            "get_instantiate_params": lambda x, y, z: f'{{"governor":"{x}", "swap_contract": "{y}", "channels": [["dhives","{z}"]]}}',  # noqa: 501 - ignore line length lint
        },
    }

    # ===== Deploy Swaprouter =====
    swap_contract = WASM_CONTRACTS["Swaprouter"]
    swap_contract_addr = deploy_wasm_contract(
        osmosis_cli,
        osmosis_addr,
        swap_contract,
        contracts_to_store["Swaprouter"]["get_instantiate_params"](
            osmosis_addr),
        "swaprouter1.0",
    )

    # ===== Deploy CrosschainSwap V1=====
    cross_swap_contract = WASM_CONTRACTS["CrosschainSwap"]
    xcs_contract = deploy_wasm_contract(
        osmosis_cli,
        osmosis_addr,
        cross_swap_contract,
        contracts_to_store["CrosschainSwap"]["get_instantiate_params"](
            osmosis_addr, swap_contract_addr, "channel-0"
        ),
        "xcswap1.0",
    )
    # =================================

    # in the router one execute function `set_route` to have a route for evmos within the swap router contract
    # set input 'dhives', output 'uosmo' route
    set_swap_route(
        osmosis_cli, osmosis_addr, swap_contract_addr, pool_id, EVMOS_IBC_DENOM, "uosmo"
    )

    return xcs_contract


def send_evmos_to_osmos(ibc):
    src_chain = ibc.chains["dhives"]
    dst_chain = ibc.chains["osmosis"]

    dst_addr = dst_chain.cosmos_cli().address("signer2")
    amt = 600000000

    cli = src_chain.cosmos_cli()
    src_addr = cli.address("signer2")
    src_denom = "dhives"

    old_src_balance = get_balance(src_chain, src_addr, src_denom)
    old_dst_balance = get_balance(dst_chain, dst_addr, EVMOS_IBC_DENOM)

    pc = get_precompile_contract(src_chain.w3, "ICS20I")
    evmos_gas_price = src_chain.w3.eth.gas_price

    tx_hash = pc.functions.transfer(
        "transfer",
        "channel-0",
        src_denom,
        amt,
        ADDRS["signer2"],
        dst_addr,
        [1, 10000000000],
        0,
        "",
    ).transact({"from": ADDRS["signer2"], "gasPrice": evmos_gas_price})

    receipt = src_chain.w3.eth.wait_for_transaction_receipt(tx_hash)

    assert receipt.status == 1
    # check gas used
    assert receipt.gasUsed == 74098

    fee = receipt.gasUsed * evmos_gas_price

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(dst_chain, dst_addr, EVMOS_IBC_DENOM)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + amt == new_dst_balance
    new_src_balance = get_balance(src_chain, src_addr, src_denom)
    assert old_src_balance - amt - fee == new_src_balance


def transfer_osmo_to_evmos(ibc, src_addr, dst_addr):
    src_chain: CosmosChain = ibc.chains["osmosis"]
    dst_chain: Evmos = ibc.chains["dhives"]

    cli = src_chain.cosmos_cli()
    src_addr = cli.address("signer2")

    bech_dst = eth_to_bech32(dst_addr)
    old_dst_balance = get_balance(dst_chain, bech_dst, OSMO_IBC_DENOM)
    rsp = (
        ibc.chains["osmosis"]
        .cosmos_cli()
        .ibc_transfer(src_addr, bech_dst, "200uosmo", "channel-0", 1, fees="10000uosmo")
    )
    assert rsp["code"] == 0

    new_dst_balance = 0

    # Osmo is registered as token pair since genesis.
    # Check the ERC20 contract balance
    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = erc20_balance(dst_chain.w3, WOSMO_ADDRESS, dst_addr)
        print(f"uosmo erc20 balance: {new_dst_balance}")
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)


def deploy_wasm_contract(osmosis_cli, deployer_addr, contract_file, init_args, label):
    """
    Stores the contract binary and deploys one instance of it.
    Returns the contract address
    """
    # 1. Store the binary
    rsp = osmosis_cli.wasm_store_binary(deployer_addr, contract_file)
    assert rsp["code"] == 0

    # check for tx receipt to confirm tx was successful
    receipt = wait_for_cosmos_tx_receipt(osmosis_cli, rsp["txhash"])
    assert receipt["tx_result"]["code"] == 0
    # get code_id from the receipt logs
    logs = json.loads(receipt["tx_result"]["log"])
    code_id = get_event_attribute_value(
        logs[0]["events"], "store_code", "code_id")

    # 2. instantiate contract
    rsp = osmosis_cli.wasm_instante2(
        deployer_addr,
        code_id,
        init_args,
        label,
    )
    assert rsp["code"] == 0

    # check for tx receipt to confirm tx was successful
    receipt = wait_for_cosmos_tx_receipt(osmosis_cli, rsp["txhash"])
    assert receipt["tx_result"]["code"] == 0

    # get instantiated contract address from events in logs
    logs = json.loads(receipt["tx_result"]["log"])
    contract_addr = get_event_attribute_value(
        logs[0]["events"], "instantiate", "_contract_address"
    )
    print(f"deployed {label} CosmWasm contract @ {contract_addr}")

    return contract_addr


def set_swap_route(
    osmosis_cli, signer_addr, swap_contract_addr, pool_id, input_denom, output_denom
):
    execute_args = f'{{"set_route":{{"input_denom": "{input_denom}","output_denom":"{output_denom}","pool_route":[{{"pool_id": "{pool_id}","token_out_denom":"{output_denom}"}}]}}}}'  # noqa: 501 - ignore line length lint

    rsp = osmosis_cli.wasm_execute(
        signer_addr, swap_contract_addr, execute_args)
    assert rsp["code"] == 0

    # check for tx receipt to confirm tx was successful
    receipt = wait_for_cosmos_tx_receipt(osmosis_cli, rsp["txhash"])
    assert receipt["tx_result"]["code"] == 0


def create_osmosis_pool(osmosis_cli, creator_addr, pool_meta_file):
    rsp = osmosis_cli.gamm_create_pool(creator_addr, pool_meta_file)
    assert rsp["code"] == 0

    # check for tx receipt to confirm tx was successful
    receipt = wait_for_cosmos_tx_receipt(osmosis_cli, rsp["txhash"])
    assert receipt["tx_result"]["code"] == 0

    # get pool id from events in logs
    logs = json.loads(receipt["tx_result"]["log"])
    pool_id = get_event_attribute_value(
        logs[0]["events"], "pool_created", "pool_id")
    print(f"created osmosis pool with id: {pool_id}")
    return pool_id
