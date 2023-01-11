package eventstream

import (
	"encoding/json"
	"strconv"
	"time"
)

func getBlockHeight(msg []byte) (uint64, error) {
	t := new(newBlockJSON)
	err := json.Unmarshal(msg, t)
	if err != nil {
		return 0, err
	}
	if t.Result.Data.Value.Block.Header.Height == "" {
		return 0, nil
	}
	return strconv.ParseUint(t.Result.Data.Value.Block.Header.Height, 10, 64)
}

// todo mercilex split in concrete types instead of anonymous
type newBlockJSON struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Query string `json:"query"`
		Data  struct {
			Type  string `json:"type"`
			Value struct {
				Block struct {
					Header struct {
						ChainID        string    `json:"chain_id"`
						Height         string    `json:"height"`
						Time           time.Time `json:"time"`
						LastCommitHash string    `json:"last_commit_hash"`
					} `json:"header"`
					Data struct {
						Txs []interface{} `json:"txs"`
					} `json:"data"`
				} `json:"block"`
				ResultBeginBlock struct {
					Events []tmEvent `json:"events"`
				} `json:"result_begin_block"`
				ResultEndBlock struct {
					ValidatorUpdates []interface{} `json:"validator_updates"`
					Events           []tmEvent     `json:"events"`
				} `json:"result_end_block"`
			} `json:"value"`
		} `json:"data"`
	} `json:"result"`
}

type tmEvent struct {
	Type       string `json:"type"`
	Attributes []tmEventAttribute
}

type tmEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Index bool   `json:"index"`
}
