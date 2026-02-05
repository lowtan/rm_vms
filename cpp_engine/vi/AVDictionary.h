extern "C" {
#include <libavformat/avformat.h>
}

// Standard AVDictionary setups for NVR recording efficiency
AVDictionary* configureAVDictionary(AVDictionary* options);