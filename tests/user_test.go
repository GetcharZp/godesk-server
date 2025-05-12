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

func getUserServiceClient() (context.Context, godesk.UserServiceClient) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	return context.Background(), godesk.NewUserServiceClient(conn)
}

func TestGetUserInfo(t *testing.T) {
	ctx, client := getUserServiceClient()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
	info, err := client.GetUserInfo(ctx, &godesk.EmptyRequest{})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}

func TestUserRegister(t *testing.T) {
	ctx, client := getUserServiceClient()
	info, err := client.UserRegister(ctx, &godesk.UserRegisterRequest{
		Username: "admin",
		Password: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}

func TestUserLogin(t *testing.T) {
	ctx, client := getUserServiceClient()
	info, err := client.UserLogin(ctx, &godesk.UserLoginRequest{
		Username: "admin",
		Password: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}
