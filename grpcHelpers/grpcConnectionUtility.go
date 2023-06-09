package grpcHelpers

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"strings"
)

func GetGRPCConnection(host string, port string, dialoptions []string) (conn *grpc.ClientConn, ctx context.Context) {

	var grpcDial grpc.DialOption

	for _,value := range dialoptions {
		values:=strings.Split(value,`:`)
		if values[0] == "WithInsecure" && values[1] == "true" {
			grpcDial=grpc.WithInsecure()
		}
		if values[0] == "PemFilePath" {
			creds,_ := credentials.NewClientTLSFromFile(values[1],"")
			grpcDial=grpc.WithTransportCredentials(creds)
		}
	}

	conn, err := grpc.Dial(host+":"+port,grpcDial)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	ctx, _ = context.WithCancel(context.Background())

	return
}
