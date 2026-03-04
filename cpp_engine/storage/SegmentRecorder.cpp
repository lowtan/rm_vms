#include "SegmentRecorder.h"
#include "Log.h"

#include <fcntl.h>
#include <unistd.h>
#include <sys/stat.h>
#include <vector>
#include <string>


// Reserve 200MB (approx 10 mins of 4K H.265).
const int PREALLOCATE_SIZE = 200 * 1024 * 1024;

// The Write Callback
// FFmpeg calls this whenever it has data to flush to disk.
// 'opaque' will be our File Descriptor (int*).
static int write_packet_callback(void* opaque, uint8_t* buf, int buf_size) {
    int fd = *(int*)opaque;
    
    // Write directly to the pre-allocated Linux file descriptor
    ssize_t ret = write(fd, buf, buf_size);
    
    if (ret < 0) return -1; // IO Error
    return ret;
}

class SegmentRecorder {
private:
    AVFormatContext* outFormatCtx = nullptr;
    AVIOContext* avioCtx = nullptr;
    unsigned char* ioBuffer = nullptr; // Buffer for custom IO
    int fileDescriptor = -1;           // Linux File Descriptor
    bool isRecording = false;
    int videoIndex = -1;

    // Tuning: 64KB buffer for IO operations is usually a sweet spot
    const int IO_BUFFER_SIZE = 64 * 1024; 

public:
    bool StartSegment(const std::string& filename, AVCodecParameters* inputCodecParams) {

        // --- STEP 1: Linux File Creation & Allocation ---

        // A. Open File
        // O_CREAT: Create if missing
        // O_WRONLY: Write only
        // O_TRUNC: Overwrite if exists (reset size)
        // O_DIRECT: Optional (advanced), bypass OS cache. Only use if you know what you're doing.
        fileDescriptor = open(filename.c_str(), O_CREAT | O_WRONLY | O_TRUNC, 0644);
        if (fileDescriptor < 0) {
            Log::error("Failed to open file: " + filename);
            // std::cerr <<  << std::endl;
            return false;
        }

        // Pre-Allocate
        if (fallocate(fileDescriptor, 0, 0, PREALLOCATE_SIZE) != 0) {
            Log::error("Warning: fallocate failed. Disk might be full or FS doesn't support it.");
            // We proceed anyway, but performance might suffer.
        }


        // --- FFmpeg Context Setup ---

        // Allocate Output Context (Metadata container)
        avformat_alloc_output_context2(&outFormatCtx, nullptr, "mkv", filename.c_str());
        if (!outFormatCtx) return false;

        // Create the Custom IO Buffer
        ioBuffer = (unsigned char*)av_malloc(IO_BUFFER_SIZE);

        // Create the AVIOContext (The Bridge)
        // We pass:
        // 1. Buffer & Size
        // 2. Write Flag (1 = writeable)
        // 3. Opaque Pointer (&fileDescriptor) - passed to callback
        // 4. Read Callback (NULL)
        // 5. Write Callback (Our function above)
        // 6. Seek Callback (NULL - MKV is streamable, usually doesn't strictly need seek for writing)
        avioCtx = avio_alloc_context(ioBuffer, IO_BUFFER_SIZE, 1, 
                                     &fileDescriptor, nullptr, write_packet_callback, nullptr);

        // Assign the IO context to the Format Context
        outFormatCtx->pb = avioCtx;


        // --- Stream Setup (Same as before) ---

        AVStream* outStream = avformat_new_stream(outFormatCtx, nullptr);
        avcodec_parameters_copy(outStream->codecpar, inputCodecParams);
        outStream->codecpar->codec_tag = 0;
        videoIndex = outStream->index;

        // Write Header
        // FFmpeg now calls 'write_packet_callback' to write the header bytes
        if (avformat_write_header(outFormatCtx, nullptr) < 0) return false;

        isRecording = true;
        return true;
    }

    void WritePacket(AVPacket* packet, AVRational inputTimeBase) {
        if (!isRecording) return;

        // (Rescale logic same as previous example...)
        AVStream* outStream = outFormatCtx->streams[videoIndex];
        packet->pts = av_rescale_q_rnd(packet->pts, inputTimeBase, outStream->time_base, (AVRounding)(AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX));
        packet->dts = av_rescale_q_rnd(packet->dts, inputTimeBase, outStream->time_base, (AVRounding)(AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX));
        packet->duration = av_rescale_q(packet->duration, inputTimeBase, outStream->time_base);
        packet->stream_index = videoIndex;

        av_interleaved_write_frame(outFormatCtx, packet);
    }

    void StopSegment() {
        if (isRecording && outFormatCtx) {
            // Write Trailer (Index)
            av_write_trailer(outFormatCtx);

            // Clean up FFmpeg Internals
            // Note: We do NOT use avio_closep here because we own the FD.
            avformat_free_context(outFormatCtx);

            // Clean up Custom IO
            if (avioCtx) {
                av_free(avioCtx->buffer); // Important: Free the buffer we malloc'd
                av_free(avioCtx);
            }

            // Truncate
            // If we reserved 200MB but only used 150MB, cut off the excess slack.
            off_t actualSize = lseek(fileDescriptor, 0, SEEK_CUR);
            if (actualSize > 0) {
                ftruncate(fileDescriptor, actualSize);
            }

            // Close Linux File Descriptor
            if (fileDescriptor != -1) {
                close(fileDescriptor);
                fileDescriptor = -1;
            }
        }
        isRecording = false;
    }
};
