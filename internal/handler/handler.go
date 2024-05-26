package handler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kanataidarov/tinkoff_voicekit/internal/config"
	mc "github.com/kanataidarov/tinkoff_voicekit/internal/msgraph_client"
	"github.com/kanataidarov/tinkoff_voicekit/pkg/common"
	mapi "github.com/kanataidarov/tinkoff_voicekit/pkg/model/rest/msgraph_api"
	our "github.com/kanataidarov/tinkoff_voicekit/pkg/teams_voicein"
	tkof "github.com/kanataidarov/tinkoff_voicekit/pkg/tinkoff_voicekit/cloud/stt/v1"
	"github.com/kanataidarov/tinkoff_voicekit/pkg/types"
	"google.golang.org/grpc"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type server struct {
	our.UnimplementedSpeechToTextServer
}

func (s *server) Recognize(ctx context.Context, req *our.SttRequest) (*our.SttResponse, error) {
	cfg := config.Load()
	logger := config.InitLogger(cfg.Env)
	ctx = context.WithValue(ctx, "values", types.CtxVals{Config: cfg, Logger: logger})

	var (
		header *our.FileHeader
		buf    bytes.Buffer
	)

	header = req.Header
	fileName := filepath.Base(header.Name)
	logger.Info("", "File name", fileName)
	if header.Size != nil {
		logger.Debug("File size should be: ", "", header.Size)
	}
	if data := req.Data; data != nil {
		buf.Write(data)
	}
	logger.Debug("Total bytes received: ", "", buf.Len())

	res := doRecognize(ctx, buf, fileName)

	doTeams(ctx, res)

	return &our.SttResponse{Message: res}, nil
}

func doRecognize(ctx context.Context, buf bytes.Buffer, fileName string) string {
	ctxVals := ctx.Value("values").(types.CtxVals)
	cfg := ctxVals.Config
	logger := ctxVals.Logger

	prepareFile(ctx, buf, fileName)

	file, err := os.Open(fileName)
	if err != nil {
		logger.Error("Error opening file", "Error", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var dataReader io.Reader
	if strings.HasSuffix(file.Name(), ".wav") {
		reader, err := common.OpenWavFormat(file, cfg.Audio.Encoding, cfg.Audio.Chans, cfg.Audio.SampleRate)
		if err != nil {
			panic(err)
		}
		dataReader = reader
	} else {
		dataReader = file
	}

	client, err := common.NewSttClient(cfg)
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

	request := &tkof.RecognizeRequest{
		Config: &tkof.RecognitionConfig{
			Encoding:                   tkof.AudioEncoding(tkof.AudioEncoding_value[cfg.Audio.Encoding]),
			SampleRateHertz:            uint32(cfg.Audio.SampleRate),
			LanguageCode:               cfg.Audio.LanguageCode,
			MaxAlternatives:            uint32(cfg.Audio.MaxAlternatives),
			ProfanityFilter:            cfg.Audio.ProfanityFilter,
			EnableAutomaticPunctuation: cfg.Audio.AutomaticPunctuation,
			NumChannels:                uint32(cfg.Audio.Chans),
		},
		Audio: &tkof.RecognitionAudio{
			AudioSource: &tkof.RecognitionAudio_Content{Content: contents},
		},
	}
	if !cfg.Audio.PerformVad {
		request.Config.Vad = &tkof.RecognitionConfig_DoNotPerformVad{DoNotPerformVad: true}
	} else {
		request.Config.Vad = &tkof.RecognitionConfig_VadConfig{
			VadConfig: &tkof.VoiceActivityDetectionConfig{
				SilenceDurationThreshold: float32(cfg.Audio.SilenceDurationThreshold),
			},
		}
	}

	result, err := client.Recognize(ctx, request)
	if err != nil {
		logger.Error("Error calling Recognize", "Error", err)
		return ""
	}

	transcript := result.Results[0].Alternatives[0].Transcript

	return transcript
}

func prepareFile(ctx context.Context, buf bytes.Buffer, fileName string) *os.File {
	ctxVals := ctx.Value("values").(types.CtxVals)
	logger := ctxVals.Logger

	file, err := os.Create(fileName)
	if err != nil {
		logger.Error("Error opening file", "Error", err)
		return nil
	}

	length, err := file.Write(buf.Bytes())
	if err != nil {
		_ = file.Close()
		logger.Error("Could not write to file", "Error", err)
		return nil
	}
	logger.Info(fmt.Sprintf("Wrote %d bytes to file", length))

	err = file.Close()
	if err != nil {
		logger.Error("Could not close file", "Error", err)
		return nil
	}

	return file
}

func Serve(_ context.Context, cfg *config.Config, log *slog.Logger) error {
	port := cfg.Grpc.Port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("Failed to listen at ", "Port", port, "Error", err)
		return err
	}

	srv := grpc.NewServer()
	our.RegisterSpeechToTextServer(srv, &server{})
	log.Info("Server listening on ", "Address", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		slog.Error("Failed to serve at ", "Port", port, "Error", err)
		return err
	}

	return nil
}

func doTeams(ctx context.Context, msgContent string) {
	ctxVals := ctx.Value("values").(types.CtxVals)
	logger := ctxVals.Logger

	user, err := mc.Get[mapi.User](ctx, mc.ProfileUrl())
	if err != nil {
		logger.Error("Error getting user profile", "Error", err)
	}
	userId := user.Id
	logger.Debug("", "User id", userId)

	chats, err := mc.Get[mapi.ChatsResponse](ctx, mc.ChatsUrl(userId))
	if err != nil {
		logger.Error("Error getting chats", "Error", err)
	}

	directChat := mapi.Chat{}
	for _, chatItem := range chats.Value {
		if chatItem.ChatType == types.OneOnOne && chatItem.LastMsg.LastMsgFrom.User.Id == userId {
			logger.Info("My Last Direct Msg Id: " + chatItem.LastMsg.Id)
			directChat = chatItem
			break
		}
	}

	msgRequest := mapi.MsgRequest{MsgBody: mapi.MsgBody{Content: msgContent, ContentType: "text"}}
	msgResponse, err := mc.Post[mapi.MsgResponse](ctx, mc.PostMsgUrl(userId, directChat.Id), msgRequest)
	if err != nil {
		logger.Error("Error posting msg", "Error", err)
	}
	logger.Debug("", "MsgResponse", msgResponse)
}
