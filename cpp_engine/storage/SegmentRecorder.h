#include <fcntl.h>
#include <unistd.h>
#include <sys/stat.h>
#include <vector>
#include <string>

extern "C" {
#include <libavformat/avformat.h>
}

class SegmentRecorder {

public:
    bool StartSegment(const std::string& filename, AVCodecParameters* inputCodecParams);
    void WritePacket(AVPacket* packet, AVRational inputTimeBase);
    void StopSegment();
};