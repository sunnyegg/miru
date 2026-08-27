package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sunnyegg/miru/internal/anilist"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) AnilistStatus() (AnilistStatus, error) {
	if err := a.ready(); err != nil {
		return AnilistStatus{}, err
	}
	token, err := a.tokens.Get()
	if err != nil {
		return AnilistStatus{Connected: false}, nil
	}
	client, err := a.newAnilist(token)
	if err != nil {
		return AnilistStatus{Connected: false}, nil
	}
	name, err := client.ViewerName()
	if err != nil {
		return AnilistStatus{Connected: false}, nil
	}
	return AnilistStatus{Connected: true, Username: name}, nil
}

func (a *App) OpenAnilistLogin() error {
	if err := a.ready(); err != nil {
		return err
	}
	settings, err := a.loadSettings()
	if err != nil {
		return err
	}
	loginURL, err := anilist.LoginURL(settings.AnilistClientId)
	if err != nil {
		return err
	}
	if err := a.startLoginServer(); err != nil {
		return err
	}
	runtime.BrowserOpenURL(a.ctx, loginURL)
	return nil
}

func (a *App) startLoginServer() error {
	a.loginMu.Lock()
	defer a.loginMu.Unlock()
	if a.loginSrv != nil {
		return nil
	}

	ln, err := net.Listen("tcp", anilist.ListenAddr)
	if err != nil {
		return fmt.Errorf("login port %s is in use; try again after closing the other process: %w", anilist.ListenAddr, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	settings, err := a.loadSettings()
	if err != nil {
		_ = ln.Close()
		return err
	}
	secret := envTrim("ANILIST_CLIENT_SECRET")
	if secret == "" {
		secret, _ = a.store.GetSetting("anilist_client_secret")
	}
	clientID := settings.AnilistClientId
	mux := anilist.NewMux(anilist.MuxConfig{
		ExchangeCode: func(code string) (string, error) {
			httpClient, err := a.networkHTTPClient()
			if err != nil {
				return "", err
			}
			return anilist.ExchangeCode(httpClient, anilist.TokenURL, clientID, secret, code)
		},
		OnToken: func(token string) error {
			if err := a.SaveAnilistToken(token); err != nil {
				return err
			}
			runtime.EventsEmit(a.ctx, "anilist:connected", true)
			cancel()
			return nil
		},
	})
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	a.loginSrv = srv
	a.loginCancel = cancel

	go func() {
		_ = srv.Serve(ln)
	}()
	go func() {
		<-ctx.Done()
		a.stopLoginServer()
	}()
	return nil
}

func (a *App) stopLoginServer() {
	a.loginMu.Lock()
	srv := a.loginSrv
	cancel := a.loginCancel
	a.loginSrv = nil
	a.loginCancel = nil
	a.loginMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if srv != nil {
		ctx, done := context.WithTimeout(context.Background(), 2*time.Second)
		defer done()
		_ = srv.Shutdown(ctx)
	}
}

func (a *App) SaveAnilistToken(token string) error {
	if err := a.ready(); err != nil {
		return err
	}
	token = strings.TrimSpace(token)
	client, err := a.newAnilist(token)
	if err != nil {
		return err
	}
	name, err := client.ViewerName()
	if err != nil {
		return err
	}
	if err := a.tokens.Set(token); err != nil {
		return err
	}
	runtime.LogInfo(a.ctx, "AniList connected as "+name)
	return nil
}

func (a *App) LogoutAnilist() error {
	if err := a.ready(); err != nil {
		return err
	}
	_ = a.store.DeleteAPICache(watchingCacheKey)
	return a.tokens.Delete()
}
