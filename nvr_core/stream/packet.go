package stream

func payloadForWebSocket(packet StreamPacket) []byte {

	// Allocate a new slice: 1 byte (Header) + Length of Payload
	msg := make([]byte, 1+len(packet.Payload))

	// Set the first byte to the MediaType (0 = Video, 1 = Audio)
	msg[0] = packet.MediaType

	// Copy the raw media data right after the header
	copy(msg[1:], packet.Payload)

	return msg

}