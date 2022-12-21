package gidevice

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

var _ Lockdown = (*lockdown)(nil)

func newLockdown(dev *device) *lockdown {
	return &lockdown{
		umClient: dev.umClient,
		client:   dev.lockdownClient,
		dev:      dev,
	}
}

type lockdown struct {
	umClient  *libimobiledevice.UsbmuxClient
	client    *libimobiledevice.LockdownClient
	sessionID string

	dev        *device
	iOSVersion []int
	pairRecord *PairRecord
}

func (c *lockdown) QueryType() (LockdownType, error) {
	pkt, err := c.client.NewXmlPacket(
		c.client.NewBasicRequest(libimobiledevice.RequestTypeQueryType),
	)
	if err != nil {
		return LockdownType{}, err
	}

	if err = c.client.SendPacket(pkt); err != nil {
		return LockdownType{}, err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = c.client.ReceivePacket(); err != nil {
		return LockdownType{}, err
	}

	var reply libimobiledevice.LockdownTypeResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return LockdownType{}, err
	}

	return LockdownType{Type: reply.Type}, nil
}

func (c *lockdown) GetValue(domain, key string) (v interface{}, err error) {
	pkt, err := c.client.NewXmlPacket(
		c.client.NewGetValueRequest(domain, key),
	)
	if err != nil {
		return nil, err
	}

	if err = c.client.SendPacket(pkt); err != nil {
		return nil, err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = c.client.ReceivePacket(); err != nil {
		return nil, err
	}

	var reply libimobiledevice.LockdownValueResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return nil, err
	}

	v = reply.Value

	return
}

func (c *lockdown) SetValue(domain, key string, value interface{}) (err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = c.client.NewXmlPacket(
		c.client.NewSetValueRequest(domain, key, value),
	); err != nil {
		return err
	}

	if err = c.client.SendPacket(pkt); err != nil {
		return err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = c.client.ReceivePacket(); err != nil {
		return err
	}

	var reply libimobiledevice.LockdownValueResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return err
	}

	if !reply.Value.(bool) {
		return errors.New("lockdown SetValue: Failed")
	}
	return
}

func (c *lockdown) EnterRecovery() (err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = c.client.NewXmlPacket(
		c.client.NewEnterRecoveryRequest(),
	); err != nil {
		return err
	}

	if err = c.client.SendPacket(pkt); err != nil {
		return err
	}

	if _, err = c.client.ReceivePacket(); err != nil {
		return err
	}

	return
}

func (c *lockdown) handshake() (err error) {
	var lockdownType LockdownType
	if lockdownType, err = c.QueryType(); err != nil {
		return err
	}
	if lockdownType.Type != "com.apple.mobile.lockdown" {
		return fmt.Errorf("lockdown handshake 'QueryType': %s", lockdownType.Type)
	}

	// if (device->version < DEVICE_VERSION(7,0,0))
	// for older devices, we need to validate pairing to receive trusted host status

	if c.pairRecord, err = c.dev.ReadPairRecord(); err == nil {
		return nil
	}

	if !strings.Contains(err.Error(), libimobiledevice.ReplyCodeBadDevice.String()) {
		return err
	}

	if c.pairRecord, err = c.Pair(); err != nil {
		return err
	}

	err = c.dev.SavePairRecord(c.pairRecord)

	return
}

func (c *lockdown) Pair() (pairRecord *PairRecord, err error) {
	var buid string
	if buid, err = newUsbmux(c.umClient).ReadBUID(); err != nil {
		return nil, err
	}

	var devPublicKeyPem []byte
	var devWiFiAddr string

	if lockdownValue, err := c.GetValue("", "DevicePublicKey"); err != nil {
		return nil, err
	} else {
		devPublicKeyPem = lockdownValue.([]byte)
	}
	if lockdownValue, err := c.GetValue("", "WiFiAddress"); err != nil {
		return nil, err
	} else {
		devWiFiAddr = lockdownValue.(string)
	}

	if pairRecord, err = generatePairRecord(devPublicKeyPem); err != nil {
		return nil, err
	}

	pairRecord.SystemBUID = buid
	pairRecord.HostID = strings.ToUpper(uuid.NewV4().String())
	hostPrivateKey := pairRecord.HostPrivateKey
	pairRecord.HostPrivateKey = nil
	rootPrivateKey := pairRecord.RootPrivateKey
	pairRecord.RootPrivateKey = nil

	var pkt libimobiledevice.Packet
	if pkt, err = c.client.NewXmlPacket(
		c.client.NewPairRequest(pairRecord),
	); err != nil {
		return nil, err
	}

	if err = c.client.SendPacket(pkt); err != nil {
		return nil, err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = c.client.ReceivePacket(); err != nil {
		return nil, err
	}

	var reply libimobiledevice.LockdownPairResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return nil, err
	}

	pairRecord.EscrowBag = reply.EscrowBag
	pairRecord.WiFiMACAddress = devWiFiAddr
	pairRecord.HostPrivateKey = hostPrivateKey
	pairRecord.RootPrivateKey = rootPrivateKey

	return
}

func (c *lockdown) startSession(pairRecord *PairRecord) (err error) {
	// if we have a running session, stop current one first
	if c.sessionID != "" {
		if err = c.stopSession(); err != nil {
			return err
		}
	}

	var pkt libimobiledevice.Packet
	if pkt, err = c.client.NewXmlPacket(
		c.client.NewStartSessionRequest(pairRecord.SystemBUID, pairRecord.HostID),
	); err != nil {
		return err
	}

	if err = c.client.SendPacket(pkt); err != nil {
		return err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = c.client.ReceivePacket(); err != nil {
		return fmt.Errorf("lockdown start session: %w", err)
	}

	var reply libimobiledevice.LockdownStartSessionResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return err
	}

	if reply.EnableSessionSSL {
		if err = c.client.EnableSSL(c.iOSVersion, pairRecord); err != nil {
			return err
		}
	}

	c.sessionID = reply.SessionID
	return
}

func (c *lockdown) stopSession() (err error) {
	if c.sessionID == "" {
		return nil
	}

	var pkt libimobiledevice.Packet
	if pkt, err = c.client.NewXmlPacket(
		c.client.NewStopSessionRequest(c.sessionID),
	); err != nil {
		return err
	}

	if err = c.client.SendPacket(pkt); err != nil {
		return err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = c.client.ReceivePacket(); err != nil {
		return fmt.Errorf("lockdown stop session: %w", err)
	}

	var reply libimobiledevice.LockdownBasicResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return err
	}

	c.sessionID = ""
	return
}

func (c *lockdown) startService(service string, escrowBag []byte) (dynamicPort int, enableSSL bool, err error) {
	req := c.client.NewStartServiceRequest(service)
	if escrowBag != nil {
		req.EscrowBag = escrowBag
	}

	var pkt libimobiledevice.Packet
	if pkt, err = c.client.NewXmlPacket(req); err != nil {
		return 0, false, err
	}

	if err = c.client.SendPacket(pkt); err != nil {
		return 0, false, err
	}

	respPkt, err := c.client.ReceivePacket()
	if err != nil {
		return 0, false, err
	}

	var reply libimobiledevice.LockdownStartServiceResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return 0, false, err
	}

	if reply.Error != "" {
		return 0, false, fmt.Errorf("lockdown start service: %s", reply.Error)
	}

	dynamicPort = reply.Port
	enableSSL = reply.EnableServiceSSL

	return
}

func (c *lockdown) ImageMounterService() (imageMounter ImageMounter, err error) {
	var innerConn InnerConn
	if innerConn, err = c._startService(libimobiledevice.ImageMounterServiceName, nil); err != nil {
		return nil, err
	}
	imageMounterClient := libimobiledevice.NewImageMounterClient(innerConn)
	imageMounter = newImageMounter(imageMounterClient)
	return
}

func (c *lockdown) ScreenshotService() (screenshot Screenshot, err error) {
	var innerConn InnerConn
	if innerConn, err = c._startService(libimobiledevice.ScreenshotServiceName, nil); err != nil {
		return nil, err
	}
	screenshotClient := libimobiledevice.NewScreenshotClient(innerConn)
	screenshot = newScreenshot(screenshotClient)
	return
}

func (c *lockdown) SimulateLocationService() (simulateLocation SimulateLocation, err error) {
	var innerConn InnerConn
	if innerConn, err = c._startService(libimobiledevice.SimulateLocationServiceName, nil); err != nil {
		return nil, err
	}
	simulateLocationClient := libimobiledevice.NewSimulateLocationClient(innerConn)
	simulateLocation = newSimulateLocation(simulateLocationClient)
	return
}

func (c *lockdown) InstallationProxyService() (installationProxy InstallationProxy, err error) {
	var innerConn InnerConn
	if innerConn, err = c._startService(libimobiledevice.InstallationProxyServiceName, nil); err != nil {
		return nil, err
	}
	installationProxyClient := libimobiledevice.NewInstallationProxyClient(innerConn)
	installationProxy = newInstallationProxy(installationProxyClient)
	return
}

func (c *lockdown) InstrumentsService() (instruments Instruments, err error) {
	service := libimobiledevice.InstrumentsServiceName
	if DeviceVersion(c.iOSVersion...) >= DeviceVersion(14, 0, 0) {
		service = libimobiledevice.InstrumentsSecureProxyServiceName
	}

	var innerConn InnerConn
	if innerConn, err = c._startService(service, nil); err != nil {
		return nil, err
	}
	instrumentsClient := libimobiledevice.NewInstrumentsClient(innerConn)
	instruments = newInstruments(instrumentsClient)

	if service == libimobiledevice.InstrumentsServiceName {
		_ = innerConn.DismissSSL()
	}

	if err = instruments.notifyOfPublishedCapabilities(); err != nil {
		return nil, err
	}

	return
}

func (c *lockdown) TestmanagerdService() (testmanagerd Testmanagerd, err error) {
	service := libimobiledevice.TestmanagerdServiceName
	if DeviceVersion(c.iOSVersion...) >= DeviceVersion(14, 0, 0) {
		service = libimobiledevice.TestmanagerdSecureServiceName
	}

	var innerConn InnerConn
	if innerConn, err = c._startService(service, nil); err != nil {
		return nil, err
	}
	testmanagerdClient := libimobiledevice.NewTestmanagerdClient(innerConn)
	testmanagerd = newTestmanagerd(testmanagerdClient, c.iOSVersion)

	if service == libimobiledevice.TestmanagerdServiceName {
		_ = innerConn.DismissSSL()
	}

	if err = testmanagerd.notifyOfPublishedCapabilities(); err != nil {
		return nil, err
	}

	return
}

func (c *lockdown) AfcService() (afc Afc, err error) {
	var innerConn InnerConn
	if innerConn, err = c._startService(libimobiledevice.AfcServiceName, nil); err != nil {
		return nil, err
	}
	afcClient := libimobiledevice.NewAfcClient(innerConn)
	afc = newAfc(afcClient)
	return
}

func (c *lockdown) HouseArrestService() (houseArrest HouseArrest, err error) {
	var innerConn InnerConn
	if innerConn, err = c._startService(libimobiledevice.HouseArrestServiceName, nil); err != nil {
		return nil, err
	}
	houseArrestClient := libimobiledevice.NewHouseArrestClient(innerConn)
	houseArrest = newHouseArrest(houseArrestClient)
	return
}

func (c *lockdown) SyslogRelayService() (syslogRelay SyslogRelay, err error) {
	var innerConn InnerConn
	if innerConn, err = c._startService(libimobiledevice.SyslogRelayServiceName, nil); err != nil {
		return nil, err
	}
	syslogRelayClient := libimobiledevice.NewSyslogRelayClient(innerConn)
	syslogRelay = newSyslogRelay(syslogRelayClient)
	return
}

func (c *lockdown) PcapdService(targetPID int, targetProcName string) (pcapd Pcapd, err error) {
	var innerConn InnerConn
	if innerConn, err = c._startService(libimobiledevice.PcapdServiceName, nil); err != nil {
		return nil, err
	}
	pcapdClient := libimobiledevice.NewPcapdClient(innerConn, targetPID, targetProcName)
	return newPcapdClient(pcapdClient), nil
}

func (c *lockdown) DiagnosticsRelayService() (diagnostics DiagnosticsRelay, err error) {
	var innerConn InnerConn
	if innerConn, err = c._startService(libimobiledevice.DiagnosticsRelayServiceName, nil); err != nil {
		return nil, err
	}
	diagnosticsRelayClient := libimobiledevice.NewDiagnosticsRelayClient(innerConn)
	diagnostics = newDiagnosticsRelay(diagnosticsRelayClient)

	return
}

func (c *lockdown) SpringBoardService() (springboard SpringBoard, err error) {
	var innerConn InnerConn
	if innerConn, err = c._startService(libimobiledevice.SpringBoardServiceName, nil); err != nil {
		return nil, err
	}
	springBoardServiceClient := libimobiledevice.NewSpringBoardClient(innerConn)
	springboard = newSpringBoard(springBoardServiceClient)
	return
}

func (c *lockdown) CrashReportMoverService() (crashReportMover CrashReportMover, err error) {
	var innerConn InnerConn
	if innerConn, err = c._startService(libimobiledevice.CrashReportMoverServiceName, nil); err != nil {
		return nil, err
	}

	mover := newCrashReportMover(libimobiledevice.NewCrashReportMoverClient(innerConn))
	if err = mover.readPing(); err != nil {
		return nil, err
	}

	if innerConn, err = c._startService(libimobiledevice.CrashReportCopyMobileServiceName, nil); err != nil {
		return nil, err
	}
	mover.afc = newAfc(libimobiledevice.NewAfcClient(innerConn))

	crashReportMover = mover
	return
}

func (c *lockdown) _startService(serviceName string, escrowBag []byte) (innerConn InnerConn, err error) {
	if err = c.handshake(); err != nil {
		return nil, err
	}

	if err = c.startSession(c.pairRecord); err != nil {
		return nil, err
	}

	dynamicPort, enableSSL, err := c.startService(serviceName, escrowBag)
	if err != nil {
		return nil, err
	}

	if err = c.stopSession(); err != nil {
		return nil, err
	}

	if innerConn, err = c.dev.NewConnect(dynamicPort, 0); err != nil {
		return nil, err
	}
	// clean deadline
	innerConn.Timeout(0)

	if enableSSL {
		if err = innerConn.Handshake(c.iOSVersion, c.pairRecord); err != nil {
			return nil, err
		}
	}
	return
}

func (c *lockdown) _getProductVersion() (version []int, err error) {
	if c.iOSVersion != nil {
		return c.iOSVersion, nil
	}

	var devProductVersion []string
	if lockdownValue, err := c.GetValue("", "ProductVersion"); err != nil {
		return nil, err
	} else {
		devProductVersion = strings.Split(lockdownValue.(string), ".")
	}

	version = make([]int, len(devProductVersion))
	for i, v := range devProductVersion {
		version[i], _ = strconv.Atoi(v)
	}

	// if len(version) == 2 {
	// 	version = append(version, 0)
	// }
	c.iOSVersion = version

	return
}

func generatePairRecord(devPublicKeyPem []byte) (pairRecord *PairRecord, err error) {
	block, _ := pem.Decode(devPublicKeyPem)
	var deviceKey *rsa.PublicKey
	if deviceKey, err = x509.ParsePKCS1PublicKey(block.Bytes); err != nil {
		return nil, err
	}

	var rootKey, hostKey *rsa.PrivateKey
	if rootKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return nil, err
	}
	if hostKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return nil, err
	}
	serialNumber := big.NewInt(0)
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * (24 * 365) * 10)

	rootTemplate := x509.Certificate{
		IsCA:                  true,
		SerialNumber:          serialNumber,
		Version:               2,
		SignatureAlgorithm:    x509.SHA1WithRSA,
		PublicKeyAlgorithm:    x509.RSA,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
	}

	var caCert, cert []byte
	if caCert, err = x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, rootKey.Public(), rootKey); err != nil {
		return nil, err
	}

	hostTemplate := x509.Certificate{
		SerialNumber:          serialNumber,
		Version:               2,
		SignatureAlgorithm:    x509.SHA1WithRSA,
		PublicKeyAlgorithm:    x509.RSA,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	if cert, err = x509.CreateCertificate(rand.Reader, &hostTemplate, &rootTemplate, hostKey.Public(), rootKey); err != nil {
		return nil, err
	}

	h := sha1.New()
	if _, err = h.Write(rootKey.N.Bytes()); err != nil {
		return nil, err
	}
	subjectKeyId := h.Sum(nil)

	deviceTemplate := x509.Certificate{
		SerialNumber:          serialNumber,
		Version:               2,
		SignatureAlgorithm:    x509.SHA1WithRSA,
		PublicKeyAlgorithm:    x509.RSA,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		SubjectKeyId:          subjectKeyId,
	}

	var deviceCert []byte
	if deviceCert, err = x509.CreateCertificate(rand.Reader, &deviceTemplate, &rootTemplate, deviceKey, rootKey); err != nil {
		return nil, err
	}

	var deviceCertPEM []byte
	if deviceCertPEM, err = encodePemCertificate(deviceCert); err != nil {
		return nil, err
	}

	var caPEM, caPrivatePEM []byte
	if caPEM, caPrivatePEM, err = encodePairPemFormat(caCert, rootKey); err != nil {
		return nil, err
	}

	var certPEM, certPrivatePEM []byte
	if certPEM, certPrivatePEM, err = encodePairPemFormat(cert, hostKey); err != nil {
		return nil, err
	}

	pairRecord = new(PairRecord)

	pairRecord.DeviceCertificate = deviceCertPEM
	pairRecord.HostCertificate = certPEM
	pairRecord.HostPrivateKey = certPrivatePEM
	pairRecord.RootCertificate = caPEM
	pairRecord.RootPrivateKey = caPrivatePEM

	return
}

func encodePairPemFormat(cert []byte, key *rsa.PrivateKey) ([]byte, []byte, error) {
	p, err := encodePemCertificate(cert)
	if err != nil {
		return nil, nil, err
	}

	buf := new(bytes.Buffer)
	if err := pem.Encode(buf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return nil, nil, err
	}

	privy := buf.Bytes()

	return p, privy, nil
}

func encodePemCertificate(cert []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := pem.Encode(buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
