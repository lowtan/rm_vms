package discovery

import (
	"context"
	"fmt"
	"sync"
)

// SweepSubnet sends a unicast probe to all 254 IPs in a /24 subnet simultaneously.
// Example baseIP: "192.168.1."
func (v *Verifier) SweepSubnet(ctx context.Context, baseIP string) []VerifyResult {
	var wg sync.WaitGroup
	resultsChan := make(chan VerifyResult, 255)

	// Sweep IPs .1 through .254
	for i := 1; i < 255; i++ {
		targetIP := fmt.Sprintf("%s%d", baseIP, i)
		
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			
			// Use the UnicastProbe method we built earlier
			isONVIF, rawData := v.unicastProbe(ip)
			if isONVIF {
				resultsChan <- VerifyResult{
					IsValid:   true,
					Protocol:  "onvif",
					PortFound: 3702,
					RawData:   rawData, // Contains the XAddrs
				}
			}
		}(targetIP)
	}

	// Wait for all probes to finish or timeout
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var activeCameras []VerifyResult
	for result := range resultsChan {
		activeCameras = append(activeCameras, result)
	}

	return activeCameras
}