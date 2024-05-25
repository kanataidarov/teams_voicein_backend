package main

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/kanataidarov/tinkoff_voicekit/pkg/args"
	"github.com/kanataidarov/tinkoff_voicekit/pkg/common"
	sttPb "github.com/kanataidarov/tinkoff_voicekit/pkg/tinkoff_voicekit/cloud/stt/v1"
)

func main() {
	opts := args.ParseRecognizeOptions()
	if opts == nil {
		os.Exit(1)
	}
	defer func(InputFile *os.File) {
		_ = InputFile.Close()
	}(opts.InputFile)

	var dataReader io.Reader
	if strings.HasSuffix(opts.InputFile.Name(), ".wav") {
		reader, err := common.OpenWavFormat(opts.InputFile, *opts.Encoding, *opts.NumChannels, *opts.Rate)
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

	// NOTE: in production code you should probably use context.WithCancel, context.WithDeadline or context.WithTimeout
	// instead of context.Background()
	result, err := client.Recognize(context.Background(), request)
	if err != nil {
		panic(err)
	}
	if common.PrettyPrintProtobuf(result) != nil {
		panic(err)
	}
}
