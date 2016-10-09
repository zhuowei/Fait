package main

import (
	"net"
	"fmt"
	"encoding/binary"
	"encoding/hex"
	"bytes"
	goplist "github.com/DHowett/go-plist"
)

// a very basic Usbmux client. Not as good as https://github.com/Mitchell-Riley/mux; just need it for basic tcp.
// referenced http://squirrel-lang.org/forums/default.aspx?g=posts&m=6816 for protocol documentation, but doesn't use its code.

type UsbmuxClient struct {
	conn net.Conn
	tag uint32
}

func (client *UsbmuxClient) Write(version, packettype uint32, payload map[string] interface{}) (error) {
	tag := client.tag
	client.tag += 1
	dat, err := goplist.Marshal(payload, goplist.XMLFormat)
	if err != nil {
		return err
	}
	length := 4 + 4 + 4 + 4 + len(dat)

	w := new(bytes.Buffer)
	
	binary.Write(w, binary.LittleEndian, uint32(length))
	binary.Write(w, binary.LittleEndian, version)
	binary.Write(w, binary.LittleEndian, packettype)
	binary.Write(w, binary.LittleEndian, tag)
	w.Write(dat)
	fmt.Println(hex.Dump(w.Bytes()))
	_, err = w.WriteTo(client.conn)
	return err
}

func (client *UsbmuxClient) Read() (uint32, uint32, uint32, map[string] interface{}, error) {
	var size, version, packettype, tag uint32
	err := binary.Read(client.conn, binary.LittleEndian, &size);
	if err != nil {
		return 0, 0, 0, nil, err
	}
	b := make([]byte, size - 4)
	_, err = client.conn.Read(b)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	r := bytes.NewReader(b)
	fmt.Println(hex.Dump(b))
	binary.Read(r, binary.LittleEndian, &version)
	binary.Read(r, binary.LittleEndian, &packettype)
	binary.Read(r, binary.LittleEndian, &tag)
	var plist map[string]interface{}
	_, err = goplist.Unmarshal(b[12:], &plist)
	
	return version, packettype, tag, plist, err
}

func (client *UsbmuxClient) Transact(version, packettype uint32, payload map[string]interface{}) (uint32, uint32, uint32, map[string]interface{}, error) {
	err := client.Write(version, packettype, payload)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	return client.Read()
}

func (client *UsbmuxClient) Connect(deviceid int, portnumber int) (map[string]interface{}, error) {
	d := map[string] interface{} {
		"BundleID": "net.zhuoweizhang.fait",
		"ClientVersionString": "0.1",
		"ProgName": "fait",
		"MessageType": "Connect",
		"DeviceID": deviceid,
		"PortNumber": ((portnumber & 0xff) << 8) | ((portnumber >> 8) & 0xff),
	}
	_, _, _, ret, err := client.Transact(1, 8, d)
	return ret, err
}

func (client *UsbmuxClient) ListDevices() (map[string]interface{}, error) {
	d := map[string] interface{} {
		"BundleID": "net.zhuoweizhang.fait",
		"ClientVersionString": "0.1",
		"ProgName": "fait",
		"MessageType": "ListDevices",
	}
	_, _, _, ret, err := client.Transact(1, 8, d)
	return ret, err
}

func (client *UsbmuxClient) Conn() (net.Conn) {
	return client.conn
}

func NewUsbmuxClientTCP() (*UsbmuxClient, error) {
	conn, err := net.Dial("tcp", "127.0.0.1:27015")
	if err != nil {
		return nil, err
	}
	client := new(UsbmuxClient)
	client.conn = conn
	return client, nil
}

func UsbmuxConnect(deviceid int, portnumber int) (net.Conn, error) {
	client, err := NewUsbmuxClientTCP()
	if err != nil {
		// todo close client
		return nil, err
	}
	d, err := client.ListDevices()
	if err != nil {
		return nil, err
	}
	fmt.Println(d)
	d, err = client.Connect(deviceid, portnumber)
	if err != nil {
		return nil, err
	}
	fmt.Println(d)
	return client.Conn(), nil
}