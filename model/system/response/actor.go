package response

type ActorInfo struct {
	Actor    string `json:"actor" form:"actor"`
	Operator string `json:"operator" form:"operator"`
	Owner    string `json:"owner" form:"owner"`
	Nonce    uint64 `json:"nonce" form:"nonce"`
}

type PondActorInfo struct {
	ActorId      uint64 `json:"actorId" form:"actorId"`
	Operator     string `json:"operator" form:"operator"`
	OwnerId      uint64 `json:"ownerId" form:"ownerId"`
	MortgageType int    `json:"mortgageType" form:"mortgageType"` // 1 owner, 2 beneficiaries
}

type PondActorStatus struct {
	Actor       string `json:"actor" form:"actor"`
	Owner       string `json:"owner" form:"owner"`
	OldOwner    string `json:"oldOwner" form:"oldOwner"`
	OldOwnerId  string `json:"oldOwnerId" form:"oldOwnerId"`
	Status      uint64 `json:"status" form:"status"`           // 0 is inactive, 1 is active, 2 has resigned but the owner has not been changed, 3 has resigned but the beneficiary has not been changed
	Beneficiary string `json:"beneficiary" form:"beneficiary"` // Beneficiary ID
	Debt        string `json:"debt" form:"debt"`
}
