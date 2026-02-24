#include <iostream>
#include <string>
#include <vector>
#include <map>

#include "Log.h"
#include "VideoIngestion.h"

#include "Command.h"
#include "SharedMemory.h"

const std::shared_ptr<ISharedMemory> SHM = ISharedMemory::CreateInstance();


int main() {
    // Optimize I/O
    std::ios_base::sync_with_stdio(false);
    std::cin.tie(NULL);

    // 
    std::map<int, std::unique_ptr<VideoIngestion>> activeCameras;

    std::string line;
    while (std::getline(std::cin, line)) {
        if (line == "EXIT") break;

        Command cmd = parseCommand(line);

        if(cmd.Name != "") {

            if(cmd.Name == "WORKER") {
                // Initialize SharedMemory
                try {
                    std::string worker = cmd.Args.front();
                    std::string name = ringBufferNameFor(worker);
                    // Initial with some basic 
                    if(SHM->Create(name, 10, 1024000000000)==false){
                        Log::error("Failed to create RingBuffer for:" + name);
                        Log::send("{\"status\":\"shmerr\", \"worker\":" + name + "}");
                    }
                } catch (...) {
                    Log::error("Error initializing SharedMemory.");
                }
            }

            if(cmd.Name == "START") {
                try {

                    std::string idStr = cmd.Args.front();
                    cmd.Args.pop();
                    std::string url = cmd.Args.front();
                    cmd.Args.pop();

                    Log::info("id, url:" + idStr + " " + url );

                    // Respond to Go
                    Log::send("{\"status\":\"starting\", \"cam\":" + idStr + "}");

                    // Run logic
                    int camID = std::stoi(idStr);
                    activeCameras[camID] = std::make_unique<VideoIngestion>(SHM, camID, url);

                } catch (...) {
                    Log::error("Error starting video ingestion.");
                }
            }

        }

    }
    return 0;
}