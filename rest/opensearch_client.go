package rest

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/opensearch-project/opensearch-go"
)

// OpenSearchClientConfig OpenSearch配置项。
type OpenSearchClientConfig struct {
	Protocol string
	Host     string
	Port     int
	Username string
	Password string `json:"-"`
}

// NewOpenSearchClient 初始化OpenSearchClient实例。
func NewOpenSearchClient(cfg OpenSearchClientConfig) *opensearch.Client {
	// 初始化http.Client
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second, // 连接超时时间
			KeepAlive: 60 * time.Second, // 保持长连接的时间
		}).DialContext, // 设置连接的参数
		MaxIdleConns:          1000,             // 最大空闲连接
		IdleConnTimeout:       60 * time.Second, // 空闲连接的超时时间
		ExpectContinueTimeout: 30 * time.Second, // 等待服务第一个响应的超时时间
		MaxIdleConnsPerHost:   500,              // 每个host保持的空闲连接数
		TLSHandshakeTimeout:   30 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// 连接地址
	address := fmt.Sprintf("%s://%s:%d", cfg.Protocol, cfg.Host, cfg.Port)

	// 重试
	retryBackoff := backoff.NewExponentialBackOff()

	// 初始化openSearch.Clisnt
	osc, _ := opensearch.NewClient(opensearch.Config{
		Addresses: []string{
			address,
		},
		Username:      cfg.Username,
		Password:      cfg.Password,
		Transport:     transport,
		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff: func(attempt int) time.Duration {
			if attempt == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},
		MaxRetries: 1,
	})

	CheckConnection(osc)
	return osc
}

// CheckConnection 检查连接状态。
func CheckConnection(osc *opensearch.Client) bool {
	res, err := osc.Info()
	if err != nil {
		log.Printf("new opensearch client failed: %v", err.Error())
		return false
	}

	if res.IsError() {
		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("new opensearch client failed: %v", err.Error())
			return false
		}
		log.Printf("new opensearch client failed: %s", string(resBytes))
		return false
	}
	return true
}
