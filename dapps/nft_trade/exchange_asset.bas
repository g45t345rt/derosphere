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

Function CreateExchange(sellAssetId String, buyAssetId String, buyAmount Uint64, expireTimestamp Uint64) Uint64
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
140 storeStateInt(exKey(exId, "expireTimestamp"), expireTimestamp)
150 endCommit()
160 STORE("ex_ctr", exId + 1)
170 RETURN 0
End Function

Function CancelExchange(exId Uint64) Uint64
10 DIM signer as String
20 LET signer = SIGNER()
30 IF LOAD("owner") == signer THEN GOTO 60
40 IF loadStateString(exKey(exId, "seller")) == ADDRESS_STRING(signer) THEN GOTO 60
50 RETURN 1
60 IF loadStateInt(exKey(exId, "complete")) == 0 THEN GOTO 80
70 RETURN 1
80 beginCommit()
90 SEND_ASSET_TO_ADDRESS(signer, loadStateInt(exKey(exId, "sellAmount")), HEXDECODE(loadStateString(exKey(exId, "sellAssetId"))))
100 deleteState(exKey(exId, "sellAmount"))
110 deleteState(exKey(exId, "sellAssetId"))
120 deleteState(exKey(exId, "buyAssetId"))
130 deleteState(exKey(exId, "buyAmount"))
140 deleteState(exKey(exId, "seller"))
150 deleteState(exKey(exId, "complete"))
160 deleteState(exKey(exId, "timestamp"))
170 deleteState(exKey(exId, "expireTimestamp"))
180 endCommit()
190 RETURN 0
End Function

Function Buy(exId Uint64) Uint64
10 DIM sellAssetId, buyAssetId, seller, signer as String
20 DIM sellAmount, buyAmount, deroCut, complete, timestamp, expireTimestamp as Uint64
30 IF loadStateInt(exKey(exId, "complete")) == 0 THEN GOTO 50
40 RETURN 1
50 LET timestamp = BLOCK_TIMESTAMP()
60 LET expireTimestamp = loadStateInt(exKey(exId, "expireTimestamp"))
70 IF expireTimestamp == 0 THEN GOTO 100
80 IF timestamp < expireTimestamp THEN GOTO 100
90 RETURN 1
100 LET sellAssetId = loadStateString(exKey(exId, "sellAssetId"))
110 LET sellAmount = loadStateInt(exKey(exId, "sellAmount"))
120 LET buyAssetId = loadStateString(exKey(exId, "buyAssetId"))
130 LET buyAmount = loadStateInt(exKey(exId, "buyAmount"))
140 LET seller = loadStateString(exKey(exId, "seller"))
150 LET signer = SIGNER()
160 LET deroCut = 0
170 IF ASSETVALUE(HEXDECODE(buyAssetId)) == buyAmount THEN GOTO 190
180 RETURN 1
190 IF sellAssetId != "0000000000000000000000000000000000000000000000000000000000000000" THEN GOTO 220
200 LET deroCut = sellAmount * LOAD("dero_fee") / 100
210 LET sellAmount = sellAmount - deroCut
220 IF buyAssetId != "0000000000000000000000000000000000000000000000000000000000000000" THEN GOTO 250
230 LET deroCut = buyAmount * LOAD("dero_fee") / 100
240 LET buyAmount = buyAmount - deroCut
250 SEND_ASSET_TO_ADDRESS(ADDRESS_RAW(seller), buyAmount, HEXDECODE(buyAssetId))
260 SEND_ASSET_TO_ADDRESS(signer, sellAmount, HEXDECODE(sellAssetId))
270 SEND_DERO_TO_ADDRESS(LOAD("owner"), deroCut)
280 beginCommit()
290 storeStateString(exKey(exId, "buyer"), ADDRESS_STRING(signer))
300 storeStateInt(exKey(exId, "complete"), 1)
310 storeStateInt(exKey(exId, "completeTimestamp"), timestamp)
320 endCommit()
330 RETURN 0
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
