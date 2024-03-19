package response

import (
	"github.com/filecoin-project/lotus/chain/types"
)

type MinerInfo struct {
	Balance          types.BigInt `json:"balance"`                          // Total balance of nodes
	Available        types.BigInt `json:"available"`                        // Node available balance
	BorrowableAmount types.BigInt `json:"borrowableAmount"`                 // Node available balance
	Vesting          string       `json:"vesting"`                          // Lock in rewards
	Pledge           string       `json:"pledge"`                           // Sector pledge
	Power            string       `json:"power"`                            // Effective computing power
	SectorSize       string       `json:"sectorSize"`                       // Sector Size
	Active           uint64       `json:"active"`                           // Effective sector
	Faulty           uint64       `json:"faulty"`                           // Wrong sector
	Live             uint64       `json:"live"`                             // Total sector
	Owner            string       `json:"owner"`                            // Owner Wallet
	OldOwner         string       `json:"oldOwner"`                         // Owner Wallet - Old
	Beneficiary      string       `json:"beneficiary"`                      // Beneficiary wallet
	OwnerId          string       `json:"ownerId"`                          // Owner ID
	BeneficiaryId    string       `json:"beneficiaryId"`                    // Beneficiary ID
	MortgageType     int          `json:"mortgageType" form:"mortgageType"` // 1 owner, 2 beneficiaries
}
