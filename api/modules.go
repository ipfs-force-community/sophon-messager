package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
	"go.uber.org/fx"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type JsonRpcRequest struct {
	// common
	Jsonrpc string            `json:"jsonrpc"`
	ID      int64             `json:"id,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`

	// request
	Method string        `json:"method,omitempty"`
	Params []interface{} `json:"params,omitempty"`
}

var _ io.ReadCloser = (*CloserReader)(nil)

type CloserReader struct {
	reader io.Reader
}

func (c *CloserReader) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}

func (c *CloserReader) Close() error {
	return nil
}

type RewriteJsonRpcToRustful struct {
	*gin.Engine
}

func (r *RewriteJsonRpcToRustful) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost && req.URL.Path == "/rpc/v0" {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(503)
			w.Write([]byte("failed to read json rpc body"))
			return
		}

		jsonReq := &JsonRpcRequest{}
		err = json.Unmarshal(body, jsonReq)
		if err != nil {
			w.WriteHeader(503)
			w.Write([]byte("failed to unmarshal json rpc body"))
			return
		}
		methodSeq := strings.Split(jsonReq.Method, ".")
		//	methodPath := strings.Join(strings.Split(jsonReq.Method, "."), "/")
		newRequestUrl := req.RequestURI + "/" + methodSeq[len(methodSeq)-1] + "/" + strconv.FormatInt(jsonReq.ID, 10)
		newUrl, err := url.Parse(newRequestUrl)
		if err != nil {
			w.WriteHeader(503)
			w.Write([]byte("failed to parser new url"))
			return
		}
		req.URL = newUrl
		req.RequestURI = newRequestUrl
		params, _ := json.Marshal(jsonReq.Params)
		req.Body = &CloserReader{bytes.NewBuffer(params)}
	}

	r.Engine.ServeHTTP(w, req)
}

func UseMiddleware(log *logrus.Logger, r *gin.Engine) error {
	r.Use(ginlogrus.Logger(log), gin.Recovery())
	//r.Use(middleware.RewriteJsonRpcMiddleware)
	return nil
}

func InitRouter() *gin.Engine {
	return gin.New()

	/*	// Simple group: v1
		v1 := router.Group("rpc/v1")
		{
			v1.POST("PushMessage", func)
		}*/
}

func RunAPI(lc fx.Lifecycle, r *gin.Engine, lst net.Listener) error {
	skipContextPathRouter := &RewriteJsonRpcToRustful{
		Engine: r,
	}

	handler := http.NewServeMux()
	handler.Handle("/", skipContextPathRouter)
	apiserv := &http.Server{
		Handler: handler,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				apiserv.Serve(lst)
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return lst.Close()
		},
	})
	return nil
}
