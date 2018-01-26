# Minichain project

Project implements basic ideas of blockchain that
stores transactions composed to blocks and linked
in chain. Each transaction and block is hashed using
SHA-256 algorithm.

## API

### Transaction endpoint

`/tx?key=<key>&<value>`

```json
    {
        "id" : "hash-value",
        "key": "hello",
        "value": "world",
        "timestamp" : "epoch-time-stamp"
    }
```

Response code `202` Accepted

### Search endpoint

`/search?key=<key>`

Searches for all transactions that have particular key

Response body

```json
    [
        {
            "id" : "hash-value",
            "key": "hello",
            "value": "world",
            "timestamp" : "epoch-time-stamp"
        },
        {
            "id" : "hash-value",
            "key": "banana",
            "value": "apple",
            "timestamp" : "epoch-time-stamp"
        },
    ]
```

Response codes 200, 404, 504

## Blockchain layout

All blocks are appended to the file and file record
have following format.


```
     4 byte      n bytes       32 byte
  +----------+------------+---------------+
  |block_size| block data |   blockhash   |
  +----------+------------+---------------+
```

1. Header of record has size of 4 bytes that contains a number of
bytes in block.
2. Block data contains serialized json of block
3. Blockhash contains sha-256 hash of block and it helps
   to restart blockchain and know set prev block hash.

## Configuration

```
[Main]
# Level of log verbosity 0 - panic, 1 - fatal, 2 - error, 3 - warn, 4 -info, 5 - debug
LogLevel=4

[BlockChain]
# Amount of transactions that can be included to block
BlockSize=1
# Flush timeout in seconds when block sync to disk
Timeout=60
# Restrictions on transactions size
KeyMaxSize=5
ValueMaxSize=512
IndexOn=false
# Path to file that contains blockchain records
DataFile="../data/blockchain.dat"

[Http]
ListenStr="0.0.0.0:8080"
# Read/Write timeout in seconds
Timeout=10
```

## Run server

Create datadir if not exists

`mkdir data`

Change dir

`cd cmd/`

Build and run

`go build && ./cmd`

Default config file name is `config.toml` in cmd directory

Server handles `SIGINT` and ensures that all request received
are processed and flushed to disk.