package domain

type ProducerId uint

type DegradationRate uint

type Restoration struct {
	Require ProcessSheet
	Power   Power
}

type Upgrade struct {
	Require ProcessSheet
	Power   Power
}

type Task struct {
	OrderId OrderId
	Agent   *OrderingAgent
}

type ProducerInfo struct {
	PowerType   PowerType
	Capacity    Power
	CutOffPrice Tokens
}

type ProducingAgentConfig struct {
	Id          ProducerId
	Type        PowerType
	Capacity    Power
	Degradation DegradationRate
	Restoration Restoration
	Upgrade     Upgrade
}

type Bid struct {
	Power   Power
	Tokens  Tokens
	OrderId OrderId
}

type ProducingAgent struct {
	id          ProducerId
	capacity    Power
	degradation DegradationRate
	restoration Restoration
	upgrade     Upgrade
	bids        []Bid // replace with MaxHeap
	inProgress  []Task
}
