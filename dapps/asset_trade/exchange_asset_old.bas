Function storeCommitString(action String, key String, value String)
10 DIM commit_count as Uint64
20 LET commit_count = MAPGET("commit_count")
30 STORE("commit_" + commit_count, action + "::" + key + "::" + value)
40 MAPSTORE("commit_count", commit_count + 1)
50 RETURN
End Function

Function storeCommitInt(action String, key String, value Uint64)
10 DIM commit_count as Uint64
20 LET commit_count = MAPGET("commit_count")
30 STORE("commit_" + commit_count, action + "::" + key + "::" + value)
40 MAPSTORE("commit_count", commit_count + 1)
50 RETURN
End Function

Function initCommit()
10 STORE("commit_count", 0)
20 RETURN
End Function

Function beginCommit()
10 MAPSTORE("commit_count", LOAD("commit_count"))
20 RETURN
End Function

Function endCommit()
10 STORE("commit_count", MAPGET("commit_count"))
20 RETURN
End Function

Function storeStateString(key String, value String)
10 STORE("state_" + key, value)
20 storeCommitString("S", "state_" + key, value)
30 RETURN
End Function

Function storeStateInt(key String, value Uint64)
10 STORE("state_" + key, value)
20 storeCommitInt("S", "state_" + key, value)
30 RETURN
End Function

Function loadStateString(key String) String
10 RETURN LOAD("state_" + key)
End Function

Function loadStateInt(key String) Uint64
10 RETURN LOAD("state_" + key)
End Function

Function stateExists(key String) Uint64
10 RETURN EXISTS("state_" + key)
End Function

Function odKey(id Uint64, key String) String
10 RETURN "od_" + id + "_" + key
End Function

Function Initialize() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", SIGNER())
40 STORE("fee_0000000000000000000000000000000000000000000000000000000000000000", 5)
50 STORE("od_ctr", 0)
60 initCommit()
70 RETURN 0
End Function

Function CreateOrder(odType String, assetId String, priceAssetId String, unitPrice Uint64, expireTimestamp Uint64, oneTxOnly Uint64) Uint64
10 DIM odId, assetAmount, assetBalance, priceBalance, priceAmount as Uint64
20 LET odId = LOAD("od_ctr")
30 IF oneTxOnly <= 1 THEN GOTO 50
40 RETURN 1
50 IF odType != "sell" THEN GOTO 120
60 LET assetBalance = ASSETVALUE(HEXDECODE(assetId))
70 LET assetAmount = assetBalance
80 LET priceAmount = unitPrice * assetAmount
90 LET priceBalance = 0
100 IF assetAmount > 0 THEN GOTO 200
110 RETURN 1
120 IF odType != "buy" THEN GOTO 190
140 LET priceBalance = ASSETVALUE(HEXDECODE(priceAssetId))
150 LET priceAmount = priceBalance
160 LET assetAmount = priceBalance / unitPrice
170 LET assetBalance = 0
180 IF priceBalance > 0 THEN GOTO 200
190 RETURN 1
200 beginCommit()
210 storeStateInt(odKey(odId, "assetAmount"), assetAmount)
220 storeStateInt(odKey(odId, "assetBalance"), assetBalance)
230 storeStateString(odKey(odId, "assetId"), assetId)
240 storeStateString(odKey(odId, "type"), odType)
250 storeStateString(odKey(odId, "priceAssetId"), priceAssetId)
260 storeStateInt(odKey(odId, "priceAmount"), priceAmount)
270 storeStateInt(odKey(odId, "priceBalance"), priceBalance)
280 storeStateInt(odKey(odId, "unitPrice"), unitPrice)
290 storeStateString(odKey(odId, "creator"), ADDRESS_STRING(SIGNER()))
300 storeStateInt(odKey(odId, "close"), 0)
310 storeStateInt(odKey(odId, "timestamp"), BLOCK_TIMESTAMP())
320 storeStateInt(odKey(odId, "expireTimestamp"), expireTimestamp)
330 storeStateInt(odKey(odId, "oneTxOnly"), oneTxOnly)
340 storeStateInt(odKey(odId, "txCtr"), 0)
350 endCommit()
360 STORE("od_ctr", odId + 1)
370 RETURN 0
End Function

Function CloseOrder(odId Uint64) Uint64
10 DIM signer, odType as String
20 LET signer = SIGNER()
30 LET odType = loadStateString(odKey(odId, "type"))
40 IF loadStateString(odKey(odId, "creator")) == ADDRESS_STRING(signer) THEN GOTO 60
50 RETURN 1
60 IF loadStateInt(odKey(odId, "close")) == 0 THEN GOTO 80
70 RETURN 1
80 beginCommit()
90 IF odType != "sell" THEN GOTO 120
100 SEND_ASSET_TO_ADDRESS(signer, loadStateInt(odKey(odId, "assetBalance")), HEXDECODE(loadStateString(odKey(odId, "assetId"))))
110 storeStateInt(odKey(odId, "assetBalance"), 0)
120 IF odType != "buy" THEN GOTO 150
130 SEND_ASSET_TO_ADDRESS(signer, loadStateInt(odKey(odId, "priceBalance")), HEXDECODE(loadStateString(odKey(odId, "priceAssetId"))))
140 storeStateInt(odKey(odId, "priceBalance"), 0)
150 storeStateInt(odKey(odId, "close"), 1)
160 endCommit()
170 RETURN 0
End Function

Function BuyOrSell(odId Uint64) Uint64
10 DIM assetId, priceAssetId, creator, signer, txId, odType as String
20 DIM ctr, unitPrice, amountSent, amountReceived, priceCut, timestamp, assetBalance, priceBalance, expireTimestamp, assetSent, assetReceived as Uint64
30 IF loadStateInt(odKey(odId, "close")) == 0 THEN GOTO 50
40 RETURN 1
50 LET timestamp = BLOCK_TIMESTAMP()
60 LET expireTimestamp = loadStateInt(odKey(odId, "expireTimestamp"))
70 IF expireTimestamp == 0 THEN GOTO 100
80 IF timestamp < expireTimestamp THEN GOTO 100
90 RETURN 1
100 LET assetId = loadStateString(odKey(odId, "assetId"))
110 LET assetBalance = loadStateInt(odKey(odId, "assetBalance"))
120 LET priceAssetId = loadStateString(odKey(odId, "priceAssetId"))
130 LET odType = loadStateString(odKey(odId, "type"))
140 LET priceBalance = loadStateInt(odKey(odId, "priceBalance"))
150 LET unitPrice = loadStateInt(odKey(odId, "unitPrice"))
160 LET creator = loadStateString(odKey(odId, "creator"))
170 LET signer = SIGNER()
180 LET priceCut = 0
190 LET assetSent = 0
200 LET assetReceived = 0
210 LET amountSent = 0
220 LET amountReceived = 0
230 beginCommit()
240 IF odType != "sell" THEN GOTO 370
250 LET amountSent = ASSETVALUE(HEXDECODE(priceAssetId))
255 IF amountSent == 0 THEN GOTO 280
260 IF loadStateInt(odKey(odId, "oneTxOnly")) == 0 THEN GOTO 290
270 IF amountSent == assetBalance * unitPrice THEN GOTO 290
280 RETURN 1
290 LET assetReceived = amountSent / unitPrice
295 IF assetBalance < assetReceived THEN GOTO 410
300 LET assetBalance = assetBalance - assetReceived
310 IF EXISTS("fee_" + priceAssetId) == 0 THEN GOTO 330
320 LET priceCut = amountSent * LOAD("fee_" + priceAssetId) / 100
330 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(creator), amountSent - priceCut, HEXDECODE(priceAssetId))
340 SEND_ASSET_TO_ADDRESS(signer, assetReceived, HEXDECODE(assetId))
345 storeStateInt(odKey(odId, "assetBalance"), assetBalance)
350 IF assetBalance > 0 THEN GOTO 370
360 storeStateInt(odKey(odId, "close"), 1)
370 IF odType != "buy" THEN GOTO 500
380 LET assetSent = ASSETVALUE(HEXDECODE(assetId))
385 IF assetSent == 0 THEN GOTO 410
390 IF loadStateInt(odKey(odId, "oneTxOnly")) == 0 THEN GOTO 420
400 IF assetSent == assetBalance THEN GOTO 420
410 RETURN 1
420 LET amountReceived = assetSent * unitPrice
425 IF priceBalance < amountReceived THEN GOTO 410
430 LET priceBalance = priceBalance - amountReceived
440 IF EXISTS("fee_" + priceAssetId) == 0 THEN GOTO 460
450 LET priceCut = amountReceived * LOAD("fee_" + priceAssetId) / 100
460 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(creator), assetSent, HEXDECODE(assetId))
470 SEND_ASSET_TO_ADDRESS(signer, amountReceived - priceCut, HEXDECODE(priceAssetId))
475 storeStateInt(odKey(odId, "priceBalance"), priceBalance)
480 IF priceBalance > 0 THEN GOTO 500
490 storeStateInt(odKey(odId, "close"), 1)
500 SEND_ASSET_TO_ADDRESS(LOAD("owner"), priceCut, HEXDECODE(priceAssetId))
530 LET ctr = loadStateInt(odKey(odId, "txCtr"))
540 storeStateString(odKey(odId, "tx_" + ctr + "_sender"), ADDRESS_STRING(signer))
550 storeStateInt(odKey(odId, "tx_" + ctr + "_assetSent"), assetSent)
560 storeStateInt(odKey(odId, "tx_" + ctr + "_assetReceived"), assetReceived)
570 storeStateInt(odKey(odId, "tx_" + ctr + "_amountSent"), amountSent)
580 storeStateInt(odKey(odId, "tx_" + ctr + "_amountReceived"), amountReceived)
590 storeStateInt(odKey(odId, "tx_" + ctr + "_timestamp"), timestamp)
600 storeStateString(odKey(odId, "tx_" + ctr + "_txId"), HEX(TXID()))
610 storeStateInt(odKey(odId, "tx_" + ctr + "_fee"), priceCut)
620 storeStateInt(odKey(odId, "txCtr"), ctr + 1)
630 endCommit()
640 RETURN 0
End Function

Function SetAssetFee(assetId String, fee Uint64) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF fee <= 100 THEN GOTO 50
40 RETURN 1
50 STORE("fee_" + assetId, fee)
60 RETURN 0
End Function

Function UpdateCode(code String) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 UPDATE_SC_CODE(code)
40 RETURN 0
End Function
