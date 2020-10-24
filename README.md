
![Build](https://github.com/gocardano/shelley-socket-proxy/workflows/Build/badge.svg?branch=master)

# shelley-socket-proxy

Want to see the data that gets transmitted between `cardano-cli` and `cardano-node` via unix socket?  This tool does it for you!

It essentially proxies via another unix socket, and prints out the payload that gets transmitted.

# Usage

Below is an example of running this on a Mac.  After starting the proxy, query the `cardano-node`.

## Start the Proxy

```
$ shelley-socket-proxy \
	-source-socket-file ~/read.socket \
	-destination-socket-file ~/Library/Application\ Support/Daedalus\ Mainnet/cardano-node.socket 

>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> R E Q U E S T >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
==========================================================================================
Header: Transmission Time: [794923165], Mode: [0], Protocol ID: [0], Payload Length: [17]
------------------------------------------------------------------------------------------
Array: [2]
  PositiveInteger8(0)
  Map - Items: [2]
    - key: PositiveInteger8(1) / value: PositiveInteger32(764824073)
    - key: PositiveInteger16(32770) / value: PositiveInteger32(764824073)
==========================================================================================


<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< R E S P O N S E <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
==========================================================================================
Header: Transmission Time: [794929988], Mode: [1], Protocol ID: [0], Payload Length: [10]
------------------------------------------------------------------------------------------
Array: [3]
  PositiveInteger8(1)
  PositiveInteger16(32770)
  PositiveInteger32(764824073)
==========================================================================================


>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> R E Q U E S T >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
==========================================================================================
Header: Transmission Time: [794937820], Mode: [0], Protocol ID: [5], Payload Length: [2]
------------------------------------------------------------------------------------------
Array: [1]
  PositiveInteger8(0)
==========================================================================================


<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< R E S P O N S E <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
==========================================================================================
Header: Transmission Time: [794981374], Mode: [1], Protocol ID: [5], Payload Length: [49]
------------------------------------------------------------------------------------------
Array: [3]
  PositiveInteger8(3)
  Array: [0]
  Array: [2]
    Array: [2]
      PositiveInteger32(12014123)
      ByteString - Length: [32]; Value: [bd6e9e1ad793d6dcf25976133216c704f0fcb7ea3cb80b375dbf53acdb198a48];
    PositiveInteger32(4862240)
==========================================================================================


>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> R E Q U E S T >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
==========================================================================================
Header: Transmission Time: [794982467], Mode: [0], Protocol ID: [5], Payload Length: [2]
------------------------------------------------------------------------------------------
Array: [1]
  PositiveInteger8(7)
==========================================================================================
```

## Query Cardano Node

```
$ export CARDANO_NODE_SOCKET_PATH=~/read.socket 

$ cardano-cli shelley query tip --shelley-mode --mainnet
{
    "blockNo": 4862240,
    "headerHash": "bd6e9e1ad793d6dcf25976133216c704f0fcb7ea3cb80b375dbf53acdb198a48",
    "slotNo": 12014123
}
```

# Developer Notes

The packets transmitted to/from `cardano-node` is multiplexed into an envelope (referred to as container in this codebase/library).

```
Container represent a message envelope that is sent and receive
to/from the Shelley node.  It contains one or more segments as payload.
The following is the wire format of a container:

+---------------------------------------------------------------+
|0|1|2|3|4|5|6|7|8|9|0|1|2|3|4|5|6|7|8|9|0|1|2|3|4|5|6|7|8|9|0|1|
+---------------------------------------------------------------+
|                       TRANSMISSION TIME                       |
+---------------------------------------------------------------+
|M|     MINI PROTOCOL ID      |         PAYLOAD LENGTH          |
+---------------------------------------------------------------+
|                                                               |
|                       PAYLOAD of n BYTES                      |
|                                                               |
+---------------------------------------------------------------+

Container header:
- Transmission Time The transmission time is a time stamp based the wall clock
  of the peer with a resolution of one microsecond.
- Mini Protocol ID The unique ID of the mini protocol as in Table 3.2.
- Payload Length The payload length is the size of the segment payload in Bytes.
  The maximum payload length that is supported by the multiplexing wire format
  is 2^16 − 1. Note, that an instance of the protocol can choose a smaller
  limit for the size of segments it transmits.
- Mode The single bit M (the mode) is used to distinct the dual instances of a
  mini protocol. The mode is set to 0 in segments from the initiator, i.e. the
  side that initially has agency and 1 in segments from the responder.
```

The payload is encoded as CBOR (see https://en.wikipedia.org/wiki/CBOR, and https://tools.ietf.org/html/rfc7049).


# References

* https://hydra.iohk.io/build/4110312/download/2/network-spec.pdf
* https://github.com/input-output-hk/ouroboros-network/blob/master/ouroboros-network/test/messages.cddl
