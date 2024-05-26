package msgraph_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func ChatsUrl() string {
	return `https://graph.microsoft.com/beta/users/a2619b8a-6136-43df-9910-692af6264c0b/chats?$expand=lastMessagePreview&$orderby=lastMessagePreview/createdDateTime%20desc`

}

func Get[T any](ctx context.Context, url string) (T, error) {
	var x T

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("Error creating Get request. %v", err)
		return x, err
	}
	req.Header = http.Header{"Authorization": []string{"Bearer " + "eyJ0eXAiOiJKV1QiLCJub25jZSI6IktXNGhER3VDRGs2eDI2ZExDdWMtZUItSS1Fb2lxeGpVdnpTcTlNdGVFOFEiLCJhbGciOiJSUzI1NiIsIng1dCI6IkwxS2ZLRklfam5YYndXYzIyeFp4dzFzVUhIMCIsImtpZCI6IkwxS2ZLRklfam5YYndXYzIyeFp4dzFzVUhIMCJ9.eyJhdWQiOiIwMDAwMDAwMy0wMDAwLTAwMDAtYzAwMC0wMDAwMDAwMDAwMDAiLCJpc3MiOiJodHRwczovL3N0cy53aW5kb3dzLm5ldC9kMDUyOTVkMy1hODA4LTRlMWItYjJlYi1kMTVlNjk5NTY5NzYvIiwiaWF0IjoxNzE2NzE4ODk0LCJuYmYiOjE3MTY3MTg4OTQsImV4cCI6MTcxNjgwNTU5NCwiYWNjdCI6MCwiYWNyIjoiMSIsImFpbyI6IkFWUUFxLzhXQUFBQU11dGRXRzBQSlJtTzNRTjFpRk5jY0IwR2VJS0Fybi9xV3UxeVo5UWkvYVU1dk4xQk5iQTNHR2YvbWdLT3Bqem1BQXBoTENDNGViVEdBaCtzeWdKSlltbEpjRkhrTzdXTGxlRDNWUzBFTEVnPSIsImFtciI6WyJwd2QiLCJtZmEiXSwiYXBwX2Rpc3BsYXluYW1lIjoiR3JhcGggRXhwbG9yZXIiLCJhcHBpZCI6ImRlOGJjOGI1LWQ5ZjktNDhiMS1hOGFkLWI3NDhkYTcyNTA2NCIsImFwcGlkYWNyIjoiMCIsImZhbWlseV9uYW1lIjoi0JDQudC00LDRgNC-0LIiLCJnaXZlbl9uYW1lIjoi0JrQsNC90LDRgiIsImlkdHlwIjoidXNlciIsImlwYWRkciI6IjIuNzIuNDAuMTEiLCJuYW1lIjoiQWlkYXJvdiBLYW5hdCIsIm9pZCI6ImEyNjE5YjhhLTYxMzYtNDNkZi05OTEwLTY5MmFmNjI2NGMwYiIsIm9ucHJlbV9zaWQiOiJTLTEtNS0yMS0zNjgwODIyNjAtODE4OTAxNDEwLTE3NjI5NDIxNTctMTc0NjgwIiwicGxhdGYiOiI1IiwicHVpZCI6IjEwMDMyMDAxMTAzOTA2MTgiLCJyaCI6IjAuQVYwQTA1VlMwQWlvRzA2eTY5RmVhWlZwZGdNQUFBQUFBQUFBd0FBQUFBQUFBQUJkQUw0LiIsInNjcCI6IkFuYWx5dGljcy5SZWFkIENoYXQuUmVhZCBDaGF0LlJlYWRCYXNpYyBEaXJlY3RvcnkuUmVhZFdyaXRlLkFsbCBNYWlsLlJlYWRXcml0ZSBvcGVuaWQgcHJvZmlsZSBSZXBvcnRzLlJlYWQuQWxsIFVzZXIuUmVhZCBlbWFpbCBDaGF0LlJlYWRXcml0ZSIsInNpZ25pbl9zdGF0ZSI6WyJrbXNpIl0sInN1YiI6IkI4dHV1LXQweU5KaU9mYXFqTXVYSGt2dkx2UFdxTDlDZmtULVFtZGJ4QmsiLCJ0ZW5hbnRfcmVnaW9uX3Njb3BlIjoiRVUiLCJ0aWQiOiJkMDUyOTVkMy1hODA4LTRlMWItYjJlYi1kMTVlNjk5NTY5NzYiLCJ1bmlxdWVfbmFtZSI6IktBaWRhcm92QGJlZWxpbmUua3oiLCJ1cG4iOiJLQWlkYXJvdkBiZWVsaW5lLmt6IiwidXRpIjoiUlFyQmd6c0ZOMGk2YVdRSl9FOFJBQSIsInZlciI6IjEuMCIsIndpZHMiOlsiYjc5ZmJmNGQtM2VmOS00Njg5LTgxNDMtNzZiMTk0ZTg1NTA5Il0sInhtc19jYyI6WyJDUDEiXSwieG1zX3NzbSI6IjEiLCJ4bXNfc3QiOnsic3ViIjoiYUlFcHRselo0SGtCU1ptYzNTeWNfLVJmNFF2LWhlcXd2Z05UM29Gb29YRSJ9LCJ4bXNfdGNkdCI6MTU4NzU0OTc5Mn0.MczgDHC3hwijmQbkUJtwYrU7Mm9vhhq7arZBSoiSxWeFG5KrGHJk7O0pFJt4-iK0OwDCbcmAQq839jrylFYqx7fV8x5NMm-NbepFKjqrGw1y5t8ltZY8DzfXbZO1zTg_z78RtPfbnXBmcyZrdkNoZmsCa6hSZqDquTw1mLKazwvX1Ch6mif7zvXWTViKl9X1BnRjGtJ07GS9YVo5XJlJKAZEgF2dVgn_o93qNoml2tPmlgWQKHa1DB2HMECuIuBcjhq-E06iyN7CKW61B9ouG8bRN9FbmkcJsGq1zLTtcvWFYANSU0RJbX0ls4QVomWj_0LleuouEWKnx7w9VPXmzg"}}

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

func PostMsgUrl(chatId string) string {
	return fmt.Sprintf(`https://graph.microsoft.com/beta/users/a2619b8a-6136-43df-9910-692af6264c0b/chats/%s/messages`, chatId)
}

func Post[T any](ctx context.Context, url string, data any) (T, error) {
	var x T

	bts, err := toJson(data)
	if err != nil {
		log.Printf("Error marshalling json. %v", err)
		return x, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bts))
	if err != nil {
		log.Printf("Error creating Post request. %v", err)
		return x, err
	}
	req.Header.Add("Authorization", "Bearer "+"eyJ0eXAiOiJKV1QiLCJub25jZSI6IktXNGhER3VDRGs2eDI2ZExDdWMtZUItSS1Fb2lxeGpVdnpTcTlNdGVFOFEiLCJhbGciOiJSUzI1NiIsIng1dCI6IkwxS2ZLRklfam5YYndXYzIyeFp4dzFzVUhIMCIsImtpZCI6IkwxS2ZLRklfam5YYndXYzIyeFp4dzFzVUhIMCJ9.eyJhdWQiOiIwMDAwMDAwMy0wMDAwLTAwMDAtYzAwMC0wMDAwMDAwMDAwMDAiLCJpc3MiOiJodHRwczovL3N0cy53aW5kb3dzLm5ldC9kMDUyOTVkMy1hODA4LTRlMWItYjJlYi1kMTVlNjk5NTY5NzYvIiwiaWF0IjoxNzE2NzE4ODk0LCJuYmYiOjE3MTY3MTg4OTQsImV4cCI6MTcxNjgwNTU5NCwiYWNjdCI6MCwiYWNyIjoiMSIsImFpbyI6IkFWUUFxLzhXQUFBQU11dGRXRzBQSlJtTzNRTjFpRk5jY0IwR2VJS0Fybi9xV3UxeVo5UWkvYVU1dk4xQk5iQTNHR2YvbWdLT3Bqem1BQXBoTENDNGViVEdBaCtzeWdKSlltbEpjRkhrTzdXTGxlRDNWUzBFTEVnPSIsImFtciI6WyJwd2QiLCJtZmEiXSwiYXBwX2Rpc3BsYXluYW1lIjoiR3JhcGggRXhwbG9yZXIiLCJhcHBpZCI6ImRlOGJjOGI1LWQ5ZjktNDhiMS1hOGFkLWI3NDhkYTcyNTA2NCIsImFwcGlkYWNyIjoiMCIsImZhbWlseV9uYW1lIjoi0JDQudC00LDRgNC-0LIiLCJnaXZlbl9uYW1lIjoi0JrQsNC90LDRgiIsImlkdHlwIjoidXNlciIsImlwYWRkciI6IjIuNzIuNDAuMTEiLCJuYW1lIjoiQWlkYXJvdiBLYW5hdCIsIm9pZCI6ImEyNjE5YjhhLTYxMzYtNDNkZi05OTEwLTY5MmFmNjI2NGMwYiIsIm9ucHJlbV9zaWQiOiJTLTEtNS0yMS0zNjgwODIyNjAtODE4OTAxNDEwLTE3NjI5NDIxNTctMTc0NjgwIiwicGxhdGYiOiI1IiwicHVpZCI6IjEwMDMyMDAxMTAzOTA2MTgiLCJyaCI6IjAuQVYwQTA1VlMwQWlvRzA2eTY5RmVhWlZwZGdNQUFBQUFBQUFBd0FBQUFBQUFBQUJkQUw0LiIsInNjcCI6IkFuYWx5dGljcy5SZWFkIENoYXQuUmVhZCBDaGF0LlJlYWRCYXNpYyBEaXJlY3RvcnkuUmVhZFdyaXRlLkFsbCBNYWlsLlJlYWRXcml0ZSBvcGVuaWQgcHJvZmlsZSBSZXBvcnRzLlJlYWQuQWxsIFVzZXIuUmVhZCBlbWFpbCBDaGF0LlJlYWRXcml0ZSIsInNpZ25pbl9zdGF0ZSI6WyJrbXNpIl0sInN1YiI6IkI4dHV1LXQweU5KaU9mYXFqTXVYSGt2dkx2UFdxTDlDZmtULVFtZGJ4QmsiLCJ0ZW5hbnRfcmVnaW9uX3Njb3BlIjoiRVUiLCJ0aWQiOiJkMDUyOTVkMy1hODA4LTRlMWItYjJlYi1kMTVlNjk5NTY5NzYiLCJ1bmlxdWVfbmFtZSI6IktBaWRhcm92QGJlZWxpbmUua3oiLCJ1cG4iOiJLQWlkYXJvdkBiZWVsaW5lLmt6IiwidXRpIjoiUlFyQmd6c0ZOMGk2YVdRSl9FOFJBQSIsInZlciI6IjEuMCIsIndpZHMiOlsiYjc5ZmJmNGQtM2VmOS00Njg5LTgxNDMtNzZiMTk0ZTg1NTA5Il0sInhtc19jYyI6WyJDUDEiXSwieG1zX3NzbSI6IjEiLCJ4bXNfc3QiOnsic3ViIjoiYUlFcHRselo0SGtCU1ptYzNTeWNfLVJmNFF2LWhlcXd2Z05UM29Gb29YRSJ9LCJ4bXNfdGNkdCI6MTU4NzU0OTc5Mn0.MczgDHC3hwijmQbkUJtwYrU7Mm9vhhq7arZBSoiSxWeFG5KrGHJk7O0pFJt4-iK0OwDCbcmAQq839jrylFYqx7fV8x5NMm-NbepFKjqrGw1y5t8ltZY8DzfXbZO1zTg_z78RtPfbnXBmcyZrdkNoZmsCa6hSZqDquTw1mLKazwvX1Ch6mif7zvXWTViKl9X1BnRjGtJ07GS9YVo5XJlJKAZEgF2dVgn_o93qNoml2tPmlgWQKHa1DB2HMECuIuBcjhq-E06iyN7CKW61B9ouG8bRN9FbmkcJsGq1zLTtcvWFYANSU0RJbX0ls4QVomWj_0LleuouEWKnx7w9VPXmzg")
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error executing Post request. %v", err)
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

func toJson(data any) ([]byte, error) {
	return json.Marshal(data)
}
