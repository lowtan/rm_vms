#include <string>
#include "Log.h"

struct AVDictionary;

// Standard AVDictionary setups for NVR recording efficiency
AVDictionary* configureAVDictionary(AVDictionary* options);
void logUnusedOptions(AVDictionary* dict, const std::string& tag = "Stream");