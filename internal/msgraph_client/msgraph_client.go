package msgraph_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kanataidarov/tinkoff_voicekit/pkg/types"
	"io"
	"log"
	"net/http"
)

const baseUrl = "https://graph.microsoft.com/beta"

func ChatsUrl(userId string) string {
	return fmt.Sprintf(`%s/users/%s/chats?$expand=lastMessagePreview&$orderby=lastMessagePreview/createdDateTime+desc`, baseUrl, userId)
}

func PostMsgUrl(userId string, chatId string) string {
	return fmt.Sprintf(`%s/users/%s/chats/%s/messages`, baseUrl, userId, chatId)
}

func ProfileUrl() string {
	return fmt.Sprintf(`%s/me`, baseUrl)
}

func Get[T any](ctx context.Context, url string) (T, error) {
	ctxVals := ctx.Value("values").(types.CtxVals)
	cfg := ctxVals.Config
	logger := ctxVals.Logger

	var x T

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.Error("Error creating Get request", "Error", err)
		return x, err
	}
	req.Header.Add("Authorization", "Bearer "+cfg.MsGraph.Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error executing Get request. %v", err)
		return x, err
	}

	body, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		log.Printf("Error reading response body. %v", err)
		return x, err
	}

	return parseJson[T](body)
}

func parseJson[T any](s []byte) (T, error) {
	var r T
	if err := json.Unmarshal(s, &r); err != nil {
		return r, err
	}
	return r, nil
}

func Post[T any](ctx context.Context, url string, data any) (T, error) {
	ctxVals := ctx.Value("values").(types.CtxVals)
	cfg := ctxVals.Config
	logger := ctxVals.Logger

	var x T

	bts, err := toJson(data)
	if err != nil {
		logger.Error("Error marshalling json", "Error", err)
		return x, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bts))
	if err != nil {
		logger.Error("Error creating Post request", "Error", err)
		return x, err
	}
	req.Header.Add("Authorization", "Bearer "+cfg.MsGraph.Token)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("Error executing Post request", "Error", err)
		return x, err
	}

	body, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		logger.Error("Error reading response body", "Error", err)
		return x, err
	}

	return parseJson[T](body)
}

func toJson(data any) ([]byte, error) {
	return json.Marshal(data)
}
