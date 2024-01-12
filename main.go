package main

import (
	"context"

	"github.com/mmfshirokan/GoProject1/proto/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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

	conn, err := grpc.Dial(target)
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
		log.Debug("SignUp method failed")
		log.Error(err)
	} else {
		log.Debug("SignUp method passed")
	}

	respSignIn, err := tokClient.SignIn(ctx, &pb.RequestSignIn{
		UserID:   defaultUserData.Id,
		Password: defaultPassword,
	})
	if err != nil {
		log.Debug("SignIn method failed")
		log.Error(err)
	} else {
		log.Debug("SignIn mrthod passed with: %s, %s", respSignIn.GetTokens().GetAuthToken(), respSignIn.GetTokens().GetRft().Hash)
	}

	respRefresh, err := tokClient.Refresh(ctx, &pb.RequestRefresh{
		Rft: respSignIn.Tokens.Rft,
	})
	if err != nil {
		log.Debug("Refrsh method failed")
		log.Error(err)
	} else {
		log.Debug("Refrsh method passed with: %s, %s", respSignIn.GetTokens().GetAuthToken(), respRefresh.Tokens.GetRft().Hash)
	}

	//userServer testing:

	respGetUser, err := usrClient.GetUser(ctx, &pb.RequestGetUser{
		AuthToken: respSignIn.GetTokens().GetAuthToken(),
		UserID:    defaultUserData.Id,
	})
	if err != nil {
		log.Debug("GetUser failed")
		log.Error(err)
	} else {
		log.Debug("GetUser passed with: %d, %s, %t", respGetUser.GetData().GetId(), respGetUser.GetData().GetName(), respGetUser.GetData().GetMale())
	}

	_, err = usrClient.UpdateUser(ctx, &pb.RequestUpdateUser{
		AuthToken: respSignIn.GetTokens().GetAuthToken(),
		Data: &pb.UserData{
			// Id: defaultUserData.Id, // ?
			Name: "Markus",
			Male: true,
		},
	})
	if err != nil {
		log.Debug("UpdateUser failed")
		log.Error(err)
	} else {
		log.Debug("UpdateUser passed")
	}

	usrClient.DeleteUser(ctx, &pb.RequestDelete{
		AuthToken: respSignIn.GetTokens().GetAuthToken(),
		UserID:    defaultUserData.Id,
	})
	if err != nil {
		log.Debug("DeleteUser failed")
		log.Error(err)
	} else {
		log.Debug("DeleteUser passed")
	}
}
