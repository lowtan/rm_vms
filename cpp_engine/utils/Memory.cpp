#include "Memory.h"


/**
R: Camera Bitrate in bits per second (bps).
T: Buffer Duration in seconds. This is how much "time" the ring buffer holds. For local IPC between Go and C++, 1 to 2 seconds is incredibly safe.
M: Safety Margin multiplier. Usually 1.5 to 2.0. This absorbs sudden VBR spikes when the camera sees a lot of motion or generates a massive Keyframe.
 */
const size_t TM = 5;

size_t calcBuffer(size_t r) {
    return (r / 8) * TM;
}