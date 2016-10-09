package main
import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"bytes"
	"io/ioutil"
	goplist "github.com/DHowett/go-plist"
	"crypto/tls"
)
type LockdownClient struct {
	conn net.Conn
	originalConn net.Conn
	pairingRecord LockdownPairRecord
}

type LockdownPairRecord struct {
	DeviceCertificate []byte
	HostPrivateKey []byte
	HostCertificate []byte
	RootPrivateKey []byte
	RootCertificate []byte
	SystemBUID string
	HostID string
}

func (record *LockdownPairRecord) toMapForLockdownClient() (map[string] interface{}) {
	return map[string] interface{} {
		"DeviceCertificate": record.DeviceCertificate,
		"HostCertificate": record.HostCertificate,
		"RootCertificate": record.RootCertificate,
		"HostID": record.HostID,
	}
}

func (client *LockdownClient) Write(plist interface{}) (error) {
	dat, error := goplist.Marshal(plist, goplist.XMLFormat)
	if error != nil {
		return error
	}
	error = binary.Write(client.conn, binary.BigEndian, uint32(len(dat)))
	if error != nil {
		return error
	}
	_, error = client.conn.Write(dat)
	if error != nil {
		return error
	}
	return nil
}

func (client *LockdownClient) Read() (interface{}, error) {
	var size uint32;
	error := binary.Read(client.conn, binary.BigEndian, &size);
	if error != nil {
		return nil, error
	}
	b := make([]byte, size)
	_, error = client.conn.Read(b)
	if error != nil {
		return nil, error
	}
	print(hex.Dump(b))
	var plist interface{} = nil;
	_, error = goplist.Unmarshal(b, &plist)
	return plist, error
}

func (client *LockdownClient) Transact(request interface{}) (interface{}, error) {
	error := client.Write(request)
	if error != nil {
		return nil, error
	}
	ret, error := client.Read()
	return ret, error

}

func (client *LockdownClient) GetValue(domain string, key string) (interface{}, error) {
	d := map[string] interface{} {
		"Request": "GetValue",
		"Label": "iTunesHelper",
		"Key": key,
	}
	return client.Transact(d)
}

func (client *LockdownClient) QueryType() (interface{}, error) {
	d := map[string] interface{} {
		"Request": "QueryType",
		"Label": "iTunesHelper",
	}
	return client.Transact(d)
}

func (client *LockdownClient) ValidatePair(pairingRecord LockdownPairRecord) (interface{}, error) {
	d := map[string] interface{} {
		"Request": "ValidatePair",
		"Label": "iTunesHelper",
		"PairRecord": pairingRecord.toMapForLockdownClient(),
	}
	out, error := client.Transact(d)
	// todo check result
	client.pairingRecord = pairingRecord
	return out, error
}

func (client *LockdownClient) StartSession(hostID string, systemBUID string) (interface{}, error) {
	d := map[string] interface{} {
		"Request": "StartSession",
		"Label": "iTunesHelper",
		"HostID": hostID,
		"SystemBUID": systemBUID,
	}
	out, error := client.Transact(d)
	// todo check result
	client.enableTLS()
	return out, error
}

func (client *LockdownClient) StartService(serviceName string) (interface{}, error) {
	d := map[string] interface{} {
		"Request": "StartService",
		"Label": "iTunesHelper",
		"Service": serviceName,
	}
	return client.Transact(d)
}

func (client *LockdownClient) enableTLS() (error) {
	var config tls.Config
	hostcert, err := tls.X509KeyPair(client.pairingRecord.HostCertificate, client.pairingRecord.HostPrivateKey)
	if err != nil {
		return err
	}
	config.Certificates = []tls.Certificate{hostcert}
	config.InsecureSkipVerify = true; // fixme: yolo
	conn := tls.Client(client.originalConn, &config)
	client.conn = conn
	return nil
}

func NewLockdownClient(conn net.Conn) (*LockdownClient) {
	client := new(LockdownClient)
	client.conn = conn
	client.originalConn = conn
	return client
}

func ReadLockdownPairRecord(plistdata []byte) (LockdownPairRecord, error) {
	var record LockdownPairRecord
	decoder := goplist.NewDecoder(bytes.NewReader(plistdata))
	error := decoder.Decode(&record)
	return record, error
}

func main() {
	pairrecordbytes, error := ioutil.ReadFile("pairrecord.plist")
	if error != nil {
		panic(error)
	}
	pairrecord, error := ReadLockdownPairRecord(pairrecordbytes)
	if error != nil {
		panic(error)
	}
	conn, error := net.Dial("tcp", "localhost:62078")
	if error != nil {
		panic(error)
	}
	client := NewLockdownClient(conn)
	d, error := client.ValidatePair(pairrecord)
	if error != nil {
		panic(error)
	}
	fmt.Println(d)
	d, error = client.StartSession(pairrecord.HostID, pairrecord.SystemBUID)
	if error != nil {
		panic(error)
	}
	d, error = client.StartService("com.apple.ait.aitd")
	if error != nil {
		panic(error)
	}
	fmt.Println(d)
}