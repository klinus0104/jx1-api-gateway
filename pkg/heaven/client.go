package heaven

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const heavenAdd uint32 = 0x2E6D23C1
const heavenSub uint32 = 0x2E6D2398
const heavenXor uint32 = 0x2E6D23CF

type Client struct {
	conn    net.Conn
	table   []uint32
	k1, k2  uint32
	request uint32
}

func LoadTable(path string) ([]uint32, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 || len(b)%4 != 0 {
		return nil, fmt.Errorf("invalid Heaven table")
	}
	t := make([]uint32, len(b)/4)
	for i := range t {
		t[i] = binary.LittleEndian.Uint32(b[i*4:])
	}
	return t, nil
}
func recoverHeavenKey(encoded uint32) uint32 {
	mixed := heavenSub - (encoded ^ heavenXor)
	return (mixed >> 24) | (mixed & 0x00ffff00) | (mixed << 24)
}
func heavenCipher(table []uint32, data []byte, seed uint32) {
	words, rem := len(data)/4, len(data)%4
	state := seed
	for i := 0; i < words; i++ {
		idx := (words - 1 - i + int(state)) % len(table)
		state = table[idx] + heavenAdd
		v := binary.LittleEndian.Uint32(data[i*4:]) ^ state
		binary.LittleEndian.PutUint32(data[i*4:], v)
	}
	if rem > 0 {
		key := state ^ table[rem]
		for i := 0; i < rem; i++ {
			data[words*4+i] ^= byte(key >> uint(8*i))
		}
	}
}
func (h *Client) recvPlain(ctx context.Context) ([]byte, error) {
	if d, ok := ctx.Deadline(); ok {
		_ = h.conn.SetReadDeadline(d)
	}
	var head [2]byte
	if _, err := io.ReadFull(h.conn, head[:]); err != nil {
		return nil, err
	}
	n := int(binary.LittleEndian.Uint16(head[:]))
	if n < 2 {
		return nil, fmt.Errorf("invalid frame length")
	}
	b := make([]byte, n-2)
	_, err := io.ReadFull(h.conn, b)
	return b, err
}
func (h *Client) recv(ctx context.Context) ([]byte, error) {
	b, err := h.recvPlain(ctx)
	if err != nil {
		return nil, err
	}
	heavenCipher(h.table, b, h.k2)
	return b, nil
}
func (h *Client) send(ctx context.Context, body []byte) error {
	if d, ok := ctx.Deadline(); ok {
		_ = h.conn.SetWriteDeadline(d)
	}
	b := append([]byte(nil), body...)
	heavenCipher(h.table, b, h.k1)
	frame := make([]byte, 2+len(b))
	binary.LittleEndian.PutUint16(frame, uint16(len(frame)))
	copy(frame[2:], b)
	_, err := h.conn.Write(frame)
	return err
}
func Dial(ctx context.Context, target, tablePath string) (*Client, error) {
	table, err := LoadTable(tablePath)
	if err != nil {
		return nil, err
	}
	d := net.Dialer{}
	c, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		return nil, err
	}
	h := &Client{conn: c, table: table, request: 1}
	hs, err := h.recvPlain(ctx)
	if err != nil || len(hs) != 42 {
		c.Close()
		return nil, fmt.Errorf("invalid Heaven handshake: %v", err)
	}
	h.k1 = recoverHeavenKey(binary.LittleEndian.Uint32(hs[8:]))
	h.k2 = recoverHeavenKey(binary.LittleEndian.Uint32(hs[17:]))
	return h, nil
}
func (h *Client) Close() { _ = h.conn.Close() }
func (h *Client) Verify(ctx context.Context, server, password, identity string) (uint32, error) {
	body := make([]byte, 0x7e)
	body[1] = 0x24
	binary.LittleEndian.PutUint16(body[2:], 0x7c)
	binary.LittleEndian.PutUint16(body[4:], 1)
	binary.LittleEndian.PutUint16(body[6:], 3)
	binary.LittleEndian.PutUint32(body[8:], h.request)
	h.request++
	copy(body[0xc:], []byte(server))
	copy(body[0x2c:], []byte(password))
	parts := strings.Split(strings.ReplaceAll(identity, "-", ""), "")
	_ = parts
	copy(body[0x70:], []byte{0, 12, 41, 74, 97, 69})
	if err := h.send(ctx, body); err != nil {
		return 0, err
	}
	reply, err := h.recv(ctx)
	if err != nil {
		return 0, err
	}
	if len(reply) != 0x34 || reply[1] != 0x24 {
		return 0, fmt.Errorf("invalid verify response")
	}
	return binary.LittleEndian.Uint32(reply[0x2c:]), nil
}
func (h *Client) gatewayInfo(ctx context.Context, action uint16, account, value string) (uint32, error) {
	body := make([]byte, 0x6a)
	body[1] = 0x26
	binary.LittleEndian.PutUint16(body[2:], 0x68)
	binary.LittleEndian.PutUint16(body[4:], 1)
	binary.LittleEndian.PutUint16(body[6:], action)
	binary.LittleEndian.PutUint32(body[8:], h.request)
	h.request++
	copy(body[0xa:], []byte(account))
	copy(body[0x2a:], []byte(value))
	if err := h.send(ctx, body); err != nil {
		return 0, err
	}
	reply, err := h.recv(ctx)
	if err != nil {
		return 0, err
	}
	if len(reply) < 0x12 {
		return 0, fmt.Errorf("invalid gateway-info response")
	}
	return binary.LittleEndian.Uint32(reply[0x0e:]), nil
}

func (h *Client) LegacyKick(ctx context.Context, account string) error {
	body := make([]byte, 0x6a)
	body[1] = 0x26
	binary.LittleEndian.PutUint16(body[2:], 0x68)
	binary.LittleEndian.PutUint16(body[4:], 7)
	binary.LittleEndian.PutUint32(body[6:], h.request)
	h.request++
	copy(body[0xa:], []byte(account))
	if err := h.send(ctx, body); err != nil {
		return err
	}
	reply, err := h.recv(ctx)
	if err != nil {
		return err
	}
	if len(reply) < 14 || reply[1] != 0x26 {
		return fmt.Errorf("invalid kick response")
	}
	if binary.LittleEndian.Uint32(reply[10:]) != 1 {
		return fmt.Errorf("relay kick rejected")
	}
	return nil
}

func parseTarget(target string) (string, error) {
	if !strings.Contains(target, ":") {
		return target, nil
	}
	_, err := strconv.Atoi(target[strings.LastIndex(target, ":")+1:])
	return target, err
}
