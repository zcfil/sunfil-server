package system

import "zcfil-server/global"

type SysContractAbi struct {
	global.ZC_MODEL
	AbiId      string `gorm:"comment:Abi ID" json:"abiId"`
	AbiName    string `gorm:"comment:Abi name" json:"abiName"`
	ActorID    string `gorm:"comment:Contract MinerId" json:"actorID"`
	Wallet     string `gorm:"comment:Sending a message wallet" json:"wallet"`
	AbiContent string `gorm:"type:text;comment:Abi Content" json:"abiContent"`
	Param      []byte `gorm:"type:text;comment:parameter" json:"param"`
}

func (SysContractAbi) TableName() string {
	return "sys_contract_abi"
}
