
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



== 买家保护系统 ==

 - 预约算力制
 - VIP优先同价位制
 - 自动加价系统
 - FEE0.2%  挂单设置0.4%自动VIP
 - 全池算力10G VIP算力3G不能再获得VIP资格



=== Nicehash ===

Not true.
1. Each provided job is valid for at least 1.5 seconds
2. Each job has 0.8 seconds stale window
3. We reward miners with fake accepted shares if jobs are being switched fast
4. If share is rejected by NiceHash, it is not sent to the pool    


>> How exactly does the order system work? For instance, if I configure my miner with p=0.70 but there are orders with 0.8x, will I always mine the highest order automatically?

No, all miners are paid average price, which can be seen on front page under "Currently paying".
