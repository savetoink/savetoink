// Package server provides HTTP routing and middleware for the savetoink application.
package server

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/shaftoe/savetoink/backend/lib/config"
	"github.com/shaftoe/savetoink/backend/lib/consts"
	"github.com/shaftoe/savetoink/backend/lib/logging"
)

func newCorsMiddleware(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.CorsAllowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", cfg.CorsAllowOrigin)
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
				w.Header().Set("Access-Control-Allow-Credentials", "true")

				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := ""

		if lc, ok := lambdacontext.FromContext(r.Context()); ok {
			requestID = lc.AwsRequestID
		}

		if requestID == "" {
			requestID = r.Header.Get("X-Request-ID")
		}

		if requestID == "" {
			requestID = r.Header.Get("x-amzn-request-id")
		}

		if requestID == "" {
			requestID = generateRequestID()
		}

		ctx := logging.WithRequestID(r.Context(), requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestID() string {
	return strings.ReplaceAll(time.Now().Format(consts.RequestIDFormat), ".", "")
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// processorInfoMiddleware adds a "processed_by" log attribute to the request context
// indicating the lambda function that processes the article creation.
func processorInfoMiddleware(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.ProcessArticleLambda != "" {
				logging.AddLogAttr(r.Context(), slog.String("processed_by", cfg.ProcessArticleLambda))
			}
			next.ServeHTTP(w, r)
		})
	}
}
