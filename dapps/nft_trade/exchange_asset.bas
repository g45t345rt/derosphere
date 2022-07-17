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

Function deleteState(key String)
10 DELETE("state_" + key)
20 storeCommitInt("D", "state_" + key, 0)
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

Function storeTX()
10 STORE("txid_" + HEX(TXID()), 1)
20 RETURN
End Function

Function exKey(id Uint64, key String) String
10 RETURN "ex_" + id + "_" + key
End Function

Function Initialize() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", SIGNER())
40 STORE("dero_fee", 5)
50 STORE("ex_ctr", 0)
60 initCommit()
70 RETURN 0
End Function

Function CreateExchange(sellAssetId String, buyAssetId String, buyAmount Uint64) Uint64
10 DIM exId, sellAmount as Uint64
20 LET exId = LOAD("ex_ctr")
30 LET sellAmount = ASSETVALUE(HEXDECODE(sellAssetId))
40 IF sellAmount > 0 THEN GOTO 60
50 RETURN 1
60 beginCommit()
70 storeStateInt(exKey(exId, "sellAmount"), sellAmount)
80 storeStateString(exKey(exId, "sellAssetId"), sellAssetId)
90 storeStateString(exKey(exId, "buyAssetId"), buyAssetId)
100 storeStateInt(exKey(exId, "buyAmount"), buyAmount)
110 storeStateString(exKey(exId, "seller"), ADDRESS_STRING(SIGNER()))
120 storeStateInt(exKey(exId, "complete"), 0)
130 storeStateInt(exKey(exId, "timestamp"), BLOCK_TIMESTAMP())
140 endCommit()
150 STORE("ex_ctr", exId + 1)
160 storeTX()
170 RETURN 0
End Function

Function CancelExchange(exId Uint64) Uint64
10 DIM signer as String
20 LET signer = SIGNER()
30 IF loadStateString(exKey(exId, "seller")) == ADDRESS_STRING(signer) THEN GOTO 50
40 RETURN 1
50 IF loadStateInt(exKey(exId, "complete")) == 0 THEN GOTO 70
60 RETURN 1
70 beginCommit()
80 SEND_ASSET_TO_ADDRESS(signer, loadStateInt(exKey(exId, "sellAmount")), HEXDECODE(loadStateString(exKey(exId, "sellAssetId"))))
90 deleteState(exKey(exId, "sellAmount"))
100 deleteState(exKey(exId, "sellAssetId"))
110 deleteState(exKey(exId, "buyAssetId"))
120 deleteState(exKey(exId, "buyAmount"))
130 deleteState(exKey(exId, "seller"))
140 deleteState(exKey(exId, "complete"))
150 deleteState(exKey(exId, "timestamp"))
160 endCommit()
170 storeTX()
180 RETURN 0
End Function

Function Buy(exId Uint64) Uint64
10 DIM sellAssetId, buyAssetId, seller, signer as String
20 DIM sellAmount, buyAmount, deroCut, complete as Uint64
30 IF loadStateInt(exKey(exId, "complete")) == 0 THEN GOTO 50
40 RETURN 1
50 LET sellAssetId = loadStateString(exKey(exId, "sellAssetId"))
60 LET sellAmount = loadStateInt(exKey(exId, "sellAmount"))
70 LET buyAssetId = loadStateString(exKey(exId, "buyAssetId"))
80 LET buyAmount = loadStateInt(exKey(exId, "buyAmount"))
90 LET seller = loadStateString(exKey(exId, "seller"))
95 LET signer = SIGNER()
100 LET deroCut = 0
110 IF ASSETVALUE(HEXDECODE(buyAssetId)) == buyAmount THEN GOTO 130
120 RETURN 1
130 IF sellAssetId != "0000000000000000000000000000000000000000000000000000000000000000" THEN GOTO 160
140 LET deroCut = sellAmount * LOAD("dero_fee") / 100
150 LET sellAmount = sellAmount - deroCut
160 IF buyAssetId != "0000000000000000000000000000000000000000000000000000000000000000" THEN GOTO 190
170 LET deroCut = buyAmount * LOAD("dero_fee") / 100
180 LET buyAmount = buyAmount - deroCut
190 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(seller), buyAmount, HEXDECODE(buyAssetId))
200 SEND_ASSET_TO_ADDRESS(signer, sellAmount, HEXDECODE(sellAssetId))
210 SEND_DERO_TO_ADDRESS(LOAD("owner"), deroCut)
220 beginCommit()
230 storeStateString(exKey(exId, "buyer"), ADDRESS_STRING(signer))
240 storeStateInt(exKey(exId, "complete"), 1)
250 storeStateInt(exKey(exId, "completeTimestamp"), BLOCK_TIMESTAMP())
260 endCommit()
270 storeTX()
280 RETURN 0
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
