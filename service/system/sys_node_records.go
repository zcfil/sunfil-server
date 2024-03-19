package system

import (
	"log"
	"zcfil-server/define"
	"zcfil-server/global"
	"zcfil-server/model/system"
	sysReq "zcfil-server/model/system/request"
	"zcfil-server/model/system/response"
)

type NodeRecordsService struct{}

func (n *NodeRecordsService) GetMultisigRecords() (map[string][]system.SysNodeRecords, error) {
	var toRecords = make(map[string][]system.SysNodeRecords)
	var records []system.SysNodeRecords
	if err := global.ZC_DB.Where("is_multisig = ? and applied = ?", true, false).Find(&records).Error; err != nil {
		return nil, err
	}
	for _, record := range records {
		toRecords[record.ToAddr[1:]] = append(toRecords[record.ToAddr[1:]], record)
	}

	return toRecords, nil
}

func (n *NodeRecordsService) GetSysRecords() (records []system.SysNodeRecords, err error) {
	return records, global.ZC_DB.Where("is_multisig = ? and applied = ?", true, false).Find(&records).Error
}

func (n *NodeRecordsService) CreateSysRecords(records system.SysNodeRecords) (err error) {
	var node system.SysNodeRecords
	global.ZC_DB.Model(system.SysNodeRecords{}).Where("cid = ?", records.Cid).Find(&node)
	if node.ID > 0 {
		records.ID = node.ID
		records.CreatedAt = node.CreatedAt
		records.UpdatedAt = node.UpdatedAt
		return global.ZC_DB.Updates(&records).Error
	}
	log.Println("Write operation record：", records)
	return global.ZC_DB.Create(&records).Error
}

func (n *NodeRecordsService) UpdateSysRecords(sysNodeInfo system.SysNodeRecords) (err error) {
	return global.ZC_DB.Updates(&sysNodeInfo).Error
}

func (n *NodeRecordsService) UpdateHeight(height int64) (err error) {
	sql := `UPDATE sys_update_height SET height = ?,updated_at = now()`

	return global.ZC_DB.Exec(sql, height).Error
}

func (n *NodeRecordsService) GetHeight() (int64, error) {
	var height system.SysUpdateHeight
	return height.Height, global.ZC_DB.First(&height).Error
}

func (n *NodeRecordsService) InitHeight() (err error) {
	if height, _ := n.GetHeight(); height == 0 {
		return global.ZC_DB.Create(&system.SysUpdateHeight{Height: 1}).Error
	}
	return nil
}

// GetNodeRecordList Get node record list
func (n *NodeRecordsService) GetNodeRecordList(info sysReq.GetNodeRecordReq) (list []response.NodeRecordResp, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.ZC_DB.Model(&system.SysNodeRecords{}).Where("applied = ?", 1).Group("cid")
	if info.NodeId != "" {
		db = db.Where("actor_id = ?", info.NodeId)
	}
	switch info.OpType {
	case define.NodeRecordBorrowStr:
		db = db.Where("op_type = ?", define.NodeRecordBorrow)
	case define.NodeRecordRepaymentStr:
		db = db.Where("op_type = ?", define.NodeRecordRepayment)
	case define.NodeRecordWithdrawStr:
		db = db.Where("op_type = ?", define.NodeRecordWithdraw)
	case define.NodeRecordChangeStr:
		db = db.Where("op_type in (?)", define.NodeRecordChange)
	case define.NodeRecordLiquidationStr:
		db = db.Where("op_type in (?)", define.NodeRecordLiquidation)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, total, err
	}
	err = db.Limit(limit).Offset(offset).Order("id desc").Find(&list).Error
	return list, total, err
}
