// Package agent sets up and connects all the different components that make this service work.
// Each component (log, membership, replicator, and server) is responsible for an aspect of service behaviour.
// The WAL agent is in effect the service coordinator it references and manages the components.
package agent

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"github.com/percybear/wal/internal/auth"
	"github.com/percybear/wal/internal/discovery"
	"github.com/percybear/wal/internal/log"
	"github.com/percybear/wal/internal/server"

	// cmux is a generic Go library to multiplex connections based on their payload. Using cmux, you can
	// serve gRPC, SSH, HTTPS, HTTP, Go RPC, and pretty much any other protocol on the same TCP listener.
	"github.com/soheilhy/cmux"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// A Config contains the configuration an agent needs to setup its components.
type Config struct {
	// ServerTLSConfig is the TLS config that secures the client-server connections.
	ServerTLSConfig *tls.Config
	// PeerTLSConfig is the TLS config that secures the consensus protocol connections.
	PeerTLSConfig *tls.Config
	// DataDir stores the log and raft data.
	DataDir string
	// BindAddr is the address serf runs on.
	BindAddr string
	// RPCPort is the port for client (and Raft) connections.
	RPCPort int
	// Raft server id.
	NodeName string
	// Bootstrap should be set to true when starting the first node of the cluster.
	StartJoinAddrs []string
	ACLModelFile   string
	ACLPolicyFile  string
	// START: config
	Bootstrap bool
}

func (c Config) RPCAddr() (string, error) {
	host, _, err := net.SplitHostPort(c.BindAddr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", host, c.RPCPort), nil
}

// A Agent is the component coordinator.
type Agent struct {
	Config Config // config contains the configuration an agent needs to setup its componets.

	mux        cmux.CMux             // cmux is a library to multiplex connections.
	log        *log.DistributedLog   // log is the distributed log.
	server     *grpc.Server          // server is the gRPC server.
	membership *discovery.Membership // membership tracks the other agents in this consensus group.

	shutdown     bool          // shutdown is true when the agent is shutting down.
	shutdowns    chan struct{} // shutdowns is a channel that is closed when the agent is shutting down.
	shutdownLock sync.Mutex    // shutdownLock is a lock that protects the shutdown state.
}

// New creates a new agent including its components.
func New(config Config) (*Agent, error) {
	a := &Agent{
		Config:    config,
		shutdowns: make(chan struct{}),
	}
	// START: add_setup_mux
	setup := []func() error{
		// START_HIGHLIGHT
		a.setupMux,
		// END_HIGHLIGHT
		a.setupLog,
		a.setupServer,
		a.setupMembership,
	}
	// END: add_setup_mux
	for _, fn := range setup {
		if err := fn(); err != nil {
			return nil, err
		}
	}
	// START: new_serve
	go a.serve()
	// END: new_serve
	return a, nil
}

// setupMux create your listener and a cmux for that listener, you can use the cmux to match connections.
func (a *Agent) setupMux() error {
	//
	rpcAddr := fmt.Sprintf(
		":%d",
		a.Config.RPCPort,
	)
	// Create the main listener.
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return err
	}
	// Create a cmux for that listener, you can now use the cmux to match connections.
	a.mux = cmux.New(ln)

	return nil
}

// setupLog creates a new log for the raft stream layer.
func (a *Agent) setupLog() error {
	// identify connections by reading the raft rpc constant.
	// If mux matches this rule, pass the connection to the raft listener.
	raftLn := a.mux.Match(func(reader io.Reader) bool {
		b := make([]byte, 1)
		if _, err := reader.Read(b); err != nil {
			return false
		}
		return bytes.Equal(b, []byte{byte(log.RaftRPC)})
	})
	logConfig := log.Config{}
	logConfig.Raft.StreamLayer = log.NewStreamLayer(
		raftLn,
		a.Config.ServerTLSConfig,
		a.Config.PeerTLSConfig,
	)
	logConfig.Raft.LocalID = raft.ServerID(a.Config.NodeName)
	logConfig.Raft.Bootstrap = a.Config.Bootstrap
	var err error
	a.log, err = log.NewDistributedLog(
		a.Config.DataDir,
		logConfig,
	)
	if err != nil {
		return err
	}
	if a.Config.Bootstrap {
		return a.log.WaitForLeader(3 * time.Second)
	}
	return nil
}

// setupServer creates a new server for the gRPC connections.
func (a *Agent) setupServer() error {
	authorizer := auth.New(
		a.Config.ACLModelFile,
		a.Config.ACLPolicyFile,
	)
	serverConfig := &server.Config{
		CommitLog:  a.log,
		Authorizer: authorizer,
	}
	var opts []grpc.ServerOption
	if a.Config.ServerTLSConfig != nil {
		creds := credentials.NewTLS(a.Config.ServerTLSConfig)
		opts = append(opts, grpc.Creds(creds))
	}
	var err error
	a.server, err = server.NewGRPCServer(serverConfig, opts...)
	if err != nil {
		return err
	}

	// Use the cmux to match connections.
	grpcLn := a.mux.Match(cmux.Any())
	go func() {
		if err := a.server.Serve(grpcLn); err != nil {
			_ = a.Shutdown()
		}
	}()
	return nil
}

// setupMembership creates a new list of agents in this consensus group.
func (a *Agent) setupMembership() error {
	rpcAddr, err := a.Config.RPCAddr()
	if err != nil {
		return err
	}
	a.membership, err = discovery.New(a.log, discovery.Config{
		NodeName: a.Config.NodeName,
		BindAddr: a.Config.BindAddr,
		Tags: map[string]string{
			"rpc_addr": rpcAddr,
		},
		StartJoinAddrs: a.Config.StartJoinAddrs,
	})
	return err
}

// serve starts multiplexing the listener.
// Serve blocks and perhaps should be invoked concurrently within a go routine.
func (a *Agent) serve() error {
	if err := a.mux.Serve(); err != nil {
		_ = a.Shutdown()
		return err
	}
	return nil
}

// Shutdown ensures we shut down the agent once even if shutdown is called multiple times.
// Shutdown is managed gracefully by leaving this consensus group so that other members stop sending it events,
// then closing the replicator so that it stops replicating, then stopping the server and finally closing the log.
func (a *Agent) Shutdown() error {
	a.shutdownLock.Lock()
	defer a.shutdownLock.Unlock()
	if a.shutdown {
		return nil
	}
	a.shutdown = true
	close(a.shutdowns)

	shutdown := []func() error{
		a.membership.Leave,
		func() error {
			a.server.GracefulStop()
			return nil
		},
		a.log.Close,
	}
	for _, fn := range shutdown {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
