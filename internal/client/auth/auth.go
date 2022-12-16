package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/axiomhq/cli/internal/client/auth/assets"
	"github.com/axiomhq/cli/internal/client/auth/pkce"
)

const (
	// Endpoints served on the authorization server.
	authPath  = "/oauth/authorize"
	tokenPath = "/oauth/token" //nolint:gosec // Sigh, this is not a hardcoded credential...
)

// LoginFunc is a function that is called with the URL the user has to visit in
// order to authenticate.
type LoginFunc func(ctx context.Context, loginURL string) error

// Login to the given Axiom deployment and retrieve a personal token in
// exchange. This will execute the OAuth2 Authorization Code Flow with Proof Key
// for Code Exchange (PKCE).
func Login(ctx context.Context, clientID, baseURL string, loginFunc LoginFunc) (string, error) {
	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return "", err
	}

	authURL, err := u.Parse(authPath)
	if err != nil {
		return "", err
	}

	tokenURL, err := u.Parse(tokenPath)
	if err != nil {
		return "", err
	}

	// Start a listener for the callback. We need to do this early in order to
	// construct the correct URL for the callback.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer lis.Close()

	redirectURL, err := url.Parse("http://" + lis.Addr().String())
	if err != nil {
		return "", err
	}

	config := &oauth2.Config{
		ClientID: clientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authURL.String(),
			TokenURL:  tokenURL.String(),
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: redirectURL.String(),
		Scopes:      []string{"*"},
	}

	// Create the PKCE Code Verifier and S256 Code Challenge.
	method := pkce.MethodS256
	codeVerifier, err := pkce.New()
	if err != nil {
		return "", err
	}

	// Generate a random state to prevent CSRF. It is hex-encoded to make it
	// URL-safe.
	stateBytes := make([]byte, 16)
	if _, err = rand.Read(stateBytes); err != nil {
		return "", err
	}
	state := hex.EncodeToString(stateBytes)

	// Setup the callback handler.
	var (
		token         *oauth2.Token      // Populated by callbackHandlerHf
		callbackErrCh = make(chan error) // Closed by callbackHandlerHf
		callbackOnce  sync.Once          // Ensures callbackHandlerHf is only called once
	)
	callbackHandlerHf := func(w http.ResponseWriter, r *http.Request) {
		callbackOnce.Do(func() {
			defer close(callbackErrCh)

			writeResponse := func(err error) {
				if err == nil {
					callbackErrCh <- assets.WriteStatusPage(w)
				} else {
					_ = assets.WriteStatusPageWithError(w, err)
					callbackErrCh <- err
				}
			}

			if r.Method != http.MethodGet {
				writeResponse(errors.New("invalid method"))
				return
			}

			// Make sure state matches.
			if r.FormValue("state") != state {
				writeResponse(errors.New("invalid state"))
				return
			}

			// Check if we have an error from the authorization server.
			if r.Form.Has("error") {
				serverErr := fmt.Errorf("oauth2 authorization error %q: %s", r.FormValue("error"), r.FormValue("error_description"))
				writeResponse(serverErr)
				return
			}

			code := r.FormValue("code")
			if code == "" {
				writeResponse(errors.New("missing authorization code"))
				return
			}

			var exchangeErr error
			if token, exchangeErr = config.Exchange(r.Context(), code, codeVerifier.AuthCodeOption()); exchangeErr != nil {
				writeResponse(exchangeErr)
				return
			}

			writeResponse(nil)
		})
	}

	// Start the HTTP server that listens for the callback.
	srv := http.Server{
		Addr:              lis.Addr().String(),
		Handler:           http.HandlerFunc(callbackHandlerHf),
		BaseContext:       func(net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: time.Second * 10,
	}
	defer srv.Close()

	srvErrCh := make(chan error)
	go func() {
		if serveErr := srv.Serve(lis); serveErr != nil && serveErr != http.ErrServerClosed {
			srvErrCh <- serveErr
		}
		close(srvErrCh)
	}()

	// Construct the login URL and call the login function provided by the
	// caller.
	loginURL := config.AuthCodeURL(state,
		codeVerifier.Challenge(method).AuthCodeOption(),
		method.AuthCodeOption(),
	)

	if err = loginFunc(ctx, loginURL); err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		close(callbackErrCh)
		return "", ctx.Err()
	case err = <-srvErrCh:
		close(callbackErrCh)
		return "", err
	case callbackErr := <-callbackErrCh:
		if shutdownErr := srv.Shutdown(ctx); callbackErr != nil {
			return "", callbackErr
		} else if shutdownErr != nil {
			return "", shutdownErr
		}
	}

	return token.AccessToken, nil
}
