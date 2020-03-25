package action

import(
	"log"
	"fmt"
	"encoding/json"
	sdk "CocosSDK"
)

func SubmmitVotingRecords(contractName, voteID, filterStartDate, filterEndDate string) {
	order := fmt.Sprintf("timestamp asc")
	cond := fmt.Sprintf("votee_id=\"%v\" and timestamp>=\"%v\" and timestamp<=\"%v\"", voteID, filterStartDate, filterEndDate)
	sql := fmt.Sprintf("select block_num, trx_id, op_account as voter, votee_id as votee, vote_amount as amount, timestamp from vote_ops where %v order by %v", cond, order)
	
	fmt.Printf("sql: %#v\n", sql)
	r, err := dbConn.Execute(sql)
	if err != nil {
		log.Fatalf("DB query error: %v", err.Error())
	}

	// Submmit to smart contract
	for i := 0; i < r.RowNumber(); i++ {
		blockNum, _ := r.GetIntByName(i, "block_num")
		trxID, _ := r.GetStringByName(i, "trx_id")
		voter, _ := r.GetStringByName(i, "voter")
		votee, _ := r.GetStringByName(i, "votee")
		amount, _ := r.GetIntByName(i, "amount") 
		timestamp, _ := r.GetStringByName(i, "timestamp")

		opItem := map[string]interface{}{
			"block_num": blockNum,
			"trx_id": trxID,
			"voter": voter,
			"votee": votee,
			"amount": amount,
			"timestamp": timestamp,
		}
		opJSON, _ := json.Marshal(opItem)
		fmt.Println(string(opJSON))

		result, err := sdk.InvokeContract(contractName, "add_vote_op_item", string(opJSON))
		if err != nil {
			log.Printf("WARNNING - some error happened when submmitted: %v", err.Error())
		} else {
			log.Printf("Submmitting result: %#v", result)
		}
	}
}
