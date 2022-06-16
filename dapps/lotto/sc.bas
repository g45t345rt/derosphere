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

Function sss(key String, value String)
10 STORE("state_" + key, value)
20 storeCommitString("S", "state_" + key, value)
30 RETURN
End Function

Function ssi(key String, value Uint64)
10 STORE("state_" + key, value)
20 storeCommitInt("S", "state_" + key, value) // S - store
30 RETURN
End Function

Function ds(key String)
10 DELETE("state_" + key)
20 storeCommitInt("D", "state_" + key, 0) // D - delete
30 RETURN
End Function

Function lss(key String) String
10 RETURN LOAD("state_" + key)
End Function

Function lsi(key String) Uint64
10 RETURN LOAD("state_" + key)
End Function

Function se(key String) Uint64
10 RETURN EXISTS("state_" + key)
End Function

Function storeTX()
10 ssi("txid_" + HEX(TXID()), 1) // verify transaction within the smart contract 
20 RETURN
End Function

/** LOTTO CODE **/

Function lk(txId String, suffix String) String
10 RETURN "lotto_" + txId + "_" + suffix
End Function

Function ltk(txId String, ticketNumber Uint64, suffix String) String
10 RETURN lk(txId, "ticket_" + ticketNumber + "_" + suffix)
End Function

Function Play(txId String, userPasswordHash String) Uint64
10 DIM duration, timestamp, start_timestamp, unique_wallet, max_tickets, ticket_price, ticket_number, end_timestamp as Uint64
20 DIM password_hash, signer_string as String
30 beginCommit()
35 LET signer_string = ADDRESS_STRING(SIGNER())
40 IF se(lk(txId, "password_hash")) == 0 THEN GOTO 80
50 LET password_hash = lss(lk(txId, "password_hash"))
60 IF HEX(SHA3256(signer_string + "." + password_hash)) == userPasswordHash THEN GOTO 80
70 RETURN 1
80 LET timestamp = BLOCK_TIMESTAMP()
90 LET start_timestamp = lsi(lk(txId, "start_timestamp"))
95 IF timestamp >= start_timestamp THEN GOTO 100
98 RETURN 1
100 LET duration = lsi(lk(txId, "duration"))
105 LET end_timestamp = start_timestamp + duration
110 IF duration == 0 THEN GOTO 140
120 IF timestamp < end_timestamp THEN GOTO 140
130 RETURN 1
140 LET unique_wallet = lsi(lk(txId, "unique_wallet"))
150 LET ticket_number = lsi(lk(txId, "ticket_count"))
160 LET max_tickets = lsi(lk(txId, "max_tickets"))
170 IF max_tickets == 0 THEN GOTO 200
180 IF ticket_number < max_tickets THEN GOTO 200
190 RETURN 1
200 LET ticket_price = lsi(lk(txId, "ticket_price"))
210 IF DEROVALUE() == ticket_price THEN GOTO 230
220 RETURN 1
230 IF signer_string != "" THEN GOTO 240
235 RETURN 1
240 IF unique_wallet == 0 THEN GOTO 280
250 IF se(lk(txId, "unique_ticket_" + signer_string)) == 0 THEN GOTO 270
260 RETURN 1
270 ssi(lk(txId, "unique_ticket_" + signer_string), ticket_number) // this should be skip if unique wallet is 0
280 sss(ltk(txId, ticket_number, "owner"), signer_string)
290 ssi(ltk(txId, ticket_number, "timestamp"), timestamp)
300 sss(ltk(txId, ticket_number, "play_tx_id"), HEX(TXID()))
310 ssi(lk(txId, "ticket_count"), ticket_number + 1)
320 storeTX()
330 endCommit()
340 RETURN 0
End Function

Function Create(maxTickets Uint64, ticketPrice Uint64, duration Uint64, uniqueWallet Uint64, passwordHash String, startTimestamp Uint64) Uint64
10 DIM max_tickets, timestamp, anti_spam_fee, dero_value, base_reward as Uint64
20 DIM tx_id, signer_string as String
30 beginCommit()
40 LET timestamp = BLOCK_TIMESTAMP()
50 LET tx_id = HEX(TXID())
60 LET max_tickets = maxTickets
70 LET signer_string = ADDRESS_STRING(SIGNER())
75 IF signer_string != "" THEN GOTO 85
80 RETURN 1
85 LET dero_value = DEROVALUE()
90 IF ticketPrice <= 100000000 THEN GOTO 110
100 RETURN 1
110 IF uniqueWallet <= 1 THEN GOTO 130
120 RETURN 1
130 IF max_tickets != 1 THEN GOTO 150
140 RETURN 1
150 IF max_tickets >= 2 THEN GOTO 180
160 IF duration >= 60 THEN GOTO 190
170 RETURN 1
180 LET duration = 0
190 IF startTimestamp == 0 THEN GOTO 220
200 IF startTimestamp > timestamp THEN GOTO 221
210 RETURN 1
220 LET startTimestamp = timestamp
221 LET anti_spam_fee = LOAD("anti_spam_fee")
222 IF dero_value >= anti_spam_fee THEN GOTO 226
224 RETURN 1
226 LET base_reward = dero_value - anti_spam_fee
228 ssi(lk(tx_id, "anti_spam_fee"), anti_spam_fee)
230 ssi(lk(tx_id, "base_reward"), base_reward)
240 ssi(lk(tx_id, "max_tickets"), maxTickets)
250 ssi(lk(tx_id, "ticket_price"), ticketPrice)
260 ssi(lk(tx_id, "duration"), duration)
270 ssi(lk(tx_id, "unique_wallet"), uniqueWallet)
280 IF STRLEN(passwordHash) == 0 THEN GOTO 300
290 sss(lk(tx_id, "password_hash"), HEX(SHA3256(tx_id + "." + passwordHash)))
300 sss(lk(tx_id, "owner"), signer_string)
310 ssi(lk(tx_id, "start_timestamp"), startTimestamp)
320 ssi(lk(tx_id, "ticket_count"), 0)
330 storeTX()
340 endCommit()
350 RETURN 0
End Function

Function Draw(txId String) Uint64
10 DIM ticket_count, max_tickets, duration, winning_ticket, start_timestamp, end_timestamp, draw_timestamp, anti_spam_fee as Uint64
20 DIM winner as String
30 beginCommit()
40 IF se(lk(txId, "draw_timestamp")) == 0 THEN GOTO 60
50 RETURN 1
60 LET draw_timestamp = BLOCK_TIMESTAMP()
70 LET ticket_count = lsi(lk(txId, "ticket_count"))
80 LET max_tickets = lsi(lk(txId, "max_tickets"))
90 LET start_timestamp = lsi(lk(txId, "start_timestamp"))
100 LET duration = lsi(lk(txId, "duration"))
110 LET end_timestamp = start_timestamp + duration
120 IF max_tickets == 0 THEN GOTO 150
130 IF ticket_count == max_tickets THEN GOTO 150
140 RETURN 1
150 IF duration == 0 THEN GOTO 180
160 IF draw_timestamp > end_timestamp THEN GOTO 180
170 RETURN 1
180 IF ticket_count > 0 THEN GOTO 200
190 RETURN 1
200 LET winning_ticket = RANDOM() % ticket_count
210 LET winner = lss(ltk(txId, winning_ticket, "owner"))
220 ssi(lk(txId, "winning_ticket"), winning_ticket)
230 sss(lk(txId, "winner"), winner)
240 ssi(lk(txId, "draw_timestamp"), draw_timestamp)
250 LET anti_spam_fee = lsi(lk(txId, "anti_spam_fee"))
260 SEND_DERO_TO_ADDRESS(ADDRESS_RAW(lss(lk(txId, "owner"))), anti_spam_fee)
270 storeTX()
280 endCommit()
290 RETURN 0
End Function

Function Cancel(txId String) Uint64
10 DIM ticket_count, base_reward, anti_spam_fee as Uint64
20 DIM owner_raw as String
30 beginCommit()
40 LET owner_raw = ADDRESS_RAW(lss(lk(txId, "owner")))
50 IF owner_raw == SIGNER() THEN GOTO 70
60 RETURN 1
70 LET ticket_count = lsi(lk(txId, "ticket_count"))
80 IF ticket_count == 0 THEN GOTO 100
90 RETURN 1
100 LET base_reward = lsi(lk(txId, "base_reward"))
110 IF base_reward == 0 THEN GOTO 130
120 SEND_DERO_TO_ADDRESS(owner_raw, base_reward)
130 LET anti_spam_fee = lsi(lk(txId, "anti_spam_fee"))
140 SEND_DERO_TO_ADDRESS(owner_raw, anti_spam_fee)
150 ds(lk(txId, "max_tickets"))
160 ds(lk(txId, "ticket_price"))
170 ds(lk(txId, "duration"))
180 ds(lk(txId, "unique_wallet"))
190 ds(lk(txId, "password_hash"))
200 ds(lk(txId, "owner"))
210 ds(lk(txId, "ticket_count"))
220 ds(lk(txId, "base_reward"))
230 ds(lk(txId, "start_timestamp"))
240 ds(lk(txId, "anti_spam_fee"))
250 storeTX()
260 endCommit()
270 RETURN 0
End Function

Function ClaimReward(txId String, password String, comment String) Uint64
10 DIM ticket_price, ticket_count, reward, sc_cut, base_reward, comment_length as Uint64
20 DIM winner_string, password_hash, owner, winner_raw as String
30 beginCommit()
31 IF se(lk(txId, "password_hash")) == 0 THEN GOTO 40
32 LET password_hash = lss(lk(txId, "password_hash"))
33 LET ticket_price = lsi(lk(txId, "ticket_price"))
34 LET owner = lss(lk(txId, "owner"))
35 IF password_hash == HEX(SHA3256(txId + "." + HEX(SHA3256(owner + "." + ticket_price + "." + password)))) THEN GOTO 40
36 RETURN 1
40 IF se(lk(txId, "claim_tx_id")) == 0 THEN GOTO 60
50 RETURN 1
60 LET comment_length = STRLEN(comment)
70 IF comment_length <= 100 THEN GOTO 90
80 RETURN 1
90 LET winner_string = lss(lk(txId, "winner"))
100 LET winner_raw = ADDRESS_RAW(winner_string)
110 IF winner_raw == SIGNER() THEN GOTO 130
120 RETURN 1
130 LET base_reward = lsi(lk(txId, "base_reward"))
150 LET ticket_count = lsi(lk(txId, "ticket_count"))
160 LET reward = base_reward + (ticket_price * ticket_count)
170 LET sc_cut = reward * 10 / 100
180 SEND_DERO_TO_ADDRESS(LOAD("sc_owner"), sc_cut)
190 SEND_DERO_TO_ADDRESS(winner_raw, reward - sc_cut)
200 IF comment_length == 0 THEN GOTO 220
210 sss(lk(txId, "winner_comment"), comment)
220 ssi(lk(txId, "claim_timestamp"), BLOCK_TIMESTAMP())
230 sss(lk(txId, "claim_tx_id"), HEX(TXID()))
240 storeTX()
250 endCommit()
260 RETURN 0
End Function

/** SC OWNER CODE **/

Function Initialize() Uint64
10 STORE("sc_owner", SIGNER())
20 initCommit()
30 STORE("anti_spam_fee", 100000)
40 RETURN 0
End Function

Function SetAntiSpamFee(fee Uint64) Uint64
10 IF LOAD("sc_owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("anti_spam_fee", fee)
40 RETURN 0
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