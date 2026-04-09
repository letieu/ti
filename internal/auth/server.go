package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

type CallbackResult struct {
	Code  string
	State string
}

type CallbackServer struct {
	server   *http.Server
	resultCh chan *CallbackResult
	listener net.Listener
}

func StartCallbackServer() (*CallbackServer, error) {
	resultCh := make(chan *CallbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth-callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		errParam := q.Get("error")
		code := q.Get("code")
		state := q.Get("state")

		if errParam != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, oauthErrorHTML("Google authentication did not complete.", "Error: "+errParam))
			return
		}
		if code != "" && state != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, oauthSuccessHTML("Google authentication completed. You can close this window."))
			resultCh <- &CallbackResult{Code: code, State: state}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, oauthErrorHTML("Missing code or state parameter.", ""))
	})

	listener, err := net.Listen("tcp", "127.0.0.1:51121")
	if err != nil {
		return nil, err
	}

	srv := &http.Server{Handler: mux}
	go srv.Serve(listener)

	return &CallbackServer{
		server:   srv,
		resultCh: resultCh,
		listener: listener,
	}, nil
}

// WaitForCode blocks until a code+state is received or the context is cancelled.
func (s *CallbackServer) WaitForCode(ctx context.Context) (*CallbackResult, error) {
	select {
	case result := <-s.resultCh:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Shutdown stops the server.
func (s *CallbackServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func oauthErrorHTML(message, detail string) string {
	return fmt.Sprintf(`<html><body><h1>Error</h1><p>%s</p><p>%s</p></body></html>`, message, detail)
}

func oauthSuccessHTML(message string) string {
	return fmt.Sprintf(`<html><body><h1>Success</h1><p>%s</p></body></html>`, message)
}
