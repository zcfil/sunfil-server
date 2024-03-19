package rpc

import (
	"errors"
	"zcfil-server/global"
	"zcfil-server/model/system"
)

type DbQueryService struct{}

func (l *DbQueryService) GetContractConfig(abiId string) (sysContractAbi system.SysContractAbi, err error) {

	var sca system.SysContractAbi
	if abiId == "" {
		return sca, errors.New("abiId is null")
	}
	err = global.ZC_DB.Model(&system.SysContractAbi{}).Where("abi_id", abiId).First(&sca).Error
	if err != nil {
		return sca, err
	}
	return sca, nil
}
