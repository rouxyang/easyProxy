package server

import (
	"github.com/cnlh/easyProxy/bridge"
	"github.com/cnlh/easyProxy/utils"
	"net"
	"strings"
)

type UdpModeServer struct {
	server
	listener *net.UDPConn
	udpMap   map[string]*utils.Conn
}

func NewUdpModeServer(bridge *bridge.Bridge, task *utils.Tunnel) *UdpModeServer {
	s := new(UdpModeServer)
	s.bridge = bridge
	s.udpMap = make(map[string]*utils.Conn)
	s.task = task
	s.config = utils.DeepCopyConfig(task.Config)
	return s
}

//开始
func (s *UdpModeServer) Start() error {
	var err error
	s.listener, err = net.ListenUDP("udp", &net.UDPAddr{net.ParseIP("0.0.0.0"), s.task.TcpPort, ""})
	if err != nil {
		return err
	}
	buf := utils.BufPoolUdp.Get().([]byte)
	for {
		n, addr, err := s.listener.ReadFromUDP(buf)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			}
			continue
		}
		if !s.ResetConfig() {
			continue
		}
		go s.process(addr, buf[:n])
	}
	return nil
}

func (s *UdpModeServer) process(addr *net.UDPAddr, data []byte) {
	link := utils.NewLink(s.task.Client.GetId(), utils.CONN_UDP, s.task.Target, s.config.CompressEncode, s.config.CompressDecode, s.config.Crypt, nil, s.task.Flow, s.listener, s.task.Client.Rate, addr)

	if tunnel, err := s.bridge.SendLinkInfo(s.task.Client.Id, link); err != nil {
		return
	} else {
		s.task.Flow.Add(len(data), 0)
		tunnel.SendMsg(data, link)
	}
}

func (s *UdpModeServer) Close() error {
	return s.listener.Close()
}
