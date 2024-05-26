package common

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/kanataidarov/tinkoff_voicekit/internal/config"
	"io"
	"os"
	"strings"

	"github.com/go-audio/wav"
	"github.com/kanataidarov/tinkoff_voicekit/pkg/auth"
	sttPb "github.com/kanataidarov/tinkoff_voicekit/pkg/tinkoff_voicekit/cloud/stt/v1"
	"github.com/tidwall/pretty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func AuthorizationKeys(cfg *config.Config) (auth.KeyPair, error) {
	apiKey := cfg.Tinkoff.ApiKey
	secretKey := cfg.Tinkoff.SecretKey

	if apiKey == "" || secretKey == "" {
		return auth.KeyPair{}, errors.New("no TINKOFF_API_KEY or TINKOFF_SECRET_KEY in config")
	}

	return auth.KeyPair{
		ApiKey:    apiKey,
		SecretKey: secretKey,
	}, nil
}

func isEndpointSecure(endpoint string) bool {
	parts := strings.Split(endpoint, ":")
	if len(parts) != 2 {
		return false
	}

	return parts[1] == "443"
}

func makeConnection(cfg *config.Config, creds *auth.JwtPerRPCCredentials) (*grpc.ClientConn, error) {
	if isEndpointSecure(cfg.Grpc.Endpoint) {
		var rootCAs *x509.CertPool
		if cfg.Grpc.CAFile != "" {
			pemServerCA, err := os.ReadFile(cfg.Grpc.CAFile)
			if err != nil {
				return nil, err
			}

			rootCAs = x509.NewCertPool()
			if !rootCAs.AppendCertsFromPEM(pemServerCA) {
				return nil, fmt.Errorf("failed to add server CA's certificate")
			}
		}

		return grpc.Dial(
			cfg.Grpc.Endpoint,
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
				RootCAs: rootCAs,
			})),
			grpc.WithPerRPCCredentials(creds),
		)
	}

	return grpc.Dial(
		cfg.Grpc.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

type SpeechToTextClient interface {
	sttPb.SpeechToTextClient
	Close() error
}

type speechToTextClient struct {
	sttPb.SpeechToTextClient
	conn *grpc.ClientConn
}

func (client *speechToTextClient) Close() error {
	return client.conn.Close()
}

func NewSttClient(cfg *config.Config) (SpeechToTextClient, error) {
	keyPair, err := AuthorizationKeys(cfg)
	if err != nil {
		return nil, err
	}

	perRPCCredentials := auth.NewJwtPerRPCCredentials(keyPair, "test_issuer", "test_subject")
	connection, err := makeConnection(cfg, perRPCCredentials)

	return &speechToTextClient{
		SpeechToTextClient: sttPb.NewSpeechToTextClient(connection),
		conn:               connection,
	}, err
}

func OpenWavFormat(file *os.File, expectedEncoding string, expectedNumChannels int, expectedRate int) (io.Reader, error) {
	wavDecoder := wav.NewDecoder(file)
	wavDecoder.ReadInfo()

	encodingAudioFormat := map[string]uint16{
		"LINEAR16": 0x0001,
		"ALAW":     0x0006,
		"MULAW":    0x0007,
	}
	encodingBitDepth := map[string]uint16{
		"LINEAR16": 16,
		"ALAW":     8,
		"MULAW":    8,
	}
	if encodingAudioFormat[expectedEncoding] != wavDecoder.WavAudioFormat {
		return nil, fmt.Errorf("bad audio format, expected %s, found %v", expectedEncoding, wavDecoder.WavAudioFormat)
	}
	if encodingBitDepth[expectedEncoding] != wavDecoder.BitDepth {
		return nil, fmt.Errorf("expected bid depth %v, but found %v", encodingBitDepth[expectedEncoding], wavDecoder.BitDepth)
	}
	if expectedNumChannels != int(wavDecoder.NumChans) {
		return nil, fmt.Errorf("expected %v channels, but found %v", expectedNumChannels, wavDecoder.NumChans)
	}
	if expectedRate != int(wavDecoder.SampleRate) {
		return nil, fmt.Errorf("expected %v sample rate, but found %v", expectedRate, wavDecoder.SampleRate)
	}

	if wavDecoder.FwdToPCM() != nil {
		return nil, fmt.Errorf("forwarding to data chunk failed")
	}
	return wavDecoder.PCMChunk.R, nil
}

func PrettyPrintProtobuf(message proto.Message) error {
	marshaller := protojson.MarshalOptions{
		Indent: "  ",
	}
	jsonMessage, err := marshaller.Marshal(message)
	if err != nil {
		return err
	}

	fmt.Println(string(pretty.Color(jsonMessage, pretty.TerminalStyle)))
	return nil
}
