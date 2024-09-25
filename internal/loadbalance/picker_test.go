// START: start
package loadbalance_test

import (
	"testing"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"

	"github.com/stretchr/testify/require"

	"github.com/percybear/wal/internal/loadbalance"
)

func TestPickerNoSubConnAvailable(t *testing.T) {
	picker := &loadbalance.Picker{}
	for _, method := range []string{
		"/log.vX.Log/Produce",
		"/log.vX.Log/Consume",
	} {
		info := balancer.PickInfo{
			FullMethodName: method,
		}
		result, err := picker.Pick(info)
		require.Equal(t, balancer.ErrNoSubConnAvailable, err)
		require.Nil(t, result.SubConn)
	}
}

// END: start

// START: next
func TestPickerProducesToLeader(t *testing.T) {
	picker, subConns := setupTest()
	info := balancer.PickInfo{
		FullMethodName: "/log.vX.Log/Produce",
	}
	for i := 0; i < 5; i++ {
		gotPick, err := picker.Pick(info)
		require.NoError(t, err)
		require.Equal(t, subConns[0], gotPick.SubConn)
	}
}

func TestPickerConsumesFromFollowers(t *testing.T) {
	picker, subConns := setupTest()
	info := balancer.PickInfo{
		FullMethodName: "/log.vX.Log/Consume",
	}
	for i := 0; i < 5; i++ {
		pick, err := picker.Pick(info)
		require.NoError(t, err)
		require.Equal(t, subConns[i%2+1], pick.SubConn)
	}
}

// END: next

// START: end
func setupTest() (*loadbalance.Picker, []*subConn) {
	var subConns []*subConn
	buildInfo := base.PickerBuildInfo{
		ReadySCs: make(map[balancer.SubConn]base.SubConnInfo),
	}
	for i := 0; i < 3; i++ {
		sc := &subConn{}
		addr := resolver.Address{
			Attributes: attributes.New("is_leader", i == 0),
		}
		// 0th sub conn is the leader
		sc.UpdateAddresses([]resolver.Address{addr})
		buildInfo.ReadySCs[sc] = base.SubConnInfo{Address: addr}
		subConns = append(subConns, sc)
	}
	picker := &loadbalance.Picker{}
	picker.Build(buildInfo)
	return picker, subConns
}

// subConn implements balancer.SubConn.
type subConn struct {
	addrs []resolver.Address
}

func (s *subConn) UpdateAddresses(addrs []resolver.Address) {
	s.addrs = addrs
}

// GetOrBuildProducer returns a reference to the existing Producer for this
// ProducerBuilder in this SubConn, or, if one does not currently exist,
// creates a new one and returns it.  Returns a close function which must
// be called when the Producer is no longer needed.
func (s *subConn) GetOrBuildProducer(balancer.ProducerBuilder) (p balancer.Producer, close func()) {
	return nil, func() {}
}

// func (s *subConn) UpdateClientConnState(balancer.ClientConnState) error {
// 	return nil
// }

// Connect starts the connecting for this SubConn.
func (s *subConn) Connect() {}

// Shutdown shuts down the SubConn gracefully.  Any started RPCs will be
// allowed to complete.  No future calls should be made on the SubConn.
// One final state update will be delivered to the StateListener (or
// UpdateSubConnState; deprecated) with ConnectivityState of Shutdown to
// indicate the shutdown operation.  This may be delivered before
// in-progress RPCs are complete and the actual connection is closed.
func (s *subConn) Shutdown() {}

// END: end
