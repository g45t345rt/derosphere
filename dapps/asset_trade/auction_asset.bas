Function initStore()
10 STORE("commit_ctr", 0)
20 RETURN
End Function

Function beginStore()
10 MAPSTORE("commit", "{")
20 MAPSTORE("commit_ctr", LOAD("commit_ctr"))
30 RETURN
End Function

Function appendStore(key String, value String)
10 DIM commit as String
20 LET commit = MAPGET("commit")
30 IF commit == "{" THEN GOTO 50
40 LET commit = commit + ","
50 MAPSTORE("commit", commit + "\"" + key + "\":" + value)
60 RETURN
End Function

Function storeString(key String, value String)
10 DIM commit as String
20 STORE(key, value)
30 appendStore(key, "\"" + value + "\"")
40 RETURN
End Function

Function storeUint64(key String, value Uint64)
10 DIM commit as String
20 STORE(key, value)
30 appendStore(key, "" + value + "")
40 RETURN
End Function

Function deleteKey(key String)
10 DIM commit as String
20 DELETE(key)
30 appendStore(key, "-1")
40 RETURN
End Function

Function endStore()
10 DIM ctr as Uint64
20 LET ctr = MAPGET("commit_ctr")
30 STORE("commit_" + ctr, MAPGET("commit") + "}")
40 STORE("commit_ctr", ctr + 1)
50 RETURN
End Function

Function auKey(id Uint64, key String) String
10 RETURN "au_" + id + "_" + key
End Function

Function Initialize() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", SIGNER())
40 STORE("fee_0000000000000000000000000000000000000000000000000000000000000000", 2)
50 STORE("au_ctr", 0)
60 initStore()
70 RETURN 0
End Function

Function CreateAuction(sellAssetId String, bidAssetId String, startAmount Uint64, minBidAmount Uint64, startTimestamp Uint64, duration Uint64) Uint64
10 DIM auId, sellAmount as Uint64
20 LET auId = LOAD("au_ctr")
30 LET sellAmount = ASSETVALUE(HEXDECODE(sellAssetId))
40 IF sellAmount > 0 THEN GOTO 60
50 RETURN 1
60 IF startTimestamp > 0 THEN GOTO 80
70 LET startTimestamp = BLOCK_TIMESTAMP()
80 beginStore()
90 storeUint64(auKey(auId, "startAmount"), startAmount)
100 storeString(auKey(auId, "sellAssetId"), sellAssetId)
110 storeUint64(auKey(auId, "sellAmount"), sellAmount)
120 storeUint64(auKey(auId, "startTimestamp"), startTimestamp)
130 storeUint64(auKey(auId, "duration"), duration)
140 storeString(auKey(auId, "seller"), ADDRESS_STRING(SIGNER()))
150 storeString(auKey(auId, "bidAssetId"), bidAssetId)
160 storeUint64(auKey(auId, "minBidAmount"), minBidAmount)
170 storeUint64(auKey(auId, "currentBid"), 0)
180 storeUint64(auKey(auId, "bidCount"), 0)
190 storeUint64(auKey(auId, "timestamp"), BLOCK_TIMESTAMP())
200 storeUint64(auKey(auId, "close"), 0)
210 endStore()
220 STORE("au_ctr", auId + 1)
230 RETURN 0
End Function

Function SetAuctionMinBid(auId Uint64, amount Uint64) Uint64
10 IF LOAD(auKey(auId, "seller")) == ADDRESS_STRING(SIGNER()) THEN GOTO 30
20 RETURN 1
30 beginStore()
40 storeUint64(auKey(auId, "minBidAmount"), amount)
50 endStore()
60 RETURN 0
End Function

Function CloseAuction(auId Uint64) Uint64
10 IF LOAD(auKey(auId, "seller")) == ADDRESS_STRING(SIGNER()) THEN GOTO 30
20 RETURN 1
30 IF LOAD(auKey(auId, "bidCount")) == 0 THEN GOTO 50
40 RETURN 1
50 IF LOAD(auKey(auId, "close")) == 0 THEN GOTO 70
60 RETURN 1
70 SEND_ASSET_TO_ADDRESS(SIGNER(), LOAD(auKey(auId, "sellAmount")), HEXDECODE(LOAD(auKey(auId, "sellAssetId"))))
80 beginStore()
90 storeUint64(auKey(auId, "close"), 1)
100 endStore()
110 RETURN 0
End Function

Function isAuctionFinished (auId Uint64) Uint64
10 DIM startTimestamp, duration, timestamp as Uint64
20 LET startTimestamp = LOAD(auKey(auId, "startTimestamp"))
30 LET duration = LOAD(auKey(auId, "duration"))
40 LET timestamp = BLOCK_TIMESTAMP()
50 IF timestamp <= startTimestamp + duration THEN GOTO 70
60 RETURN 1
70 RETURN 0
End Function

Function Bid(auId Uint64) Uint64
10 DIM minBidAmount, bidAmount, bidCount, lockedAmount, startAmount, currentBid, startTimestamp, duration, timestamp as Uint64
20 DIM bidAssetId, signerString as String
30 LET signerString = ADDRESS_STRING(SIGNER())
40 LET minBidAmount = LOAD(auKey(auId, "minBidAmount"))
50 LET bidAssetId = LOAD(auKey(auId, "bidAssetId"))
60 LET bidCount = LOAD(auKey(auId, "bidCount"))
70 LET bidAmount = ASSETVALUE(HEXDECODE(bidAssetId))
80 LET startAmount = LOAD(auKey(auId, "startAmount"))
90 LET currentBid = LOAD(auKey(auId, "currentBid"))
100 LET lockedAmount = 0
110 IF EXISTS(auKey(auId, "bid_" + signerString + "_lockedAmount")) == 0 THEN GOTO 130
120 LET lockedAmount = LOAD(auKey(auId, "bid_" + signerString + "_lockedAmount"))
130 LET lockedAmount = lockedAmount + bidAmount
135 IF LOAD(auKey(auId, "close")) == 1 THEN GOTO 150
140 IF isAuctionFinished(auId) == 0 THEN GOTO 160
150 RETURN 1
160 IF bidAmount >= minBidAmount THEN GOTO 180
170 RETURN 1
180 IF currentBid > 0 THEN GOTO 210
190 IF bidAmount >= startAmount THEN GOTO 230
200 RETURN 1
210 IF lockedAmount > currentBid THEN GOTO 230
220 RETURN 1
230 beginStore()
240 LET currentBid = currentBid + (lockedAmount - currentBid)
250 storeUint64(auKey(auId, "bid_" + signerString + "_lockedAmount"), lockedAmount)
260 storeUint64(auKey(auId, "bid_" + signerString + "_bidAmount"), lockedAmount)
270 storeUint64(auKey(auId, "bid_" + signerString + "_timestamp"), BLOCK_TIMESTAMP())
280 storeUint64(auKey(auId, "currentBid"), currentBid)
290 storeUint64(auKey(auId, "bidCount"), bidCount + 1)
300 storeString(auKey(auId, "lastBidder"), signerString)
310 endStore()
320 RETURN 0
End Function

Function CheckoutAuction(auId Uint64) Uint64
10 DIM sellAssetId, bidAssetId, seller, winner as String
20 DIM currentBid, sellAmount, auctionCut as Uint64
30 LET sellAssetId = LOAD(auKey(auId, "sellAssetId"))
40 LET sellAmount = LOAD(auKey(auId, "sellAmount"))
50 LET currentBid = LOAD(auKey(auId, "currentBid"))
60 LET bidAssetId = LOAD(auKey(auId, "bidAssetId"))
70 LET seller = LOAD(auKey(auId, "seller"))
80 LET winner = LOAD(auKey(auId, "lastBidder"))
90 LET auctionCut = 0
100 IF LOAD(auKey(auId, "close")) == 0 THEN GOTO 120
110 RETURN 1
120 IF isAuctionFinished(auId) == 1 THEN GOTO 140
130 RETURN 1
140 IF EXISTS("fee_" + bidAssetId) == 0 THEN GOTO 170
150 LET auctionCut = currentBid * LOAD("fee_" + bidAssetId) / 100
160 LET currentBid = currentBid - auctionCut
170 beginStore()
180 storeUint64(auKey(auId, "close"), 1)
190 storeUint64(auKey(auId, "bid_" + winner + "_lockedAmount"), 0)
200 storeUint64(auKey(auId, "checkoutFee"), auctionCut)
210 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(winner), sellAmount, HEXDECODE(sellAssetId))
220 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(seller), currentBid, HEXDECODE(bidAssetId))
230 SEND_ASSET_TO_ADDRESS(LOAD("owner"), auctionCut, HEXDECODE(bidAssetId))
240 endStore()
250 RETURN 0
End Function

Function RetrieveLockedFunds(auId Uint64) Uint64
10 DIM lockedAmount as Uint64
20 DIM bidAssetId, signerString, winner as String
30 LET bidAssetId = LOAD(auKey(auId, "bidAssetId"))
40 LET signerString = ADDRESS_STRING(SIGNER())
50 LET winner = LOAD(auKey(auId, "lastBidder"))
60 IF isAuctionFinished(auId) == 1 THEN GOTO 80
70 RETURN 1
80 IF winner != signerString THEN GOTO 100
90 RETURN 1
100 LET lockedAmount = LOAD(auKey(auId, "bid_" + signerString + "_lockedAmount"))
110 IF lockedAmount > 0 THEN GOTO 130
120 RETURN 1
130 beginStore()
140 SEND_ASSET_TO_ADDRESS(SIGNER(), lockedAmount, HEXDECODE(bidAssetId))
150 storeUint64(auKey(auId, "bid_" + signerString + "_lockedAmount"), 0)
160 endStore()
170 RETURN 0
End Function

Function SetAssetFee(assetId String, fee Uint64) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF fee <= 100 THEN GOTO 50
40 RETURN 1
50 STORE("fee" + assetId, fee)
60 RETURN 0
End Function

Function UpdateCode(code String) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 UPDATE_SC_CODE(code)
40 RETURN 0
End Function
