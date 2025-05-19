package tests

import (
	"context"
	godesk "github.com/getcharzp/godesk-serve/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"testing"
)

func getDeviceServiceClient() (context.Context, godesk.DeviceServiceClient) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
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

func TestGetDeviceList(t *testing.T) {
	ctx, client := getDeviceServiceClient()
	resp, err := client.GetDeviceList(ctx, &godesk.DeviceListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(prettyPrint(resp))
}

func TestAddDevice(t *testing.T) {
	ctx, client := getDeviceServiceClient()
	resp, err := client.AddDevice(ctx, &godesk.AddDeviceRequest{
		Code:     100000003,
		Password: "123456",
		Remark:   "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}

func TestEditDevice(t *testing.T) {
	ctx, client := getDeviceServiceClient()
	resp, err := client.EditDevice(ctx, &godesk.EditDeviceRequest{
		Uuid:     "ba555df1-7167-cdd0-9cc3-97fbbc50e011",
		Code:     100000003,
		Password: "123456",
		Remark:   "100000003（test）",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}

func TestDeleteDevice(t *testing.T) {
	ctx, client := getDeviceServiceClient()
	resp, err := client.DeleteDevice(ctx, &godesk.DeleteDeviceRequest{
		Uuid: "e372d186-648b-e777-f99c-8a3134ef0333",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}
