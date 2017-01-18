package main

import (
	. "gopkg.in/check.v1"

	jpegp "image/jpeg"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ahmetalpbalkan/go-httpbin"
	"github.com/barnacs/compy/proxy"
	tc "github.com/barnacs/compy/transcoder"
	"github.com/chai2010/webp"
)

func Test(t *testing.T) {
	TestingT(t)
}

type CompyTest struct {
	client *http.Client
	server *httptest.Server
	proxy  *proxy.Proxy
}

var _ = Suite(&CompyTest{})

func (s *CompyTest) SetUpSuite(c *C) {
	s.server = httptest.NewServer(httpbin.GetMux())

	s.proxy = proxy.New()
	s.proxy.AddTranscoder("image/jpeg", tc.NewJpeg(50))
	s.proxy.AddTranscoder("text/html", &tc.Gzip{&tc.Identity{}, *gzip, true})
	go func() {
		err := s.proxy.Start(*host)
		if err != nil {
			c.Fatal(err)
		}
	}()

	proxyUrl := &url.URL{Scheme: "http", Host: "localhost" + *host}
	s.client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
}

func (s *CompyTest) TearDownSuite(c *C) {
	s.server.Close()

	// TODO: Go 1.8 will provide http.Server.Shutdown for proxy.Proxy
}

func (s *CompyTest) TestHttpBin(c *C) {
	resp, err := s.client.Get(s.server.URL + "/status/200")
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 200)
}

func (s *CompyTest) TestJpeg(c *C) {
	req, err := http.NewRequest("GET", s.server.URL+"/image/jpeg", nil)
	c.Assert(err, IsNil)

	resp, err := s.client.Do(req)
	c.Assert(err, IsNil)
	defer resp.Body.Close()
	c.Assert(resp.StatusCode, Equals, 200)
	c.Assert(resp.Header.Get("Content-Type"), Equals, "image/jpeg")

	_, err = jpegp.Decode(resp.Body)
	c.Assert(err, IsNil)
}

func (s *CompyTest) TestWebP(c *C) {
	req, err := http.NewRequest("GET", s.server.URL+"/image/jpeg", nil)
	c.Assert(err, IsNil)
	req.Header.Add("Accept", "image/webp,image/jpeg")

	resp, err := s.client.Do(req)
	c.Assert(err, IsNil)
	defer resp.Body.Close()
	c.Assert(resp.StatusCode, Equals, 200)
	c.Assert(resp.Header.Get("Content-Type"), Equals, "image/webp")

	_, err = webp.Decode(resp.Body)
	c.Assert(err, IsNil)
}
