publish-only
-------

In addition to the typical request/reply rpc flow, nats-protobuf can also
be configured to support a publish only style.  In this style, multiple
receivers can receive the message, but any responses are dropped and the
rpc call always returns nil, err
