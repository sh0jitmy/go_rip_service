package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"time"
)

// RIPv2ヘッダー
const (
	RIPCommandRequest  = 1
	RIPCommandResponse = 2
)

type RIPEntry struct {
	AddressFamily uint16
	RouteTag      uint16
	IP            [4]byte
	SubnetMask    [4]byte
	NextHop       [4]byte
	Metric        uint32
}

type RIPPacket struct {
	Command  uint8
	Version  uint8
	Zero     uint16
	Entries  []RIPEntry
}

// RIPv2の送信
func sendRIPUpdate(conn *net.UDPConn,dstaddr *net.UDPAddr) {
	/*
	conn, err := net.Dial("udp", "224.0.0.9:520")
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()
	*/
	packet := createRIPUpdatePacket()
	log.Printf("update packet: %v\n", packet)

	// バッファにパケットを書き込む
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, packet.Command)
	binary.Write(&buf, binary.BigEndian, packet.Version)
	binary.Write(&buf, binary.BigEndian, packet.Zero)
	for _, entry := range packet.Entries {
		binary.Write(&buf, binary.BigEndian, entry)
	}

	// パケット送信
	//if _, err := conn.Write(buf.Bytes()); err != nil {
	if _, err := conn.WriteToUDP(buf.Bytes(),dstaddr); err != nil {
		log.Printf("Failed to send RIP update: %v", err)
	}
}

// RIPv2の受信
func startRIPService(ifname, addr,port,dstaddrport string) {
	var ifi *net.Interface = nil	
	var conn *net.UDPConn = nil	
	multienable := false

	dstaddr := net.ParseIP(addr)
	if (dstaddr.IsMulticast()) {
		multienable = true
	}
	address, err := net.ResolveUDPAddr("udp", addr+":"+port)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}
	if ifname != "" {
		ifi , _ = net.InterfaceByName(ifname)
	}
	if (multienable) {
		conn, err = net.ListenMulticastUDP("udp", ifi,address)
		if err != nil {
			log.Fatalf("Failed to start listener: %v", err)
		}
	} else {
		conn, err = net.ListenUDP("udp", address)
		if err != nil {
			log.Fatalf("Failed to start listener: %v", err)
		}
	}
	defer conn.Close()
	if dstaddrport == "" {
		dstaddrport = addr+":"+port
	}
	
	go startRIPBroadcaster(conn,dstaddrport)

	for {
		buf := make([]byte, 2048)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("Error reading RIP packet: %v", err)
			continue
		}

		log.Printf("Received %d bytes from %s", n, addr)
		processRIPPacket(buf[:n])
	}
}

// startRIPBroadcaster 定期的にRIP Updateを送信する
func startRIPBroadcaster(conn *net.UDPConn,dstaddrport string) {
	dstaddress, err := net.ResolveUDPAddr("udp", dstaddrport)
	if err != nil {
		log.Fatalf("Failed to start Broadcaster: %v", err)
	}
	/*	
	conn, err := net.DialUDP("udp", address)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	*/
	ticker := time.NewTicker(30 * time.Second) // 30秒ごとに送信
	defer ticker.Stop()
	log.Println("RIP Broadcaster started")
	for {
		select {
		case <-ticker.C:
			log.Println("Sending RIP Update")
			sendRIPUpdate(conn,dstaddress)
		}
	}
}


func processRIPPacket(data []byte) {
	var packet RIPPacket
	buf := bytes.NewReader(data)

	// パケットヘッダーの読み込み
	if err := binary.Read(buf, binary.BigEndian, &packet.Command); err != nil {
		log.Printf("Failed to read RIP Command: %v", err)
		return
	}
	if err := binary.Read(buf, binary.BigEndian, &packet.Version); err != nil {
		log.Printf("Failed to read RIP Version: %v", err)
		return
	}
	if err := binary.Read(buf, binary.BigEndian, &packet.Zero); err != nil {
		log.Printf("Failed to read RIP Zero: %v", err)
		return
	}

	// RIPv2のみ対応
	if packet.Version != 2 {
		log.Printf("Unsupported RIP version: %d", packet.Version)
		return
	}

	// エントリを読み込む
	for buf.Len() > 0 {
		var entry RIPEntry
		if err := binary.Read(buf, binary.BigEndian, &entry); err != nil {
			log.Printf("Failed to read RIP Entry: %v", err)
			break
		}

		// 経路情報を更新
		destination := net.IP(entry.IP[:]).String()
		gateway := net.IP(entry.NextHop[:]).String()
		subnetMask := net.IP(entry.SubnetMask[:]).String()

		route := Route{
			Destination: destination + "/" + subnetMask,
			Gateway:     gateway,
			Interface:   "unknown", // ここで受信インターフェイスを指定可能
			ExpiresAt:   time.Now().Add(180 * time.Second), // 有効期限180秒
		}
		addRIPRoute(route)
		log.Printf("Added/Updated RIP route: %+v", route)
	}
}

func createRIPUpdatePacket() RIPPacket {
	packet := RIPPacket{
		Command: RIPCommandResponse,
		Version: 2,
		Zero:    0,
	}

	routes := getRoutes()
	for _, route := range routes {
		entry := RIPEntry{
			AddressFamily: 2, // IPv4
			RouteTag:      0,
		}

		// IP, サブネットマスク, ゲートウェイをバイナリ化
		ip, ipNet, err := net.ParseCIDR(route.Destination)
		if err != nil {
			log.Printf("Invalid destination: %s", route.Destination)
			continue
		}
		copy(entry.IP[:], ip.To4())
		copy(entry.SubnetMask[:], net.IP(ipNet.Mask).To4())
		copy(entry.NextHop[:], net.ParseIP(route.Gateway).To4())

		// メトリックを設定 (固定値)
		entry.Metric = 1

		packet.Entries = append(packet.Entries, entry)
	}

	return packet
}
