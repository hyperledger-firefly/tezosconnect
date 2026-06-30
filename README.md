[![build](https://github.com/hyperledger-firefly/firefly-tezosconnect/actions/workflows/go.yml/badge.svg?branch=main&event=push)](https://github.com/hyperledger-firefly/firefly-tezosconnect/actions/workflows/go.yml?branch=main&event=push)
[![codecov](https://codecov.io/gh/hyperledger-firefly/firefly-tezosconnect/branch/main/graph/badge.svg)](https://codecov.io/gh/hyperledger-firefly/firefly-tezosconnect)
[![Go Report Card](https://goreportcard.com/badge/github.com/hyperledger-firefly/firefly-tezosconnect)](https://goreportcard.com/report/github.com/hyperledger-firefly/firefly-tezosconnect)
[![Hits-of-Code](https://hitsofcode.com/github/hyperledger-firefly/firefly-tezosconnect?branch=main)](https://hitsofcode.com/view/github/hyperledger-firefly/firefly-tezosconnect?branch=main)
[![Go Reference](https://pkg.go.dev/badge/github.com/hyperledger-firefly/firefly-tezosconnect.svg)](https://pkg.go.dev/github.com/hyperledger-firefly/firefly-tezosconnect)
[![License](https://img.shields.io/badge/apache-2.0-blue.svg)](https://github.com/hyperledger-firefly/firefly-tezosconnect/blob/main/LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/hyperledger-firefly/firefly-tezosconnect/badge)](https://scorecard.dev/viewer/?uri=github.com/hyperledger-firefly/firefly-tezosconnect)

# Hyperledger FireFly Tezos Connector

This repo provides a reference implementation of the FireFly Connector API (FFCAPI)
for Tezos blockchain.

See the [Hyperledger Firefly Documentation](https://hyperledger.github.io/firefly/latest/overview/public_vs_permissioned/#firefly-architecture-for-public-chains)
and the [FireFly Transaction Manager](https://github.com/hyperledger/firefly-transaction-manager) repository for
more information.

# License

Apache 2.0

## Transaction signing

Tezosconnect leverages remote transaction signing through a powerful 'signatory' service, offering compatibility with multiple key management solutions. With the flexibility to use AWS KMS, Azure KMS, GCP KMS, Yubi HSM, etc. for transaction signing, you can secure your blockchain transactions efficiently and conveniently.

More info at: https://signatory.io/

## Configuration

For a full list of configuration options see [config.md](./config.md)

## Example configuration

```yaml
connector:
  blockchain:
    # other public RPCs can be found: https://docs.tezos.com/architecture/rpc#public-and-private-rpc-nodes
    rpc: https://ghostnet.ecadinfra.com
    network: ghostnet
    signatory: http://localhost:6732
```

## Blockchain node compatibility

For Tezos connector to function properly, you should check the blockchain node supports the following RPC Methods:

### Chains
- `GET /chains/<chain_id>/blocks/<block_id>/hash`
- `GET /chains/<chain_id>/blocks/<block_id>/operations/<list_offset>`
- `GET /chains/<chain_id>/blocks/<block_id>/operations/<list_offset>/<operation_offset>`
- `POST /chains/<chain_id>/blocks/<block_id>/helpers/forge/operations`
- `POST /chains/<chain_id>/blocks/<block_id>/helpers/scripts/simulate_operation`
- `POST /chains/<chain_id>/blocks/<block_id>/helpers/scripts/run_operation`

### Block monitoring
- `GET /monitor/heads/<chain_id>`

### Injection
- `POST /injection/operation`
