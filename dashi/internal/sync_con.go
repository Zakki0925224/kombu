package internal

import "fmt"

const MAGIC_BYTE_0 byte = 0xa8
const MAGIC_BYTE_1 byte = 0x9c
const MAGIC_BYTE_2 byte = 0x54

var requestMap = map[byte]string{
	0x00: "close_con",
	0x01: "get_cid",
	0x02: "get_init_opt",
	0x03: "send_mount_list",
	0x04: "unmount",
	0x05: "ping",
}

func GetRequestFromBytes(bytes []byte) (string, error) {
	if len(bytes) != 4 {
		return "", fmt.Errorf("bytes length too short")
	}

	if bytes[0] != MAGIC_BYTE_0 || bytes[1] != MAGIC_BYTE_1 || bytes[2] != MAGIC_BYTE_2 {
		return "", fmt.Errorf("invalid magic bytes")
	}

	req, ok := requestMap[bytes[3]]
	if !ok {
		return "", fmt.Errorf("invalid request type")
	}

	return req, nil
}

func RequestToBytes(req string) ([]byte, error) {
	for k, v := range requestMap {
		if v == req {
			return []byte{MAGIC_BYTE_0, MAGIC_BYTE_1, MAGIC_BYTE_2, k}, nil
		}
	}

	return nil, fmt.Errorf("invalid request type")
}
