package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/venus-messager/service"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"net/http"
	"reflect"
	"strconv"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func SetupController(router *gin.Engine, sMap service.ServiceMap, log *logrus.Logger) error {
	v1 := router.Group("rpc/v0")
	return registerController(v1, sMap, log, reflect.TypeOf(Message{}))
}

type JsonRpcResponse struct {
	ID     int64       `json:"id,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

func registerController(v1 *gin.RouterGroup, sMap service.ServiceMap, log *logrus.Logger, controllerT reflect.Type) error {
	methodNumber := controllerT.NumMethod()

	for i := 0; i < methodNumber; i++ {
		method := controllerT.Method(i)
		methodName := method.Name
		inParamsNumber := method.Type.NumIn()
		resultParamsNumber := method.Type.NumOut()
		if resultParamsNumber != 2 {
			return xerrors.Errorf("controllerT method must has 2 return as result, first one is value and second is error")
		}

		if !method.Type.Out(1).Implements(errorInterface) {
			return xerrors.Errorf("second result must be a error")
		}

		//{"jsonrpc": "2.0", "result": -19, "id": 2}
		v1.Handle(http.MethodPost, methodName+"/:id", func(c *gin.Context) {
			idStr := c.Param("id")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				c.String(http.StatusServiceUnavailable, "error id number in request body")
				return
			}

			paramsDecoder := json.NewDecoder(c.Request.Body)
			_, err = paramsDecoder.Token()
			if err != nil {
				c.String(http.StatusServiceUnavailable, "body not a json array")
				return
			}

			//controller
			controller := reflect.New(controllerT).Elem()
			//todo how to inject filed values?
			baseController := BaseController{
				//Context: c,
				Logger: log,
			}

			controller.Field(0).Set(reflect.ValueOf(baseController))
			for i := 1; i < controller.NumField(); i++ {
				if val, ok := sMap[controller.Field(i).Type()]; ok {
					controller.Field(i).Set(reflect.ValueOf(val))
				}
			}

			var inParams []reflect.Value
			inParams = append(inParams, controller)
			inParams = append(inParams, reflect.ValueOf(c.Request.Context()))
			for i := 2; i < inParamsNumber; i++ {
				argT := method.Type.In(i)
				arg := reflect.New(argT)
				err := paramsDecoder.Decode(arg.Interface())
				if err != nil {
					c.JSON(http.StatusServiceUnavailable, JsonRpcResponse{
						ID:     id,
						Result: fmt.Sprintf("expect type %t, but failed %v", argT, err),
					})
				}
				inParams = append(inParams, arg.Elem())
			}

			out := method.Func.Call(inParams)

			//result
			if out[1].IsNil() {
				c.JSON(http.StatusOK, JsonRpcResponse{
					ID:     id,
					Result: out[0].Interface(),
				})
			} else {
				err := out[1].Interface()
				c.JSON(http.StatusOK, JsonRpcResponse{
					ID:     id,
					Result: err.(error).Error(),
				})
			}
		})
	}
	return nil
}
