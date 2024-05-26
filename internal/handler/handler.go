package handler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kanataidarov/tinkoff_voicekit/internal/config"
	"github.com/kanataidarov/tinkoff_voicekit/internal/msgraph_client"
	"github.com/kanataidarov/tinkoff_voicekit/pkg/args"
	"github.com/kanataidarov/tinkoff_voicekit/pkg/common"
	"github.com/kanataidarov/tinkoff_voicekit/pkg/model/rest/msgraph_api"
	pb "github.com/kanataidarov/tinkoff_voicekit/pkg/teams_voicein"
	sttPb "github.com/kanataidarov/tinkoff_voicekit/pkg/tinkoff_voicekit/cloud/stt/v1"
	"google.golang.org/grpc"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type server struct {
	pb.UnimplementedSpeechToTextServer
}

func (s *server) Recognize(_ context.Context, req *pb.SttRequest) (*pb.SttResponse, error) {
	var (
		header *pb.FileHeader
		buf    bytes.Buffer
	)

	header = req.Header
	fileName := filepath.Base(header.Name)
	log.Printf("File name: %v", fileName)
	if header.Size != nil {
		log.Printf("File size should be: %d", header.Size)
	}
	if data := req.Data; data != nil {
		buf.Write(data)
	}

	log.Printf("Total bytes received: %v", buf.Len())

	res := doRecognize(buf, fileName)

	doTeams(res)

	return &pb.SttResponse{Message: res}, nil
}

func doRecognize(buf bytes.Buffer, fileName string) string {
	prepareFile(buf, fileName)

	opts := args.ParseRecognizeOptions()
	if opts == nil {
		os.Exit(1)
	}
	defer func(InputFile *os.File) {
		_ = InputFile.Close()
	}(opts.InputFile)

	var dataReader io.Reader
	if strings.HasSuffix(opts.InputFile.Name(), ".wav") {
		reader, err := common.OpenWavFormat(opts.InputFile, "LINEAR16", 1, 16000)
		if err != nil {
			panic(err)
		}
		dataReader = reader
	} else {
		dataReader = opts.InputFile
	}

	client, err := common.NewSttClient(opts.CommonOptions)
	if err != nil {
		panic(err)
	}
	defer func(client common.SpeechToTextClient) {
		_ = client.Close()
	}(client)

	contents, err := io.ReadAll(dataReader)
	if err != nil {
		panic(err)
	}

	request := &sttPb.RecognizeRequest{
		Config: &sttPb.RecognitionConfig{
			Encoding:                   sttPb.AudioEncoding(sttPb.AudioEncoding_value[*opts.Encoding]),
			SampleRateHertz:            uint32(*opts.Rate),
			LanguageCode:               *opts.LanguageCode,
			MaxAlternatives:            uint32(*opts.MaxAlternatives),
			ProfanityFilter:            !(*opts.DisableProfanityFilter),
			EnableAutomaticPunctuation: !(*opts.DisableAutomaticPunctuation),
			NumChannels:                uint32(*opts.NumChannels),
		},
		Audio: &sttPb.RecognitionAudio{
			AudioSource: &sttPb.RecognitionAudio_Content{Content: contents},
		},
	}
	if *opts.DoNotPerformVad {
		request.Config.Vad = &sttPb.RecognitionConfig_DoNotPerformVad{DoNotPerformVad: true}
	} else {
		request.Config.Vad = &sttPb.RecognitionConfig_VadConfig{
			VadConfig: &sttPb.VoiceActivityDetectionConfig{
				SilenceDurationThreshold: float32(*opts.SilenceDurationThreshold),
			},
		}
	}

	// context.WithCancel in prod
	result, err := client.Recognize(context.Background(), request)
	if err != nil {
		panic(err)
	}

	transcript := result.Results[0].Alternatives[0].Transcript

	return transcript
}

func prepareFile(buf bytes.Buffer, fileName string) *os.File {
	file, err := os.Create(fileName)
	if err != nil {
		log.Print("Could not create file", err)
		return nil
	}

	length, err := file.Write(buf.Bytes())
	if err != nil {
		_ = file.Close()
		log.Print("Could not write to file", err)
		return nil
	}
	log.Printf("Wrote %d bytes to file", length)

	err = file.Close()
	if err != nil {
		log.Print("Could not close file", err)
		return nil
	}

	return file
}

func Serve(cfg *config.Config, log *slog.Logger) {
	port := cfg.Grpc.Port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Error(fmt.Sprintf("Failed to listen at %d. %v", port, err))
	}

	srv := grpc.NewServer()
	pb.RegisterSpeechToTextServer(srv, &server{})
	log.Info(fmt.Sprintf("Server listening on %v", lis.Addr()))
	if err := srv.Serve(lis); err != nil {
		log.Error(fmt.Sprintf("Failed to serve at %d. %v", port, err))
	}
}

func Context() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	return ctx
}

func doTeams(msgContent string) {
	userId := "a2619b8a-6136-43df-9910-692af6264c0b"

	chats, err := msgraph_client.Get[msgraph_api.ChatsResponse](Context(), msgraph_client.ChatsUrl())
	if err != nil {
		log.Printf("Error getting chats: %v", err)
	}

	chat := msgraph_api.Chat{}
	chatItems := chats.Value
	for _, chatItem := range chatItems {
		if chatItem.LastMsg.LastMsgFrom.User.Id == userId {
			log.Printf("Last Msg Id: " + chatItem.LastMsg.Id)
			chat = chatItem
			break
		}
	}

	msgRequest := msgraph_api.MsgRequest{MsgBody: msgraph_api.MsgBody{Content: msgContent, ContentType: "text"}}
	msgResponse, err := msgraph_client.Post[msgraph_api.MsgResponse](Context(), msgraph_client.PostMsgUrl(chat.Id), msgRequest)
	if err != nil {
		log.Printf("Error posting msg: %v", err)
	}
	log.Printf("Msg response: %+v", msgResponse)
}
