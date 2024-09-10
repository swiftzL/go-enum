package example

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func Test_proxy(t *testing.T) {
	main()
}

func main() {
	addr := flag.String("s", ":1080", "proxy server address")
	flag.Parse()

	if err := proxy(*addr); err != nil {
		panic(err)
	}
}

const netTCP = "tcp"

func proxy(addr string) error {
	lc, err := net.Listen(netTCP, addr)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer lc.Close()

	fmt.Printf("proxy server listening on %s\n", addr)
	for {
		if ac, ea := lc.Accept(); ea == nil {
			go handleProxyFunc(ac)
		}
	}
}

type poolBuf struct {
	buf []byte
	br  *bufio.Reader
}

var bytePool = &sync.Pool{
	New: func() any {
		const bufSize = 64 << 10
		return &poolBuf{
			buf: make([]byte, bufSize),
			br:  bufio.NewReaderSize(nil, bufSize),
		}
	},
}

func handleProxyFunc(c net.Conn) {
	if err := handleProxy(c); err != nil {
		log.Printf("%+v", err)
	}
}

func proxyDial(ctx context.Context, network, addr string) (net.Conn, error) {
	// add the code related to connecting to the secondary proxy server here
	if ctx == nil {
		ctx = context.Background()
	}
	var d net.Dialer
	return d.DialContext(ctx, network, addr)
}

//goland:noinspection GoUnhandledErrorResult
func handleProxy(c net.Conn) error {
	defer c.Close()

	bp := bytePool.Get().(*poolBuf)
	defer bytePool.Put(bp)

	bp.br.Reset(c)
	tp, err := bp.br.Peek(2)
	if err != nil {
		return err
	}

	var rc net.Conn
	switch tp[0] {
	case socksVer5:
		rc, err = dialSocks5(c, bp.br, bp.buf, int(tp[1]))
		if err != nil {
			return err
		}
		defer rc.Close()

		_, err = c.Write(socks5Established)
		if err != nil {
			return err
		}
	case 'C', 'c', 'G', 'g', 'P', 'p', 'O', 'o', 'H', 'h', 'D', 'd', 'T', 't':
		// CONNECT,GET,POST/PUT/PATCH,OPTIONS,HEAD,DELETE,TRACE
		req, err := http.ReadRequest(bp.br)
		if err != nil {
			return err
		}

		req.Method = strings.ToUpper(req.Method)
		if req.Method == http.MethodConnect {
			rc, err = proxyDial(nil, netTCP, req.Host)
			if err != nil {
				return err
			}
			defer rc.Close()

			_, err = c.Write(httpEstablished)
			if err != nil {
				return err
			}
		} else {
			req.Close = true
			resp, err := httpProxyTransport.RoundTrip(req)
			if err != nil {
				return err
			}

			err = resp.Write(c)
			if resp.Body != nil {
				resp.Body.Close()
			}
			return err
		}
	default:
		return errors.New("unsupported protocol version")
	}

	ret := make(chan struct{})
	go func() {
		io.CopyBuffer(c, rc, bp.buf)
		ret <- struct{}{}
	}()
	bp.br.WriteTo(rc)
	<-ret
	return nil
}

var (
	httpProxyTransport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DialContext: proxyDial,
	}
	httpEstablished   = []byte("HTTP/1.1 200 Connection Established\r\n\r\n")
	socks5Established = []byte{socksVer5, 0, 0, 1, 0, 0, 0, 0, 0, 0}
)

const (
	socksVer5  = 5
	aTypIPV4   = 1
	aTypDomain = 3
	aTypIPV6   = 4
)

func dialSocks5(w io.Writer, r io.Reader, buf []byte, n int) (net.Conn, error) {
	_, err := io.ReadFull(r, buf[:n+2])
	if err != nil {
		return nil, err
	}

	_, err = w.Write(socks5Established[:2])
	if err != nil {
		return nil, err
	}

	_, err = io.ReadFull(r, buf[:5])
	if err != nil {
		return nil, err
	}
	if buf[0] != socksVer5 {
		return nil, fmt.Errorf("invalid ver %d", buf[0])
	}
	if buf[1] != 1 { // cmd connect = 1
		return nil, fmt.Errorf("invalid cmd %d", buf[1])
	}

	switch buf[3] {
	case aTypDomain:
		// VER CMD RSV TYPE LEN DST.ADDR DST.PORT
		//  1   1   1   1    1     LEN      2
		n = int(buf[4]) + 7
	case aTypIPV4:
		// VER CMD RSV TYPE DST.ADDR DST.PORT
		//  1   1   1   1   ipv4Len     2
		n = net.IPv4len + 6
	case aTypIPV6:
		// VER CMD RSV TYPE DST.ADDR DST.PORT
		//  1   1   1   1   ipv6Len     2
		n = net.IPv6len + 6
	default:
		return nil, errors.New("socks addr type not supported")
	}

	_, err = io.ReadFull(r, buf[5:n])
	if err != nil {
		return nil, err
	}

	var addr string
	switch buf[3] {
	case aTypDomain:
		n = 5 + int(buf[4])
		addr = string(buf[5:n])
	case aTypIPV4:
		n = 4 + net.IPv4len
		addr = net.IP(buf[4:n]).String()
	case aTypIPV6:
		n = 4 + net.IPv6len
		addr = net.IP(buf[4:n]).String()
	}

	addr = net.JoinHostPort(addr, strconv.FormatUint(uint64(buf[n])<<8|uint64(buf[n+1]), 10))
	return proxyDial(nil, netTCP, addr)
}
