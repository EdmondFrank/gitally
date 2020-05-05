package nodes

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/metrics"
)

type nodeCandidate struct {
	m        sync.RWMutex
	node     Node
	statuses []bool
}

// localElector relies on an in-memory datastore to track the primary
// and secondaries. A single election strategy pertains to a single
// shard. It does NOT support multiple Praefect nodes or have any
// persistence. This is used mostly for testing and development.
type localElector struct {
	m               sync.RWMutex
	failoverEnabled bool
	shardName       string
	nodes           []*nodeCandidate
	primaryNode     *nodeCandidate
	log             logrus.FieldLogger
}

// healthcheckThreshold is the number of consecutive healthpb.HealthCheckResponse_SERVING necessary
// for deeming a node "healthy"
const healthcheckThreshold = 3

func (n *nodeCandidate) checkNode(ctx context.Context) {
	status, _ := n.node.check(ctx)

	n.m.Lock()
	defer n.m.Unlock()

	n.statuses = append(n.statuses, status)

	if len(n.statuses) > healthcheckThreshold {
		n.statuses = n.statuses[1:]
	}
}

func (n *nodeCandidate) isHealthy() bool {
	n.m.RLock()
	defer n.m.RUnlock()

	if len(n.statuses) < healthcheckThreshold {
		return false
	}

	for _, status := range n.statuses[len(n.statuses)-healthcheckThreshold:] {
		if !status {
			return false
		}
	}

	return true
}

func newLocalElector(name string, failoverEnabled bool, log logrus.FieldLogger, ns []*nodeStatus) *localElector {
	nodes := make([]*nodeCandidate, len(ns))
	for i, n := range ns {
		nodes[i] = &nodeCandidate{
			node: n,
		}
	}

	return &localElector{
		shardName:       name,
		failoverEnabled: failoverEnabled,
		log:             log,
		nodes:           nodes,
		primaryNode:     nodes[0],
	}
}

// Start launches a Goroutine to check the state of the nodes and
// continuously monitor their health via gRPC health checks.
func (s *localElector) start(bootstrapInterval, monitorInterval time.Duration) {
	s.bootstrap(bootstrapInterval)
	go s.monitor(monitorInterval)
}

func (s *localElector) bootstrap(d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()

	for i := 0; i < healthcheckThreshold; i++ {
		<-timer.C

		ctx := context.TODO()
		s.checkNodes(ctx)
		timer.Reset(d)
	}
}

func (s *localElector) monitor(d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	ctx := context.Background()

	for {
		<-ticker.C

		err := s.checkNodes(ctx)

		if err != nil {
			s.log.WithError(err).Warn("error checking nodes")
		}
	}
}

// checkNodes issues a gRPC health check for each node managed by the
// shard.
func (s *localElector) checkNodes(ctx context.Context) error {
	defer s.updateMetrics()

	for _, n := range s.nodes {
		n.checkNode(ctx)
	}

	s.m.Lock()
	defer s.m.Unlock()

	if s.primaryNode != nil && s.primaryNode.isHealthy() {
		return nil
	}

	var newPrimary *nodeCandidate

	for _, node := range s.nodes {
		if node != s.primaryNode && node.isHealthy() {
			newPrimary = node
			break
		}
	}

	if newPrimary == nil {
		return ErrPrimaryNotHealthy
	}

	s.primaryNode = newPrimary

	return nil
}

// GetPrimary gets the primary of a shard. If no primary exists, it will
// be nil. If a primary has been elected but is down, err will be
// ErrPrimaryNotHealthy.
func (s *localElector) GetPrimary() (Node, error) {
	s.m.RLock()
	primary := s.primaryNode
	s.m.RUnlock()

	if primary == nil {
		return nil, ErrPrimaryNotHealthy
	}

	if s.failoverEnabled && !primary.isHealthy() {
		return primary.node, ErrPrimaryNotHealthy
	}

	return primary.node, nil
}

// GetSecondaries gets the secondaries of a shard
func (s *localElector) GetSecondaries() ([]Node, error) {
	s.m.RLock()
	primary := s.primaryNode
	s.m.RUnlock()

	var secondaries []Node
	for _, n := range s.nodes {
		if n != primary {
			secondaries = append(secondaries, n.node)
		}
	}

	return secondaries, nil
}

func (s *localElector) updateMetrics() {
	s.m.RLock()
	primary := s.primaryNode
	s.m.RUnlock()

	for _, n := range s.nodes {
		var val float64

		if n == primary {
			val = 1
		}

		metrics.PrimaryGauge.WithLabelValues(s.shardName, n.node.GetStorage()).Set(val)
	}
}