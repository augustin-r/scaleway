package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/personal/scaleway/proto"
)

var (
	port     = flag.Int("port", 50051, "The server port")
	address  = flag.String("address", "http://localhost:5555", "The HAProxy DataPlane API address")
	user     = flag.String("user", "admin", "The username to access HAProxy DataPlane API")
	password = flag.String("password", "password", "The password to access HAProxy DataPlane API")
)

func main() {
	flag.Parse()

	lis, errListen := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if errListen != nil {
		log.Fatal().Msgf("failed to listen: %v", errListen)
	}
	s := grpc.NewServer()
	pb.RegisterLoadBalancerAgentServer(s, &server{
		wrapper: NewLoadBalancerWrapperImpl(*address, *user, *password),
	})

	// WARN Deactive in prd or only add with a debug flag
	reflection.Register(s)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatal().Msgf("failed to serve: %v", err)
	}
}
