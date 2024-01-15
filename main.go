package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"strconv"

	"github.com/mmfshirokan/GoProject1/proto/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	defaultUserData = &pb.UserData{
		Id:   110,
		Name: "Jhon",
		Male: true,
	}
	defaultPassword = "abcd"
)

// TODO separate test methods into functions
func main() {
	ctx := context.Background()
	target := "localhost:9091"
	option := grpc.WithTransportCredentials(insecure.NewCredentials())

	conn, err := grpc.Dial(target, option)
	if err != nil {
		log.Fatal("can't connect to rpc on ", target)
		return
	}
	defer conn.Close()

	usrClient := pb.NewUserClient(conn)
	tokClient := pb.NewTokenClient(conn)
	imgClient := pb.NewImageClient(conn)

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

	// userServer testing:

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

	// imageServer testing:

	UploadImage(authContext, imgClient, respSignIn.GetTokens().GetAuthToken())

	DownloadImage(authContext, imgClient, respSignIn.GetTokens().GetAuthToken())

}

func UploadImage(authContext context.Context, imgClient pb.ImageClient, authToken string) {
	stream, err := imgClient.UploadImage(authContext)
	if err != nil {
		log.Fatal("Upload image failed:", err)
	}

	imgName := "defaultUpload"
	stream.Send(&pb.RequestUploadImage{
		AuthToken:  &authToken,
		UserID:     &defaultUserData.Id,
		ImageName:  &imgName,
		ImagePiece: nil,
	})

	imgFull, err := os.ReadFile(("/home/andreishyrakanau/projects/project1/grpcStub/images/110-defaultUpload.png"))
	if err != nil {
		log.Fatal("Wrong image path", err)
	}

	imgPiece := make([]byte, 128)
	imgReader := bytes.NewReader(imgFull)

	for {
		_, err := imgReader.Read(imgPiece)
		if err == io.EOF {
			stream.CloseSend()
			return
		}
		if err != nil {
			log.Fatal(err)
		}

		stream.Send(&pb.RequestUploadImage{
			ImagePiece: imgPiece,
		})
	}
}

func DownloadImage(authContext context.Context, imgClient pb.ImageClient, authToken string) {
	stream, err := imgClient.DownloadImage(authContext, &pb.RequestDownloadImage{
		AuthToken: authToken,
		UserID:    defaultUserData.Id,
		ImageName: "defaultDownload",
	})
	if err != nil {
		log.Fatal("DownloadImage failed at the start:", err)
	}

	imgFull := make([]byte, 11000)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("reciev error:", err)
		}

		imgFull = append(imgFull, req.ImagePiece...)
	}

	err = stream.CloseSend()
	if err != nil {
		log.Error("failed to close stream", err)
	}

	img, _, err := image.Decode(bytes.NewReader(imgFull))
	if err != nil {
		log.Error(err)
		return
	}

	destFile, err := os.Create(ImgNameWrap(defaultUserData.Id, "defaultDownload"))
	if err != nil {
		log.Error(err)
		return
	}

	err = png.Encode(destFile, img)
	if err != nil {
		log.Error(err)
		return
	}
}

func ImgNameWrap(id int64, name string) string {
	return "/home/andreishyrakanau/projects/project1/grpcStub/images/" + strconv.FormatInt(id, 10) + "-" + name + ".png"
}
