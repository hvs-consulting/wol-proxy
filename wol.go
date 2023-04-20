package main

// Parts taken from: https://github.com/sabhiram/go-wol.

////////////////////////////////////////////////////////////////////////////////

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"regexp"
)

////////////////////////////////////////////////////////////////////////////////

var (
	delims = ":-"
	reMAC  = regexp.MustCompile(`^([0-9a-fA-F]{2}[` + delims + `]){5}([0-9a-fA-F]{2})$`)
)

////////////////////////////////////////////////////////////////////////////////

// MACAddress represents a 6 byte network mac address.
type MACAddress [6]byte

// MagicPacket is constituted of 6 bytes of 0xFF followed by 16-groups of the
// destination MAC address.
type MagicPacket struct {
	header  [6]byte
	payload [16]MACAddress
}

// New returns a magic packet based on a mac address string.
func NewMagicPacket(mac string) (*MagicPacket, error) {
	var packet MagicPacket
	var macAddr MACAddress

	hwAddr, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}

	// We only support 6 byte MAC addresses since it is much harder to use the
	// binary.Write(...) interface when the size of the MagicPacket is dynamic.
	if !reMAC.MatchString(mac) {
		return nil, fmt.Errorf("%s is not a IEEE 802 MAC-48 address", mac)
	}

	// Copy bytes from the returned HardwareAddr -> a fixed size MACAddress.
	for idx := range macAddr {
		macAddr[idx] = hwAddr[idx]
	}

	// Setup the header which is 6 repetitions of 0xFF.
	for idx := range packet.header {
		packet.header[idx] = 0xFF
	}

	// Setup the payload which is 16 repetitions of the MAC addr.
	for idx := range packet.payload {
		packet.payload[idx] = macAddr
	}

	return &packet, nil
}

// Marshal serializes the magic packet structure into a 102 byte slice.
func (mp *MagicPacket) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, mp); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Sends a magic packet to specified MAC address using the specified broadcast address using a UDP connection
func SendMagicPacket(macAddr, broadcastAddress string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", broadcastAddress)
	if err != nil {
		return err
	}

	mp, err := NewMagicPacket(macAddr)
	if err != nil {
		return err
	}

	// convert to bytes
	bs, err := mp.Marshal()
	if err != nil {
		return err
	}

	// creating UDP socket
	// this is simpler than sending an EthernetFrame directly
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Printf("Sending magic packet to: %s\n", macAddr)
	log.Printf("Broadcasting via: %s\n", broadcastAddress)
	n, err := conn.Write(bs)
	if err == nil && n != 102 {
		err = fmt.Errorf("sent %d bytes (expected 102 bytes)", n)
	}
	if err != nil {
		return err
	}

	log.Println("Magic packet sent")
	return nil
}
