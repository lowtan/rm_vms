#include "SegmentRecorder.h"
#include <iostream>

#include "Log.h"

SegmentRecorder::~SegmentRecorder() {
    StopSegment();
}

bool SegmentRecorder::StartSegment(const std::string& filename, AVStream* inVideoStream, AVStream* inAudioStream) {
    if (isRecording) StopSegment();
    currentFilename = filename;

    Log::info("[SegmentRecorder] started for file \n " + currentFilename);

    avformat_alloc_output_context2(&outFormatCtx, nullptr, "mp4", filename.c_str());
    if (!outFormatCtx) return false;

    // --- Setup Video Stream ---
    if (inVideoStream) {
        AVStream* outVStream = avformat_new_stream(outFormatCtx, nullptr);
        avcodec_parameters_copy(outVStream->codecpar, inVideoStream->codecpar);
        outVStream->codecpar->codec_tag = 0;
        
        outVideoStreamIndex = outVStream->index;
        inVideoStreamIndex = inVideoStream->index;
        videoInputTimeBase = inVideoStream->time_base;
    }

    // --- Setup Audio Stream ---
    if (inAudioStream) {
        if (avformat_query_codec(outFormatCtx->oformat, inAudioStream->codecpar->codec_id, FF_COMPLIANCE_NORMAL) == 1) {

            AVStream* outAStream = avformat_new_stream(outFormatCtx, nullptr);
            avcodec_parameters_copy(outAStream->codecpar, inAudioStream->codecpar);
            outAStream->codecpar->codec_tag = 0;
            
            outAudioStreamIndex = outAStream->index;
            inAudioStreamIndex = inAudioStream->index;
            audioInputTimeBase = inAudioStream->time_base;

        } else {

            // The codec is illegal for MP4. Log it and gracefully ignore the audio stream.
            inAudioStreamIndex = -1;

            std::cerr << "[SegmentRecorder] Warning: Audio codec '" 
                      << avcodec_get_name(inAudioStream->codecpar->codec_id) 
                      << "' is not supported in MP4. Recording video only." << std::endl;

        }
    }

    // --- Open File & Write Header ---
    if (!(outFormatCtx->oformat->flags & AVFMT_NOFILE)) {
        if (avio_open(&outFormatCtx->pb, filename.c_str(), AVIO_FLAG_WRITE) < 0) {
            avformat_free_context(outFormatCtx);
            return false;
        }
    }

    AVDictionary* opts = nullptr;
    // av_dict_set(&opts, "movflags", "faststart", 0); // Optional: Web-optimized MP4
    if (avformat_write_header(outFormatCtx, &opts) < 0) {
        return false;
    }
    av_dict_free(&opts);

    isRecording = true;
    return true;
}

void SegmentRecorder::WritePacket(AVPacket* packet) {
    if (!isRecording || !outFormatCtx) return;

    AVStream* outStream = nullptr;
    AVRational inputTimeBase;

    // Route the packet and grab the correct timebase
    if (packet->stream_index == inVideoStreamIndex) {

        outStream = outFormatCtx->streams[outVideoStreamIndex];
        inputTimeBase = videoInputTimeBase;
        packet->stream_index = outVideoStreamIndex; // Re-map to MP4 index

    } else if (inAudioStreamIndex != -1 && packet->stream_index == inAudioStreamIndex) {

        outStream = outFormatCtx->streams[outAudioStreamIndex];
        inputTimeBase = audioInputTimeBase;
        packet->stream_index = outAudioStreamIndex; // Re-map to MP4 index

    } else {
        // Unknown stream (e.g., metadata). Safely ignore.
        return; 
    }

    // Rescale timestamps
    av_packet_rescale_ts(packet, inputTimeBase, outStream->time_base);

    // Interleave and write (FFmpeg handles the internal buffering to keep A/V in sync)
    if (av_interleaved_write_frame(outFormatCtx, packet) < 0) {
        Log::error("Error writing interleaved packet to file.");
    }
}

void SegmentRecorder::StopSegment() {
    if (isRecording && outFormatCtx) {
        av_write_trailer(outFormatCtx);
        if (!(outFormatCtx->oformat->flags & AVFMT_NOFILE)) {
            avio_closep(&outFormatCtx->pb);
        }
        avformat_free_context(outFormatCtx);
        outFormatCtx = nullptr;
        Log::info("[SegmentRecorder] stopped \n " + currentFilename);
    }
    isRecording = false;
}