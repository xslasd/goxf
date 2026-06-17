package sse

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSSE(t *testing.T) {
	// 1. 创建具有较短心跳周期的 Hub 以便于测试
	hub, err := NewHub(
		WithKeepAliveTime(100 * time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Close()

	// 2. 启动本地 HTTP 测试服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		err := hub.Upgrade(r.Context(), w, r, id)
		if err != nil {
			t.Logf("upgrade error: %v", err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// 辅助协程：读取 HTTP 响应流中的事件行
	readEvents := func(ctx context.Context, url string) (chan string, chan error) {
		out := make(chan string, 100)
		errs := make(chan error, 1)
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		go func() {
			defer close(out)
			defer close(errs)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				errs <- err
				return
			}
			defer resp.Body.Close()

			reader := bufio.NewReader(resp.Body)
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					errs <- err
					return
				}
				out <- line
			}
		}()
		return out, errs
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. 开启客户端 1 并注册 ID: cli1
	cli1Chan, cli1Err := readEvents(ctx, server.URL+"/sse?id=cli1")
	time.Sleep(50 * time.Millisecond) // 等待注册成功

	if !hub.IsClientOnline("cli1") {
		t.Fatal("client 1 should be online")
	}

	// 4. 开启客户端 2 并注册 ID: cli2
	cli2Chan, cli2Err := readEvents(ctx, server.URL+"/sse?id=cli2")
	time.Sleep(50 * time.Millisecond) // 等待注册成功

	if !hub.IsClientOnline("cli2") {
		t.Fatal("client 2 should be online")
	}

	// 5. 测试 Forward: 从 1 转发给 2（普通无名消息）
	c1, _ := hub.GetClient("cli1")
	c1.Forward("cli2", []byte("hello cli2"))

	var lines []string
	timeout := time.After(200 * time.Millisecond)
loop1:
	for {
		select {
		case line := <-cli2Chan:
			lines = append(lines, strings.TrimSpace(line))
			if len(lines) >= 2 {
				break loop1
			}
		case err := <-cli2Err:
			t.Fatalf("cli2 stream error: %v", err)
		case <-timeout:
			t.Fatalf("timeout waiting for forward message, received: %v", lines)
		}
	}
	if lines[0] != "data: hello cli2" || lines[1] != "" {
		t.Errorf("unexpected forward message structure: %v", lines)
	}

	// 6. 测试 BroadcastEvent: 1 广播 "chat" 事件给所有人
	c1.BroadcastEvent("chat", []byte("hello all"))
	lines = nil
	timeout = time.After(200 * time.Millisecond)
loop2:
	for {
		select {
		case line := <-cli2Chan:
			lines = append(lines, strings.TrimSpace(line))
			if len(lines) >= 3 { // 期待 event 行、data 行和结尾空行
				break loop2
			}
		case err := <-cli2Err:
			t.Fatalf("cli2 stream error: %v", err)
		case <-timeout:
			t.Fatalf("timeout waiting for broadcast message, received: %v", lines)
		}
	}
	if lines[0] != "event: chat" || lines[1] != "data: hello all" || lines[2] != "" {
		t.Errorf("unexpected broadcast message structure: %v", lines)
	}

	// 7. 测试心跳注释包 (: keepalive)
	lines = nil
	timeout = time.After(300 * time.Millisecond)
loop3:
	for {
		select {
		case line := <-cli2Chan:
			if strings.HasPrefix(line, ":") {
				lines = append(lines, strings.TrimSpace(line))
				break loop3
			}
		case err := <-cli2Err:
			t.Fatalf("cli2 stream error: %v", err)
		case <-timeout:
			t.Fatalf("timeout waiting for keepalive comment")
		}
	}
	if lines[0] != ": keepalive" {
		t.Errorf("unexpected keepalive comment: %s", lines[0])
	}

	// 8. 测试踢人机制：新建立一个相同 ID: cli1 的连接
	_, _ = readEvents(ctx, server.URL+"/sse?id=cli1")
	time.Sleep(50 * time.Millisecond)

	// 旧的 cli1 连接应该被强制断开（即 cli1Chan 会被关闭或返回 EOF 错误）
	select {
	case _, ok := <-cli1Chan:
		if ok {
			// 可能读到最后关闭的零值，也可能直接关闭
			select {
			case _, ok = <-cli1Chan:
				if ok {
					t.Error("old cli1 connection should be closed after conflict registration")
				}
			case <-time.After(100 * time.Millisecond):
			}
		}
	case <-cli1Err:
		// 触发错误退出也符合预期
	case <-time.After(200 * time.Millisecond):
		t.Error("old cli1 connection was not closed in time")
	}
}
