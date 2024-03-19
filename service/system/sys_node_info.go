package system

import (
	"strconv"
	"time"
	"zcfil-server/contract"
	"zcfil-server/global"
	"zcfil-server/model/common/request"
	"zcfil-server/model/system"
	sysRequest "zcfil-server/model/system/request"
	"zcfil-server/model/system/response"
)

//@function: CreateSysNodeInfo
//@description: Create node data
//@param: sysNodeInfo model.SysNodeInfo
//@return: err error

type NodeInfoService struct{}

func (n *NodeInfoService) CreateSysNodeInfo(sysNodeInfo system.SysNodeInfo) (err error) {
	var count int64
	if err = global.ZC_DB.Model(system.SysNodeInfo{}).Where("node_name = ?", sysNodeInfo.NodeName).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		sysNodeInfo.UpdatedAt = time.Now()
		return global.ZC_DB.Model(system.SysNodeInfo{}).Where("node_name = ?", sysNodeInfo.NodeName).Updates(&sysNodeInfo).Error
	}

	return global.ZC_DB.Model(system.SysNodeInfo{}).Create(&sysNodeInfo).Error
}

func (n *NodeInfoService) UpdateSysNodeApplied(actorId uint64, applied bool) (err error) {
	return global.ZC_DB.Model(system.SysNodeInfo{}).Where("node_name = ?", actorId).Update("applied", applied).Error
}

func (n *NodeInfoService) UpdateSysNodeStatus(actorId uint64, stat int) (err error) {
	return global.ZC_DB.Model(system.SysNodeInfo{}).Where("node_name = ?", actorId).Update("status", stat).Error
}

func (n *NodeInfoService) UpdateSysNodeOperator(actorId uint64, operator string) (err error) {
	return global.ZC_DB.Model(system.SysNodeInfo{}).Where("node_name = ?", actorId).Update("operator", operator).Error
}

func (n *NodeInfoService) UpdateSysNodeOperatorAndStatus(actorId uint64, nodeInfo system.SysNodeInfo) (err error) {
	return global.ZC_DB.Model(system.SysNodeInfo{}).Where("node_name = ?", actorId).Update("operator", nodeInfo.Operator).
		Update("status", nodeInfo.Status).Update("max_debt_rate", nodeInfo.MaxDebtRate).
		Update("liquidate_rate", nodeInfo.LiquidateRate).Error
}

//@function: GetSysNodeInfo
//@description: Obtain single node data based on ActorId
//@param: id int
//@return: sysNodeInfo system.SysNodeInfo, err error

func (n *NodeInfoService) GetSysNodeInfo(actorId uint64) (sysNodeInfo system.SysNodeInfo, err error) {
	err = global.ZC_DB.Where("node_name = ?", actorId).First(&sysNodeInfo).Error
	return
}

//@function: DeleteSysNodeInfoByRoomId
//@description: Batch deletion of records
//@param: ids request.IdsReq
//@return: err error

func (n *NodeInfoService) DeleteSysNodeInfoByIds(ids request.IdsReq) (err error) {
	err = global.ZC_DB.Delete(&[]system.SysNodeInfo{}, "id in (?)", ids.Ids).Unscoped().Error
	return err
}

// @function: GetList
// @description: Paging to obtain node list
// @param: info request.PageInfo
// @return: list interface{}, total int64, err error

func (n *NodeInfoService) GetSysNodeInfoList(info sysRequest.NodeListReq) (list []response.NodeListResp, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.ZC_DB.Model(&system.SysNodeInfo{}).Where("applied = ?", 1)

	nodeName := info.NodeId
	if info.NodeId != "" {
		if len(info.NodeId) > 2 {
			if _, err := strconv.ParseFloat(info.NodeId, 64); err != nil {
				nodeName = info.NodeId[2:]
			}
		}
		db = db.Where("node_name LIKE ?", "%"+nodeName+"%")
	}

	if info.Status == contract.NodeStatusJoining {
		db = db.Where("status IN ?", []int{contract.JobStatusBeOn})
	} else {
		db = db.Where("status NOT IN ?", []int{contract.JobStatusBeOn})
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, total, err
	}

	var status string
	if info.Status == contract.NodeStatusJoining {
		status = "1,3"
	} else {
		status = "0,2,3"
	}

	var sql string
	if info.Status == contract.NodeStatusJoining {
		sql = "select * from sys_node_info WHERE operator != '" + contract.EmptyAddress + "' AND applied = 1 AND node_name LIKE '" + "%" + nodeName + "%" + "' AND status IN (" + status + ")  ORDER BY FIELD(operator,'" +
			info.Operator + "') DESC,id desc limit " + strconv.Itoa(offset) + "," + strconv.Itoa(limit)
	} else {
		sql = "select * from sys_node_info WHERE operator = '" + contract.EmptyAddress + "' AND applied = 1 AND node_name LIKE '" + "%" + nodeName + "%" + "' AND status IN (" + status + ")  ORDER BY FIELD(operator,'" +
			info.Operator + "') DESC,id desc limit " + strconv.Itoa(offset) + "," + strconv.Itoa(limit)
	}

	err = db.Raw(sql).Limit(limit).Offset(offset).Order("id desc").Scan(&list).Error
	return list, total, err
}

// @function: GetList
// @description: Get a list of all nodes
// @param: info request.PageInfo
// @return: list interface{}, total int64, err error

func (n *NodeInfoService) GetNodeList() (list []system.SysNodeInfo, err error) {
	db := global.ZC_DB.Model(&system.SysNodeInfo{})
	var entities []system.SysNodeInfo
	err = db.Order("id desc").Find(&entities).Error
	return entities, err
}

//@function: UpdateNodeInfo
//@description: Update node information
//@param: nodeInfo *model.NodeInfo
//@return: err error

func (n *NodeInfoService) UpdateNodeInfo(nodeInfo *system.SysNodeInfo) (err error) {
	var dict system.SysNodeInfo
	nodeInfoMap := map[string]interface{}{
		"DebtBalance": nodeInfo.DebtBalance,
		"Balance":     nodeInfo.Balance,
	}
	db := global.ZC_DB.Where("id = ?", nodeInfo.ID).First(&dict)
	err = db.Updates(nodeInfoMap).Error
	return err
}

// GetNodeNum Obtain the number of nodes
func (n *NodeInfoService) GetNodeNum(applied string) (total int64, err error) {
	db := global.ZC_DB.Model(&system.SysNodeInfo{})
	if applied != "" {
		db = db.Where("applied = ?", applied)
	}
	err = db.Count(&total).Error
	return
}

// GetNodeBalanceSum Obtain the total value of nodes
func (n *NodeInfoService) GetNodeBalanceSum() (float64, error) {
	sql := `select sum(balance) balanceNum from sys_node_info where applied = 1 AND operator != '` +
		contract.EmptyAddress + `' AND status IN (1,3)`
	type Resp struct {
		BalanceNum float64 `json:"balanceNum" form:"balanceNum" gorm:"column:balanceNum"`
	}
	var resp Resp
	if err := global.ZC_DB.Raw(sql).Scan(&resp).Error; err != nil {
		return 0, err
	}
	return resp.BalanceNum, nil
}
