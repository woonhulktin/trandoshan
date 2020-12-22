package crawler

import (
	"bytes"
	"github.com/creekorful/trandoshan/internal/crawler/http_mock"
	"github.com/creekorful/trandoshan/internal/event"
	"github.com/creekorful/trandoshan/internal/event_mock"
	"github.com/golang/mock/gomock"
	"strings"
	"testing"
)

func TestCrawlURLForbiddenContentType(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	httpClientMock := http_mock.NewMockClient(mockCtrl)
	url := "https://example.onion"
	allowedContentTypes := []string{"text/plain"}

	httpResponseMock := http_mock.NewMockResponse(mockCtrl)
	httpResponseMock.EXPECT().Headers().Return(map[string]string{"Content-Type": "image/png"})

	httpClientMock.EXPECT().Get(url).Return(httpResponseMock, nil)

	body, headers, err := crawURL(httpClientMock, url, allowedContentTypes)
	if body != "" || headers != nil || err == nil {
		t.Fail()
	}
}

func TestCrawlURLSameContentType(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	httpClientMock := http_mock.NewMockClient(mockCtrl)
	url := "https://example.onion"
	allowedContentTypes := []string{"text/plain"}

	httpResponseMock := http_mock.NewMockResponse(mockCtrl)
	httpResponseMock.EXPECT().Headers().Times(2).Return(map[string]string{"Content-Type": "text/plain"})
	httpResponseMock.EXPECT().Body().Return(strings.NewReader("Hello"))

	httpClientMock.EXPECT().Get(url).Return(httpResponseMock, nil)

	body, headers, err := crawURL(httpClientMock, url, allowedContentTypes)
	if err != nil {
		t.Fail()
	}
	if body != "Hello" {
		t.Fail()
	}
	if len(headers) != 1 {
		t.Fail()
	}
	if headers["Content-Type"] != "text/plain" {
		t.Fail()
	}
}

func TestCrawlURLNoContentTypeFiltering(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	httpClientMock := http_mock.NewMockClient(mockCtrl)
	url := "https://example.onion"
	allowedContentTypes := []string{""}

	httpResponseMock := http_mock.NewMockResponse(mockCtrl)
	httpResponseMock.EXPECT().Headers().Times(2).Return(map[string]string{"Content-Type": "text/plain"})
	httpResponseMock.EXPECT().Body().Return(strings.NewReader("Hello"))

	httpClientMock.EXPECT().Get(url).Return(httpResponseMock, nil)

	body, headers, err := crawURL(httpClientMock, url, allowedContentTypes)
	if err != nil {
		t.Fail()
	}
	if body != "Hello" {
		t.Fail()
	}
	if len(headers) != 1 {
		t.Fail()
	}
	if headers["Content-Type"] != "text/plain" {
		t.Fail()
	}
}

func TestHandleMessage(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	subscriberMock := event_mock.NewMockSubscriber(mockCtrl)
	httpClientMock := http_mock.NewMockClient(mockCtrl)
	httpResponseMock := http_mock.NewMockResponse(mockCtrl)

	msg := bytes.NewReader(nil)
	subscriberMock.EXPECT().
		Read(msg, &event.NewURLEvent{}).
		SetArg(1, event.NewURLEvent{URL: "https://example.onion/image.png?id=12&test=2"}).
		Return(nil)

	httpResponseMock.EXPECT().Headers().Times(2).Return(map[string]string{"Content-Type": "text/plain", "Server": "Debian"})
	httpResponseMock.EXPECT().Body().Return(strings.NewReader("Hello"))

	httpClientMock.EXPECT().Get("https://example.onion/image.png?id=12&test=2").Return(httpResponseMock, nil)

	subscriberMock.EXPECT().Publish(&event.NewResourceEvent{
		URL:     "https://example.onion/image.png?id=12&test=2",
		Body:    "Hello",
		Headers: map[string]string{"Content-Type": "text/plain", "Server": "Debian"},
	}).Return(nil)

	s := State{httpClient: httpClientMock, allowedContentTypes: []string{"text/plain", "text/css"}}
	if err := s.handleNewURLEvent(subscriberMock, msg); err != nil {
		t.Fail()
	}
}
