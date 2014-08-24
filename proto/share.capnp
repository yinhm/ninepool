@0x9566fc320a1d7862;

using Go = import "go.capnp";

$Go.package("proto");

struct Prefix {
    app @0: UInt8;
    table @1: UInt16;
    unixtime @2: Int32;
}

struct Round {
    roundId @0: Int64;
    orderId @1: Int64;
    userId @2: Int64;
    createTime @3: Int64;
    startTime @4: Int64;
    stopTime @5: Int64;
    mps @6: Int32;
    ghs @7: Int32;
}

struct Share {
    username @0: Text;
    jobId @1: Text;
    pool  @2: Text;
    header  @3: Text;
    diff  @4: Float64;
    isBlock  @5: Bool;
    accepted  @6: Bool;
    extraNonce1 @7: Text;
    extraNonce2 @8: Text;
    ntime  @9: Text;
    nonce  @10: Text;
    created  @11: Int64;
}
