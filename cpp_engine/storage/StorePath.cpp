#include "StorePath.h"
#include <string>

// Setting as constant here, should read from config.json
// in the future.
const std::string ROOT = "/recordings"

// === Path concating logic ===
// 
// ROOT + camID + year + month + date + time
// 

class StorePath {

public:
    std::string& For(int camID, AVPacket* packet) {

    }
};