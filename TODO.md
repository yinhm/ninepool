
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
 - 
