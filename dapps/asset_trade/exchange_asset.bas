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

Function odKey(id Uint64, key String) String
10 RETURN "od_" + id + "_" + key
End Function

Function Initialize() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", SIGNER())
40 STORE("fee_0000000000000000000000000000000000000000000000000000000000000000", 5)
50 STORE("od_ctr", 0)
60 initStore()
70 RETURN 0
End Function

Function CreateOrder(odType String, assetId String, priceAssetId String, unitPrice Uint64, expireTimestamp Uint64, oneTxOnly Uint64) Uint64
10 DIM odId, amount as Uint64
20 LET odId = LOAD("od_ctr")
30 IF oneTxOnly > 1 THEN GOTO 280
50 IF odType != "sell" THEN GOTO 70
60 LET amount = ASSETVALUE(HEXDECODE(assetId))
70 IF odType != "buy" THEN GOTO 90
80 LET amount = ASSETVALUE(HEXDECODE(priceAssetId))
90 IF amount == 0 THEN GOTO 280
110 beginStore()
120 storeUint64(odKey(odId, "amountSent"), 0)
130 storeUint64(odKey(odId, "amount"), amount)
140 storeString(odKey(odId, "assetId"), assetId)
150 storeString(odKey(odId, "type"), odType)
160 storeString(odKey(odId, "priceAssetId"), priceAssetId)
170 storeUint64(odKey(odId, "unitPrice"), unitPrice)
180 storeString(odKey(odId, "creator"), ADDRESS_STRING(SIGNER()))
190 storeUint64(odKey(odId, "close"), 0)
200 storeUint64(odKey(odId, "timestamp"), BLOCK_TIMESTAMP())
210 storeUint64(odKey(odId, "expireTimestamp"), expireTimestamp)
220 storeUint64(odKey(odId, "oneTxOnly"), oneTxOnly)
230 storeUint64(odKey(odId, "txCount"), 0)
240 storeString(odKey(odId, "txId"), HEX(TXID()))
250 endStore()
260 STORE("od_ctr", odId + 1)
270 RETURN 0
280 RETURN 1
End Function

Function CloseOrder(odId Uint64) Uint64
5 DIM retrieveFunds as Uint64
10 DIM signer, odType as String
20 LET signer = SIGNER()
30 LET odType = LOAD(odKey(odId, "type"))
40 IF ADDRESS_STRING(signer) != LOAD(odKey(odId, "creator")) THEN GOTO 150
50 IF LOAD(odKey(odId, "close")) == 1 THEN GOTO 150
60 beginStore()
70 LET retrieveFunds = LOAD(odKey(odId, "amount")) - LOAD(odKey(odId, "amountSent"))
80 IF odType != "sell" THEN GOTO 100
90 SEND_ASSET_TO_ADDRESS(signer, retrieveFunds, HEXDECODE(LOAD(odKey(odId, "assetId"))))
100 IF odType != "buy" THEN GOTO 120
110 SEND_ASSET_TO_ADDRESS(signer, retrieveFunds, HEXDECODE(LOAD(odKey(odId, "priceAssetId"))))
120 storeUint64(odKey(odId, "close"), 1)
130 endStore()
140 RETURN 0
150 RETURN 1
End Function

Function BuyOrSell(odId Uint64) Uint64
10 DIM assetId, priceAssetId, creator, signer, odType as String
20 DIM ctr, unitPrice, txAmountSent, txAmountReceived, priceCut, timestamp, amount, expireTimestamp, amountSent as Uint64
30 IF LOAD(odKey(odId, "close")) == 0 THEN GOTO 50
40 RETURN 1
50 LET timestamp = BLOCK_TIMESTAMP()
60 LET expireTimestamp = LOAD(odKey(odId, "expireTimestamp"))
70 IF expireTimestamp == 0 THEN GOTO 100
80 IF timestamp < expireTimestamp THEN GOTO 100
90 RETURN 1
100 LET assetId = LOAD(odKey(odId, "assetId"))
110 LET amount = LOAD(odKey(odId, "amount"))
120 LET priceAssetId = LOAD(odKey(odId, "priceAssetId"))
130 LET odType = LOAD(odKey(odId, "type"))
140 LET amountSent = LOAD(odKey(odId, "amountSent"))
150 LET unitPrice = LOAD(odKey(odId, "unitPrice"))
160 LET creator = LOAD(odKey(odId, "creator"))
170 LET signer = SIGNER()
180 LET priceCut = 0
190 LET txAmountSent = 0
200 LET txAmountReceived = 0
210 beginStore()
220 IF odType != "sell" THEN GOTO 350
230 LET txAmountSent = ASSETVALUE(HEXDECODE(priceAssetId))
240 IF txAmountSent == 0 THEN GOTO 270
250 IF LOAD(odKey(odId, "oneTxOnly")) == 0 THEN GOTO 280
260 IF txAmountSent == amount * unitPrice THEN GOTO 280
270 RETURN 1
280 LET txAmountReceived = txAmountSent / unitPrice
290 IF txAmountReceived == 0 THEN GOTO 400
300 IF txAmountReceived > amount - amountSent THEN GOTO 400
310 IF EXISTS("fee_" + priceAssetId) == 0 THEN GOTO 330
320 LET priceCut = txAmountSent * LOAD("fee_" + priceAssetId) / 100
330 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(creator), txAmountSent - priceCut, HEXDECODE(priceAssetId))
340 SEND_ASSET_TO_ADDRESS(signer, txAmountReceived, HEXDECODE(assetId))
350 IF odType != "buy" THEN GOTO 480
360 LET txAmountSent = ASSETVALUE(HEXDECODE(assetId))
370 IF txAmountSent == 0 THEN GOTO 400
380 IF LOAD(odKey(odId, "oneTxOnly")) == 0 THEN GOTO 410
390 IF txAmountSent == amount / unitPrice THEN GOTO 410
400 RETURN 1
410 LET txAmountReceived = txAmountSent * unitPrice
420 IF txAmountReceived == 0 THEN GOTO 400
430 IF txAmountReceived > amount - amountSent THEN GOTO 400
440 IF EXISTS("fee_" + priceAssetId) == 0 THEN GOTO 460
450 LET priceCut = txAmountReceived * LOAD("fee_" + priceAssetId) / 100
460 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(creator), txAmountSent, HEXDECODE(assetId))
470 SEND_ASSET_TO_ADDRESS(signer, txAmountReceived - priceCut, HEXDECODE(priceAssetId))
480 LET amountSent = amountSent + txAmountReceived
490 storeUint64(odKey(odId, "amountSent"), amountSent)
500 IF amountSent < amount THEN GOTO 520
510 storeUint64(odKey(odId, "close"), 1)
520 SEND_ASSET_TO_ADDRESS(LOAD("owner"), priceCut, HEXDECODE(priceAssetId))
530 LET ctr = LOAD(odKey(odId, "txCount"))
540 storeString(odKey(odId, "tx_" + ctr + "_sender"), ADDRESS_STRING(signer))
550 storeUint64(odKey(odId, "tx_" + ctr + "_amountSent"), txAmountSent)
560 storeUint64(odKey(odId, "tx_" + ctr + "_amountReceived"), txAmountReceived)
570 storeUint64(odKey(odId, "tx_" + ctr + "_timestamp"), timestamp)
580 storeString(odKey(odId, "tx_" + ctr + "_txId"), HEX(TXID()))
590 storeUint64(odKey(odId, "tx_" + ctr + "_fee"), priceCut)
600 storeUint64(odKey(odId, "txCount"), ctr + 1)
610 endStore()
620 RETURN 0
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
