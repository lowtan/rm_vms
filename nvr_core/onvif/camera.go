package onvif

// DiscoveredCamera represents a network video transmitter found on the LAN.
type DiscoveredCamera struct {
	MessageID string
	Address   string // Unique endpoint reference
	Types     string // ONVIF device types
	Scopes    string // Location, hardware, name
	XAddrs    string // Space-separated service URLs used for further ONVIF requests
}