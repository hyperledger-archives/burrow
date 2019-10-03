# Burrow Proxy

Some client, for example burrow.js, cannot do local signing. They rely on the burrow node to do the signing
on their behalf. This is done in the burrow proxy, which can be both be run in process on the burrow node,
or as a separate process.

The burrow proxy will send the signed transaction to the burrow node. Any client using the proxy does not
need to keep track of account sequence numbers, the proxy does this for you.

The proxy is a proxy for the rpcquery, rpcinfo, rpctransact, and rpcevents burrow node services. It also
has a keys service, which is used by burrow deploy for determining addresses from key names, and by the
burrow keys command line if instructed to do so with `-c`.

## Configuring the internal proxy in burrow toml

Set the `Enabled` to true in the Proxy section.

```toml
[Proxy]
  Enabled = true
  ListenHost = "0.0.0.0"
  ListenPort = "10998"
  AllowBadFilePermissions = false
  KeysDirectory = ".keys"
```

The proxy server will use the keys in the `KeysDirectory`. 

## Starting a burrow proxy as a seperate process

The proxy will need an address of a burrow node to proxy for, and a keys directory can be provided with
`--dir`.

```shell
burrow proxy -c localhost:10997 --dir mykeys
```

## Ensuring that burrow node only has the keys it needs

The burrow node needs the Validator account and the node key to work. The keys directory specified
in the `ValidatorKeys = ...` should only have those two keys.

## Ensuring the proxy only has the keys it needs

As many proxies can be run as necessary, each one with its own set of keys. The proxy should be run
somewhere behing a firewall so it cannot be accessed by anyone who does not have access to its
keys.