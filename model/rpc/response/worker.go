package response

import "github.com/filecoin-project/go-address"

type WorkerInfo struct {
	Worker            string            `json:"worker"`            // Worker Wallet
	ControlAddresses  []address.Address `json:"control"`           // Beneficiary wallet
	NewWorker         string            `json:"newWorker"`         // New Worker Wallet
	WorkerChangeEpoch int64             `json:"workerChangeEpoch"` // Effective height of new worker wallet
	Owner             string            `json:"owner"`             // Owner Wallet
}
