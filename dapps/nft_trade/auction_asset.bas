Function auctionKey(id Uint64, key String) String
10 RETURN "auction_" + id + "_" + key
End Function

Function Initialize() Uint64
10 STORE("owner", SIGNER())
20 STORE("dero_fee", 5)
30 STORE("auction_count", 0)
40 RETURN 0
End Function

Function Auction(assetId String, startAmount Uint64, startTimestamp Uint64, duration Uint64, bidIncAmount Uint64) Uint64
10 DIM id, start_timestamp as Uint64
20 LET id = LOAD("auction_count")
30 IF ASSETVALUE(assetId) == 1 THEN GOTO
40 RETURN 1
50 STORE(itemKey(id, "startAmount"), startAmount)
60 STORE(itemKey(id, "assetId"), assetId)
70 STORE(itemKey(id, "startTimestamp"), startTimestamp)
80 STORE(itemKey(id, "duration"), duration)
90 STORE(itemKey(id, "seller"), SIGNER())
100 STORE(itemKey(id, "bidIncAmount"), bidIncAmount)
110 STORE(itemKey(id, "bid"), startAmount)
120 STORE(itemKey(id, "bidCount"), 0)
130 STORE("auction_count", id + 1)
140 RETURN 0
End Function

Function CancelAuction(id String) Uint64
10 IF LOAD(itemKey(id, "seller")) == SIGNER() THEN GOTO 30
20 RETURN 1
30 SEND_ASSET_TO_ADDRESS(SIGNER(), 1, LOAD(itemKey(id, "assetId")))
40 STORE(itemKey(id, "startAmount"))
50 STORE(itemKey(id, "assetId"))
60 STORE(itemKey(id, "startTimestamp"))
70 STORE(itemKey(id, "duration"))
80 STORE(itemKey(id, "seller"))
90 STORE(itemKey(id, "bidIncAmount"))
100 RETURN 0
End Function

Function Bid(id String) Uint64
End Function

Function Checkout(id String) Uint64
End Function

Function SetDeroFee(fee Uint64) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF fee <= 100 THEN GOTO 50
40 RETURN 1
50 STORE("dero_fee", fee)
60 RETURN 0
End Function

Function UpdateCode(code String) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 UPDATE_SC_CODE(code)
40 RETURN 0
End Function
