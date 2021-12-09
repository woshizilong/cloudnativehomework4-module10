package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
)

// 测试步骤中加入 `Report(apitest.SequenceDiagram())` 后可以在测试时生成时序图

func TestUnit_healthHandler(t *testing.T) {
	defer leaktest.Check(t)()
	apitest.New().
		HandlerFunc(healthHandler).
		Get("/healthz").
		Expect(t).
		Body(``).
		Status(http.StatusOK).
		End()
}

func TestUnit_readyHandler(t *testing.T) {
	defer leaktest.Check(t)()

	assert := assert.New(t)

	apitest.New().HandlerFunc(readyHandler).
		Get("/readyz").
		Expect(t).
		Body(``).
		Assert(func(res *http.Response, req *http.Request) error {
			assert.True(res.StatusCode >= 500)
			return nil
		}).
		End()
}

func TestUnit_infoHandler(t *testing.T) {
	defer leaktest.Check(t)()
	apitest.New().HandlerFunc(infoHandler).
		Get("/info").
		Expect(t).
		Status(http.StatusOK).
		End()
}

func TestUnit_runHandler(t *testing.T) {
	defer leaktest.Check(t)()
	apitest.New().HandlerFunc(runHandler).
		Get("/run").
		Expect(t).
		Status(http.StatusNonAuthoritativeInfo).
		End()
}

func TestUnit_readyNeedTime(t *testing.T) {
	finish := make(chan struct{})

	// 开启真实的服务
	go func() {
		if err := Start(context.Background()); err != nil {
			panic(err)
		}
	}()

	// 测试
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
			}
		}()

		cookieJar, _ := cookiejar.New(nil)
		cli := &http.Client{
			Timeout: time.Second * 5,
			Jar:     cookieJar,
		}

		time.Sleep(5 * time.Second)

		apitest.New().
			EnableNetworking(cli).
			Get("http://localhost:8000/readyz").
			Expect(t).
			Status(500).
			End()

		time.Sleep(5 * time.Second)

		apitest.New().
			EnableNetworking(cli).
			Get("http://localhost:8000/readyz").
			Expect(t).
			Status(200).
			End()

		finish <- struct{}{}
	}()

	<-finish
}
