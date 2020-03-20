package action

import (
	"log"
	"time"
	"sync"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	sdk "CocosSDK"
	"github.com/siddontang/go-mysql/client"
	"github.com/wanliqun/bcx-witnode-vote-award/lib/util"
)

func FetchBlocks(startBlock, endBlock int64) {
	defer Close()

	// fetch blocks and write to db
	syncHeight := startBlock
	blockHeight := int64(0)
	for {
		if syncHeight >= endBlock {
			log.Printf("Block number has been synced to %v. Work is done!", syncHeight)
			//syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			os.Exit(0)
		}

		if blockHeight == 0 || syncHeight >= blockHeight {
			chainInfo := sdk.GetChainInfo()
			 if chainInfo == nil {
				time.Sleep(1 * time.Second)
				continue
			 }
			 blockHeight = int64(chainInfo.LastIrreversibleBlockNum)
		}

		if syncHeight >= blockHeight {
			time.Sleep(5 * time.Second)
			continue
		}

		numDiffBlocks := util.MinInt64(blockHeight - syncHeight, 100)
		numWorkers := 1
		if numDiffBlocks >= 20 {
			numWorkers = 4
		}

		log.Printf("fetching cocosbcx blocks with syncHeight - %v; blockHeight - %v; numWorkers - %v", syncHeight, blockHeight, numWorkers)
		var wg sync.WaitGroup
		errChans := make(chan error, numWorkers)

		for i := 1; i <= numWorkers; i++ {
			mod := 0
			avg := int(numDiffBlocks) / numWorkers

			if i == numWorkers {
				mod = int(numDiffBlocks) % numWorkers 
			}

			start := syncHeight + int64(1 + (i - 1) * avg)
			end :=  start + int64((avg - 1) + mod)
			dbCon := dbConnPools[i-1]

			wg.Add(1)
			go worker(i, &wg, dbCon, errChans, start, end)
		}

		wg.Wait()
		close(errChans)

		if len(errChans) > 0 {
			for err := range errChans {
				log.Println(err.Error())
			}
		} else {
			syncHeight += numDiffBlocks
		}
	}
}

func worker (id int, wg *sync.WaitGroup, dbConn *client.Conn, errChans chan error, block_start, block_end int64) {
	defer wg.Done()

	log.Printf("Worker %d started to fetch block height start from %d to %d", id, block_start, block_end)
	logPrefix := fmt.Sprintf("Inside worker(id-%d, block_start-%d, block_end-%d)", id, block_start, block_end)

	blockOpsData := map[int64][]map[string]interface{}{}
	for i := block_start; i <= block_end; i++ {
		block := sdk.GetBlock(i)
		
		if len(block.Transactions) > 0 {
			opsData := []map[string]interface{}{}

			for _, trx := range block.Transactions {
				trxId := trx[0]
				trxInfo := trx[1].(map[string]interface{})
				
				switch trxOps := trxInfo["operations"].(type) {
				case []interface{}:
					for _, ops := range trxOps {
						trxOpData := map[string]interface{}(nil)
						
						switch opv := ops.(type) {
						case []interface{}:
							opType := opv[0].(float64)
							if int(opType) != 6 {
								log.Printf("%v: skip transaction operation with trx type - %v for trxid - %v", logPrefix, opType, trxId)
								continue
							}

							opInfo := opv[1].(map[string]interface {})
							opAccount := opInfo["account"].(string)
							opNewOptions := opInfo["new_options"].(map[string]interface{})
							
							opLock := opInfo["lock_with_vote"]
							optionVotes := opNewOptions["votes"]
							
							switch opLockWithVote := opLock.(type) {
							case []interface{}:
								voteType := opLockWithVote[0].(float64)
								if int(voteType) != 1 {
									log.Printf("%v warning: vote with type - %v is not a witness vote for transaction with trxid - %v", logPrefix, voteType, trxId)
									continue
								}

								voteInfo := opLockWithVote[1].(map[string]interface{})
								voteAmount := voteInfo["amount"].(float64)
								voteAssetId := voteInfo["asset_id"].(string)

								trxOpData = map[string]interface{}{
									"op_type": opType,
									"op_account": opAccount,
									"vote_type": voteType,
									"vote_amount": voteAmount,
									"vote_asset_id": voteAssetId,
								}
							default:
								log.Printf("%v warning: unknown op lock_with_vote for transaction with trxid - %v", logPrefix, trxId)
								continue
							}

							switch newOptionsVotes := optionVotes.(type) {
							case []interface{}:
								votes := []string{}
								for _, v := range newOptionsVotes {
									vote := v.(string)
									votes = append(votes, vote)
								}
								trxOpData["votes"] = votes
							default:
								log.Printf("%v warning: unknown new option votes for transaction with trxid - %v", logPrefix, trxId)
								continue
							}
						default:
							log.Printf("%v warning: unknown transaction operation detail for trxid - %v", logPrefix, trxId)
							continue
						}

						if trxOpData != nil {
							trxOpData["block_num"] = i
							trxOpData["block_id"] = block.BlockID
							trxOpData["trx_id"] = trxId
							trxOpData["timestamp"] = block.Timestamp

							opsData = append(opsData, trxOpData)
						}
					}
				default:
					log.Printf("%v warning: unknown transaction operations for trxid - %v", logPrefix, trxId)
					continue
				}
			}

			if len(opsData) > 0 {
				blockOpsData[i] = opsData
			}
		}
	}

	if len(blockOpsData) == 0 {
		log.Printf("%v: no witness vote operations found", logPrefix)
		return
	}
	
	log.Printf("%v: saving block opsdata to db", logPrefix)
	logMsg, err := SaveBlockOpsDataToDB(dbConn, blockOpsData)

	if len(logMsg) > 0 {
		log.Printf("%v: %v", logPrefix, logMsg)
	}
	if err != nil {
		err = errors.New(fmt.Sprintf("%v error: %v", logPrefix, err.Error()))
		errChans <- err
	}
}

func SaveBlockOpsDataToDB(dbConn *client.Conn, blockOpsData map[int64][]map[string]interface{}) (string, error) {
	// get all map keys
	keys := []string{}
	for bn, _ := range blockOpsData {
		keys = append(keys, strconv.FormatInt(bn,10))
	}

	fmt.Printf("keys: %#v\n", keys)
	sql := fmt.Sprintf("select distinct block_num from vote_ops where block_num in (%v)", strings.Join(keys, ","))
	fmt.Printf("sql: %#v\n", sql)
	r, err := dbConn.Execute(sql)
	if err != nil {
		return "", err
	}

	if r.RowNumber() == len(keys) {
		return "skip now because all blocks are already saved in DB", nil
	}

	// remove duplicate blocks
	for i := 0; i < r.RowNumber(); i++ {
		bn, _ := r.GetStringByName(i, "block_num")
		idx := util.SearchStringSlice(keys, bn)

		if idx >= 0 {
			keys = util.RemoveStringSliceAt(keys, idx)
		}
	}

	for _, key := range keys {
		blockNum, _ := strconv.ParseInt(key,10,64)
		opsd := blockOpsData[blockNum]

		bulkInsertSql := "INSERT INTO vote_ops(block_id,block_num,trx_id,op_account,op_type,vote_type,votee_id,vote_asset_id,vote_amount,timestamp) VALUES"
		for _, opData := range opsd {
			blockId := opData["block_id"]
			blockNum := opData["block_num"] 
			trxId := opData["trx_id"] 
			opAccount := opData["op_account"] 
			opType := opData["op_type"] 
			voteType := opData["vote_type"] 
			voteAssetId := opData["vote_asset_id"] 
			voteAmount := opData["vote_amount"] 
			timestamp := opData["timestamp"] 

			opVotes := opData["votes"].([]string)
			for j := 0; j < len(opVotes); j++ {
				voteeId := opVotes[j]
				seperator := ""
				if j < (len(opVotes) - 1) {
					seperator = ","
				}

				innerSql := fmt.Sprintf(`("%v",%v,"%v","%v",%v,%v,"%v","%v",%v,"%v")%v`, 
					blockId, blockNum, trxId, opAccount, opType, voteType, voteeId, voteAssetId, voteAmount, timestamp, seperator)
				bulkInsertSql += innerSql
			}
		}

		_, err := dbConn.Execute(bulkInsertSql)
		if err != nil {
			return "", err
		}

		fmt.Printf("insert result: %v\n", r)
		fmt.Printf("insert result: %v\n", r.Resultset)
		fmt.Printf("insert result: %#v\n", r.Resultset)
		if r.AffectedRows < 1 {
			return fmt.Sprintf("Mysql db unknown error (%v)", bulkInsertSql), errors.New("Mysql db insert unknown error")
		}
	}

	return fmt.Sprintf("blocks(%v) ops data saved done", strings.Join(keys, ",")), nil
}