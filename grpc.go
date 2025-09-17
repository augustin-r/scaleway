package main

import (
	"context"

	"github.com/cakturk/go-netstat/netstat"
	pb "github.com/personal/scaleway/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	pb.UnimplementedLoadBalancerAgentServer
	wrapper LoadBalancerWrapper
}

func (s *server) AddBinding(_ context.Context, in *pb.AddBindingRequest) (*pb.GenericReply, error) {
	bind := Bind{
		Name:    in.Name,
		Address: in.Address,
		Port:    int(in.Port),
	}
	if err := s.wrapper.AddBind(bind); err != nil {
		return &pb.GenericReply{Success: false}, err
	} else {
		return &pb.GenericReply{Success: true}, nil
	}
}

func (s *server) DeleteBinding(_ context.Context, in *pb.DeleteBindingRequest) (*pb.GenericReply, error) {
	if err := s.wrapper.DeleteBind(in.Name); err != nil {
		return &pb.GenericReply{Success: false}, err
	} else {
		return &pb.GenericReply{Success: true}, nil
	}
}

func (s *server) AddBackendServer(_ context.Context, in *pb.AddBackendServerRequest) (*pb.GenericReply, error) {
	backendServer := BackendServer{
		Name:    in.Name,
		Address: in.Address,
		Port:    int(in.Port),
	}
	if err := s.wrapper.AddBackendServer(backendServer); err != nil {
		return &pb.GenericReply{Success: false}, err
	} else {
		return &pb.GenericReply{Success: true}, nil
	}
}

func (s *server) DeleteBackendServer(_ context.Context, in *pb.DeleteBackendServerRequest) (*pb.GenericReply, error) {
	if err := s.wrapper.DeleteBackendServer(in.Name); err != nil {
		return &pb.GenericReply{Success: false}, err
	} else {
		return &pb.GenericReply{Success: true}, nil
	}
}

func (s *server) TrackSocks(_ context.Context, in *emptypb.Empty) (*pb.SocksReply, error) {
	result, errGetSocks := getSocks()
	if errGetSocks != nil {
		return &pb.SocksReply{}, errGetSocks
	}
	reply := &pb.SocksReply{
		SocksByState: make([]*pb.SocksByState, 0),
	}
	for k, v := range result {
		socksByState := &pb.SocksByState{State: k.String(), Socks: make([]*pb.Sock, len(v))}
		for _, sock := range v {
			socksByState.Socks = append(socksByState.Socks, &pb.Sock{
				Uuid:       sock.UID,
				LocalAddr:  sock.LocalAddr.String(),
				RemoteAddr: sock.RemoteAddr.String(),
			})
		}
		reply.SocksByState = append(reply.SocksByState, socksByState)
	}
	return reply, nil
}

func getSocks() (map[netstat.SkState][]netstat.SockTabEntry, error) {
	result := make(map[netstat.SkState][]netstat.SockTabEntry)

	// UDP sockets
	socks, err := netstat.UDPSocks(netstat.NoopFilter)
	if err != nil {
		return nil, err
	}
	for _, sock := range socks {
		result[sock.State] = append(result[sock.State], sock)
	}

	// TCP sockets
	socks, err = netstat.TCPSocks(netstat.NoopFilter)
	if err != nil {
		return nil, err
	}
	for _, sock := range socks {
		result[sock.State] = append(result[sock.State], sock)
	}

	return result, nil
}
