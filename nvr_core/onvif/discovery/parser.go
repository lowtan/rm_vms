package discovery

import (
	"encoding/xml"
	"strings"
	"nvr_core/onvif"
)

// Envelope represents the expected WS-Discovery ProbeMatch response.
type Envelope struct {
	Header struct {
		MessageID string `xml:"MessageID"`
	} `xml:"Header"`
	Body struct {
		ProbeMatches struct {
			ProbeMatch []struct {
				EndpointReference struct {
					Address string `xml:"Address"`
				} `xml:"EndpointReference"`
				Types  string `xml:"Types"`
				Scopes string `xml:"Scopes"`
				XAddrs string `xml:"XAddrs"`
			} `xml:"ProbeMatch"`
		} `xml:"ProbeMatches"`
	} `xml:"Body"`
}

// ParseProbeMatch extracts camera data from the SOAP WS-Discovery response.
func ParseProbeMatch(payload []byte) ([]onvif.DiscoveredCamera, error) {
	var env Envelope
	if err := xml.Unmarshal(payload, &env); err != nil {
		return nil, err
	}

	var cameras []onvif.DiscoveredCamera
	for _, match := range env.Body.ProbeMatches.ProbeMatch {
		// Ignore responses that don't provide a service URL
		if strings.TrimSpace(match.XAddrs) == "" {
			continue
		}
		cameras = append(cameras, onvif.DiscoveredCamera{
			MessageID: env.Header.MessageID,
			Address:   match.EndpointReference.Address,
			Types:     match.Types,
			Scopes:    match.Scopes,
			XAddrs:    match.XAddrs,
		})
	}
	return cameras, nil
}