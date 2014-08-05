@0x9566fc320a1d7862;

using Go = import "go.capnp";

$Go.package("store");

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
    ntime  @9: Int64;
    nonce  @10: Text;
}
