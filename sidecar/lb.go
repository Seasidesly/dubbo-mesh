package sidecar

import (
	"errors"
	"sync"

	"dubbo-mesh/registry"
	"dubbo-mesh/util"
)

const (
	LB_Random  = iota
	LB_RR
	LB_WRR
	LB_LAvg
	LB_LActive
	LB_WLA
)

const (
	AVG_COUNT = 16
)

func lb(elector int) Banlancer {
	switch elector {
	case LB_Random:
		return &Random{}
	case LB_RR:
		return &RoundRobin{}
	case LB_WRR:
		return &WeightRoundRobin{}
	case LB_LAvg:
		return &LeastAVG{}
	case LB_LActive:
		return &LeastActive{}
	case LB_WLA:
		return &WeightLeastLatestAvg{}
	default:
		panic(errors.New("unknown load balancer"))
	}
}

type Banlancer interface {
	Init(endpoint []*Endpoint)
	Elect(endpoints []*Endpoint) *Endpoint
}

type Endpoint struct {
	*registry.Endpoint
	Meter *Meter
}

func (this *Endpoint) String() string {
	m := make(map[string]interface{}, 3)
	m["name"] = this.System.Name
	m["avg"] = this.Meter.Avg()
	m["meter"] = this.Meter
	return util.ToJsonStr(m)
}

type Meter struct {
	mtx    sync.Mutex
	Queue  *queue.Queue
	Latest uint64 `json:"latest"`
	Count  uint64 `json:"count,omitempty"`  // 已处理的总数
	Active int32  `json:"active,omitempty"` // 当前连接数
	Total  uint64 `json:"total,omitempty"`
}

// 最近平均值
func (this *Meter) Record(latest uint64) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.Latest = latest
	this.Count += 1
	this.Total += latest
	if this.Count >= AVG_COUNT {
		val := this.Queue.Remove()
		this.Total -= val.(uint64)
	}
	this.Queue.Add(latest)
}

// 最近32平均值
func (this *Meter) Avg() uint64 {
	if this.Count == 0 {
		return 0
	}
	if this.Count < AVG_COUNT {
		return this.Total / this.Count
	} else {
		return this.Total / AVG_COUNT
	}
}

func NewEndpoint(endpoint *registry.Endpoint) *Endpoint {
	return &Endpoint{
		Endpoint: endpoint,
		Meter: &Meter{
			Queue: queue.New(),
		},
	}
}
