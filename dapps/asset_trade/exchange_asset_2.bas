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
10 IF EXISTS("owner") != 0 THEN GOTO 70
20 STORE("owner", SIGNER())
30 STORE("fee_0000000000000000000000000000000000000000000000000000000000000000", 5)
40 STORE("od_ctr", 0)
50 initStore()
60 RETURN 0
70 RETURN 1
End Function

Function CreateOrder(odType String, assetId String, tradeAssetId String, unitPrice Uint64, expireTimestamp Uint64, oneTxOnly Uint64) Uint64
10 DIM odId, lockAmount as Uint64
20 LET odId = LOAD("od_ctr")
30 IF oneTxOnly > 1 THEN GOTO 270
40 IF odType != "sell" THEN GOTO 70
50 LET lockAmount = ASSETVALUE(HEXDECODE(assetId))
60 IF odType != "buy" THEN GOTO 90
70 LET lockAmount = ASSETVALUE(HEXDECODE(tradeAssetId))
80 IF lockAmount == 0 THEN GOTO 270
90 IF EXISTS("fee_" + tradeAssetId) == 0 THEN GOTO 270
100 beginStore()
110 storeUint64(odKey(odId, "lockAmountSent"), 0)
120 storeUint64(odKey(odId, "lockAmount"), lockAmount)
130 storeString(odKey(odId, "assetId"), assetId)
140 storeString(odKey(odId, "type"), odType)
150 storeString(odKey(odId, "tradeAssetId"), tradeAssetId)
160 storeUint64(odKey(odId, "unitPrice"), unitPrice)
170 storeString(odKey(odId, "creator"), ADDRESS_STRING(SIGNER()))
180 storeUint64(odKey(odId, "close"), 0)
190 storeUint64(odKey(odId, "timestamp"), BLOCK_TIMESTAMP())
200 storeUint64(odKey(odId, "expireTimestamp"), expireTimestamp)
210 storeUint64(odKey(odId, "oneTxOnly"), oneTxOnly)
220 storeUint64(odKey(odId, "txCount"), 0)
230 storeString(odKey(odId, "txId"), HEX(TXID()))
240 endStore()
250 STORE("od_ctr", odId + 1)
260 RETURN 0
270 RETURN 1
End Function

Function CloseOrder(odId Uint64) Uint64
5 DIM retrieveFunds as Uint64
10 DIM signer, odType as String
20 LET signer = SIGNER()
30 LET odType = LOAD(odKey(odId, "type"))
40 IF ADDRESS_STRING(signer) != LOAD(odKey(odId, "creator")) THEN GOTO 150
50 IF LOAD(odKey(odId, "close")) == 1 THEN GOTO 150
60 beginStore()
70 LET retrieveFunds = LOAD(odKey(odId, "lockAmount")) - LOAD(odKey(odId, "lockAmountSent"))
80 IF odType != "sell" THEN GOTO 100
90 SEND_ASSET_TO_ADDRESS(signer, retrieveFunds, HEXDECODE(LOAD(odKey(odId, "assetId"))))
100 IF odType != "buy" THEN GOTO 120
110 SEND_ASSET_TO_ADDRESS(signer, retrieveFunds, HEXDECODE(LOAD(odKey(odId, "tradeAssetId"))))
120 storeUint64(odKey(odId, "close"), 1)
130 endStore()
140 RETURN 0
150 RETURN 1
End Function

Function Transact(odId Uint64) Uint64
10 DIM assetId, tradeAssetId, creator, signer, odType as String
20 DIM ctr, fee, unitPrice, txAmountSent, txAmountReceived, tokenCut, timestamp, lockAmount, lockAmountBalance, expireTimestamp, lockAmountSent as Uint64
30 IF LOAD(odKey(odId, "close")) != 0 THEN GOTO 570
40 LET timestamp = BLOCK_TIMESTAMP()
50 LET expireTimestamp = LOAD(odKey(odId, "expireTimestamp"))
60 IF expireTimestamp == 0 THEN GOTO 80
70 IF timestamp > expireTimestamp THEN GOTO 570
80 LET assetId = LOAD(odKey(odId, "assetId"))
90 LET lockAmount = LOAD(odKey(odId, "lockAmount"))
100 LET tradeAssetId = LOAD(odKey(odId, "tradeAssetId"))
110 LET odType = LOAD(odKey(odId, "type"))
120 LET lockAmountSent = LOAD(odKey(odId, "lockAmountSent"))
130 LET unitPrice = LOAD(odKey(odId, "unitPrice"))
140 LET lockAmountBalance = lockAmount - lockAmountSent
150 LET creator = LOAD(odKey(odId, "creator"))
160 LET signer = SIGNER()
170 LET tokenCut = 0
180 LET txAmountSent = 0
190 LET txAmountReceived = 0
195 LET fee = 0
200 beginStore()
205 IF EXISTS("fee_" + tradeAssetId) == 0 THEN GOTO 220
210 LET fee = LOAD("fee_" + tradeAssetId) 
220 IF odType != "sell" THEN GOTO 320 // transact is buying to sell order
230 LET txAmountSent = ASSETVALUE(HEXDECODE(tradeAssetId))
240 LET txAmountReceived = txAmountSent / unitPrice
250 IF LOAD(odKey(odId, "oneTxOnly")) == 0 THEN GOTO 270
260 IF txAmountReceived != lockAmount THEN GOTO 570
270 IF txAmountReceived == 0 THEN GOTO 570
280 IF txAmountReceived > lockAmountBalance THEN GOTO 570
290 LET tokenCut = txAmountSent * fee / 100
300 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(creator), txAmountSent - tokenCut, HEXDECODE(tradeAssetId))
310 SEND_ASSET_TO_ADDRESS(signer, txAmountReceived, HEXDECODE(assetId))
320 IF odType != "buy" THEN GOTO 420 // transact is selling to buy order
330 LET txAmountSent = ASSETVALUE(HEXDECODE(assetId))
340 LET txAmountReceived = txAmountSent * unitPrice
350 IF LOAD(odKey(odId, "oneTxOnly")) == 0 THEN GOTO 370
360 IF txAmountReceived != lockAmount THEN GOTO 570
370 IF txAmountReceived == 0 THEN GOTO 570
380 IF txAmountReceived > lockAmountBalance THEN GOTO 570
390 LET tokenCut = txAmountReceived * fee / 100
400 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(creator), txAmountSent, HEXDECODE(assetId))
410 SEND_ASSET_TO_ADDRESS(signer, txAmountReceived - tokenCut, HEXDECODE(tradeAssetId))
420 LET lockAmountSent = lockAmountSent + txAmountReceived
430 storeUint64(odKey(odId, "lockAmountSent"), lockAmountSent)
440 IF lockAmountSent < lockAmount THEN GOTO 460
450 storeUint64(odKey(odId, "close"), 1)
460 SEND_ASSET_TO_ADDRESS(LOAD("owner"), tokenCut, HEXDECODE(tradeAssetId))
470 LET ctr = LOAD(odKey(odId, "txCount"))
480 storeString(odKey(odId, "tx_" + ctr + "_sender"), ADDRESS_STRING(signer))
490 storeUint64(odKey(odId, "tx_" + ctr + "_amountSent"), txAmountSent)
500 storeUint64(odKey(odId, "tx_" + ctr + "_amountReceived"), txAmountReceived)
510 storeUint64(odKey(odId, "tx_" + ctr + "_timestamp"), timestamp)
520 storeString(odKey(odId, "tx_" + ctr + "_txId"), HEX(TXID()))
530 storeUint64(odKey(odId, "tx_" + ctr + "_fee"), tokenCut)
540 storeUint64(odKey(odId, "txCount"), ctr + 1)
550 endStore()
560 RETURN 0
570 RETURN 1
End Function

Function SetAssetFee(assetId String, fee Uint64) Uint64
10 IF LOAD("owner") != SIGNER() THEN GOTO 50
20 IF fee > 100 THEN GOTO 50
30 STORE("fee_" + assetId, fee)
40 RETURN 0
50 RETURN 1
End Function

Function DelAssetFee(assetId String) Uint64
10 IF LOAD("owner") != SIGNER() THEN GOTO 40
20 DELETE("fee_" + assetId)
30 RETURN 0
40 RETURN 1
End Function

Function TransferOwnership(newMinter string) Uint64
10 IF LOAD("owner") != SIGNER() THEN GOTO 40
20 STORE("tempOwner", ADDRESS_RAW(newMinter))
30 RETURN 0
40 RETURN 1
End Function

Function CancelTransferOwnership() Uint64
10 IF LOAD("owner") != SIGNER() THEN GOTO 40
20 DELETE("tempOwner")
30 RETURN 0
40 RETURN 1
End Function

Function ClaimOwnership() Uint64
10 IF LOAD("tempOwner") != SIGNER() THEN GOTO 50
20 STORE("owner", SIGNER())
30 DELETE("tempOwner")
40 RETURN 0
50 RETURN 1
End Function

Function UpdateCode(code String) Uint64
10 IF LOAD("owner") != SIGNER() THEN GOTO 40
20 UPDATE_SC_CODE(code)
30 RETURN 0
40 RETURN 1
End Function
