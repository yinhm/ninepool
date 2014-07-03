
 - detect disconnected pool
 - detect inavtive worker, last activity ago
 - multiple workers on single stratum connection.
   * multi-workers are from another tier of proxy;
   * worker can have many sub-names;
   * worker name only valid if:
     - NormalWorkerName, or
     - NomalPrefixedWorkerName.subfix
   * warning user if this is not supported;
 - Authentication should work also for unsubscribed socket. stratum-mining-proxy issues #32


== DIFF1

scrypt:
0x0000ffff00000000000000000000000000000000000000000000000000000000

SHA256
0x00000000ffff0000000000000000000000000000000000000000000000000000

QUARK
0x000000ffff000000000000000000000000000000000000000000000000000000

x11
0x00000000ffff0000000000000000000000000000000000000000000000000000

    https://github.com/zone117x/node-open-mining-portal/issues/94
    For x11 algo, just like quark, you need to set diff super low since the mining apps don't use a share multiplier. Set diff to 0.0001 and you will start seeing shares.
    
    
    https://github.com/zone117x/node-open-mining-portal/issues/287
    
    for nomp ... i would do this... in pages/workers.httml
    
    {{? it.stats.pools[pool].algorithm == 'x11'}} {{= workerstat.shares * 256 }} {{??}} {{=workerstat.shares}} {{?}}
    
    x11 abusing happens. really! #54
    https://github.com/zone117x/node-stratum-pool/issues/54
    

  * Support for proxy virtual devices has been extended to include the stratum protocol when the upstream pool selected is also stratum and supplies sufficient extranonce2 space. If the upstream pool does not meet this criteria, stratum clients will be disconnected and new ones will fail to subscribe. You can take advantage of this to failover to the getwork proxy. Support for upstream getwork pools is impossble, but GBT is planned.
