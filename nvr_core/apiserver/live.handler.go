package apiserver

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"nvr_core/stream"

	"github.com/asticode/go-astits"
)

func (api *APIServer) HandleLiveTransmuxTS(w http.ResponseWriter, r *http.Request) {

	// --- CRITICAL SECURITY OVERRIDE ---
	// The global API server has a strict 10-second WriteTimeout.
	// Since this is an endless stream, we use the ResponseController to 
	// disable the write deadline specifically for this individual request.
	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		log.Printf("[TS Handler] Warning: Failed to clear write deadline: %v", err)
	}
	// ----------------------------------

	camID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid camera ID", http.StatusBadRequest)
		return
	}

	worker := api.PM.CameraWorker(camID)
	if worker == nil {
		http.Error(w, "Camera not assigned to worker", http.StatusNotFound)
		return
	}

	hub := worker.StreamHubForCam(camID)
	if hub == nil {
		http.Error(w, "Stream not running", http.StatusNotFound)
		return
	}

	// Setup Endless HTTP Streaming Headers
	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Register as a Subscriber to the Hub
	sub := &stream.Subscriber{
		Send:               make(chan stream.StreamPacket, 256),
		WaitingForKeyframe: true, // Crucial: TS decoders need the SPS/PPS keyframe first
	}
	hub.Register <- sub

	// Ensure we unregister when the HTTP request drops
	defer func() {
		hub.Unregister <- sub
	}()

	// Initialize the pure-Go TS Muxer
	// We map the muxer's output directly to the HTTP ResponseWriter
	muxer := astits.NewMuxer(
		context.Background(),
		w,
	)

	ctx := r.Context()
	var pts int64 = 0 

	pmtWritten := false
	var audioCodec uint32 = 0

	// Pre-define PIDs
	const VideoPID uint16 = 256
	const AudioPID uint16 = 257

	for {
		select {
		case <-ctx.Done():
			log.Printf("[TS Handler] Client disconnected from Cam %d", camID)
			return

		case packet, ok := <-sub.Send:
			if !ok {
				return // Hub channel closed
			}

			// --- LATE BINDING & TABLE REPETITION ---
			if packet.MediaType == stream.MediaTypeAudio && !pmtWritten {
				// Cache audio codec but drop packet until Video Keyframe arrives
				audioCodec = packet.CodecID
				continue
			}

			if packet.MediaType == stream.MediaTypeVideo && packet.IsKeyFrame {
				// Bind the tables if this is the very first keyframe
				if !pmtWritten {
					muxer.AddElementaryStream(astits.PMTElementaryStream{
						ElementaryPID: VideoPID,
						StreamType:    astits.StreamType(stream.GetTSStreamType(packet.CodecID)),
					})
					muxer.SetPCRPID(VideoPID)

					if audioCodec != 0 {
						muxer.AddElementaryStream(astits.PMTElementaryStream{
							ElementaryPID: AudioPID,
							StreamType:    astits.StreamType(stream.GetTSStreamType(audioCodec)),
						})
					}
					pmtWritten = true
				}

				// CRITICAL FIX: Re-write the PAT/PMT tables before EVERY Keyframe.
				// This allows VLC to instantly recover if it drops packets or connects late.
				muxer.WriteTables()
			} else if !pmtWritten {
				// Drop all P-frames and Audio frames until the PMT is bound
				continue
			}

			// --- DEMUXING AND PACKETIZING ---
			var streamID uint8
			var targetPID uint16

			if packet.MediaType == stream.MediaTypeVideo {
				streamID = 224 // 0xE0
				targetPID = VideoPID
			} else if packet.MediaType == stream.MediaTypeAudio {
				streamID = 192 // 0xC0
				targetPID = AudioPID
			} else {
				continue 
			}

			// Package the raw frame into a PES
			pes := &astits.PESData{
				Header: &astits.PESHeader{
					OptionalHeader: &astits.PESOptionalHeader{
						// NOTE: Ensure your PTS clock logic is running here!
						PTS: &astits.ClockReference{Base: pts},
						DTS: &astits.ClockReference{Base: pts},
					},
					StreamID: streamID, 
				},
				Data: packet.Payload,
			}

			// Prepare the Muxer Data
			muxerData := &astits.MuxerData{
				PES: pes,
				PID: targetPID,
			}

			// Inject the Master Clock (PCR)
			// VLC will not render a single frame without a PCR to synchronize its timeline.
			if packet.MediaType == stream.MediaTypeVideo {
				muxerData.AdaptationField = &astits.PacketAdaptationField{
					HasPCR: true,
					PCR:    &astits.ClockReference{Base: pts},
				}
			}

			// Write the data to the HTTP stream
			if _, err := muxer.WriteData(muxerData); err != nil {
				log.Printf("[TS Handler] Muxer write error: %v", err)
				return
			}

			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

			if packet.MediaType == stream.MediaTypeVideo {
				pts += 3000 // Fake 30fps clock. 
			}
		}
	}

}