package contract

const (
	VoteGetWarnNode    MethodID = "getWarnNode()"
	VoteSolidifiedVote MethodID = "solidifiedVote(address)"
	VoteCheckWarnNode  MethodID = "checkWarnNode()"
	VoteDelwarnNodeMap MethodID = "delwarnNodeMap(address)"
	VoteTakeVote       MethodID = "takeVote(address,uint256)"
)

var VoteContract = Contract{
	Addr:     "0x574C255649719Fd05d6dd18A1B1d1aCd26e6620e",
	PushAddr: "f02821643",
}
