/** COMMIT & STATE LIB CODE **/
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
20 storeCommitInt("S", "state_" + key, value) // S - store
30 RETURN
End Function

Function deleteState(key String)
10 DELETE("state_" + key)
20 storeCommitInt("D", "state_" + key, 0) // D - delete
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
10 storeStateInt("txid_" + HEX(TXID()), 1) // verify transaction within the smart contract 
20 RETURN
End Function

//** NAME CODE **//
Function nameKey(addr String) String
10 RETURN "name_" + addr
End Function

Function Register(name String) Uint64
10 DIM name_length as Uint64
20 DIM signer_string, signer_name_key as String
30 beginCommit()
40 LET name_length = STRLEN(name)
50 IF name_length > 2 THEN GOTO 70
60 RETURN 1
70 IF name_length <= 100 THEN GOTO 90
80 RETURN 1
90 IF EXISTS(nameKey(name)) == 0 THEN GOTO 110
100 RETURN 1
110 LET signer_string = ADDRESS_STRING(SIGNER())
120 IF signer_string != "" THEN GOTO 140 // ring size 2 only
130 RETURN 1
140 LET signer_name_key = nameKey(signer_string)
150 IF stateExists(signer_name_key) == 0 THEN GOTO 170
160 DELETE(nameKey(loadStateString(signer_name_key))) // delete previous name for somebody else to register
170 STORE(nameKey(name), signer_string)
180 storeStateString(signer_name_key, name)
190 storeTX()
200 endCommit()
210 RETURN 0
End Function

Function Unregister() Uint64
10 DIM signer_name_key as String
20 beginCommit()
30 LET signer_name_key = nameKey(ADDRESS_STRING(SIGNER()))
40 IF stateExists(signer_name_key) == 1 THEN GOTO 60
50 RETURN 1
60 DELETE(nameKey(loadStateString(signer_name_key)))
70 deleteState(signer_name_key)
80 storeTX()
90 endCommit()
100 RETURN 0
End Function

Function Initialize() Uint64
10 STORE("sc_owner", SIGNER())
20 initCommit()
30 RETURN 0
End Function

Function ClaimOwnership() Uint64
10 IF LOAD("sc_owner_temp") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("sc_owner", SIGNER())
40 RETURN 0
End Function

Function TransferOwnership(newOwner String) Uint64
10 IF LOAD("sc_owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("sc_owner_temp", ADDRESS_RAW(newOwner))
40 RETURN 0
End Function

Function UpdateCode(code String) Uint64
10 IF LOAD("sc_owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 UPDATE_SC_CODE(code)
40 RETURN 0
End Function