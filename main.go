package main

import (
	"context"
	"fmt"

	"github.com/mmfshirokan/GoProject1/proto/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	ctx := context.Background()
	target := "localhost:9091"
	defaultUserData := &pb.UserData{
		Id:   110,
		Name: "Jhon",
		Male: true,
	}
	defaultPassword := "abcd"

	option := grpc.WithTransportCredentials(insecure.NewCredentials())

	conn, err := grpc.Dial(target, option)
	if err != nil {
		log.Fatal("can't connect to rpc on ", target)
		return
	}
	defer conn.Close()

	usrClient := pb.NewUserClient(conn)
	tokClient := pb.NewTokenClient(conn)

	// tokenServer testing:

	_, err = tokClient.SignUp(ctx, &pb.RequestSignUp{
		Data:     defaultUserData,
		Password: defaultPassword,
	})
	if err != nil {
		log.Info("SignUp method failed")
		log.Error(err)
	} else {
		log.Info("SignUp method passed")
	}

	respSignIn, err := tokClient.SignIn(ctx, &pb.RequestSignIn{
		UserID:   defaultUserData.Id,
		Password: defaultPassword,
	})
	if err != nil {
		log.Info("SignIn method failed")
		log.Error(err)
	} else {
		log.Info(fmt.Sprintf("SignIn mrthod passed with: %s, %s", respSignIn.GetTokens().GetAuthToken(), respSignIn.GetTokens().GetRft().Hash))
	}

	respRefresh, err := tokClient.Refresh(ctx, &pb.RequestRefresh{
		Rft: respSignIn.Tokens.Rft,
	})
	if err != nil {
		log.Info("Refrsh method failed")
		log.Error(err)
	} else {
		log.Info(fmt.Sprintf("Refrsh method passed with: %s, %s", respSignIn.GetTokens().GetAuthToken(), respRefresh.Tokens.GetRft().Hash))
	}

	//userServer testing:

	authContext := metadata.AppendToOutgoingContext(ctx, "authorization", respSignIn.GetTokens().GetAuthToken())

	respGetUser, err := usrClient.GetUser(authContext, &pb.RequestGetUser{
		AuthToken: respSignIn.GetTokens().GetAuthToken(),
		UserID:    defaultUserData.Id,
	})
	if err != nil {
		log.Info("GetUser failed")
		log.Error(err)
	} else {
		log.Info(fmt.Printf("GetUser passed with: %v, %v, %v", respGetUser.GetData().GetId(), respGetUser.GetData().GetName(), respGetUser.GetData().GetMale()))
	}

	_, err = usrClient.UpdateUser(authContext, &pb.RequestUpdateUser{
		AuthToken: respSignIn.GetTokens().GetAuthToken(),
		Data: &pb.UserData{
			// Id: defaultUserData.Id, // ?
			Name: "Markus",
			Male: true,
		},
	})
	if err != nil {
		log.Info("UpdateUser failed")
		log.Error(err)
	} else {
		log.Info("UpdateUser passed")
	}

	usrClient.DeleteUser(authContext, &pb.RequestDelete{
		AuthToken: respSignIn.GetTokens().GetAuthToken(),
		UserID:    defaultUserData.Id,
	})
	if err != nil {
		log.Info("DeleteUser failed")
		log.Error(err)
	} else {
		log.Info("DeleteUser passed")
	}
}
