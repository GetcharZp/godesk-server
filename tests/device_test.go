package tests

import (
	"context"
	godesk "github.com/getcharzp/godesk-serve/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"testing"
)

func getDeviceServiceClient() (context.Context, godesk.DeviceServiceClient) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	return ctx, godesk.NewDeviceServiceClient(conn)
}

func TestGetDeviceInfo(t *testing.T) {
	ctx, client := getDeviceServiceClient()
	info, err := client.GetDeviceInfo(ctx, &godesk.DeviceInfoRequest{
		Uuid: "1",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}

func TestCreateDevice(t *testing.T) {
	ctx, client := getDeviceServiceClient()
	info, err := client.CreateDevice(ctx, &godesk.CreateDeviceRequest{
		Os: "windows",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}
