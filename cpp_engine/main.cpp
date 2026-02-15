#include <iostream>
#include <string>
#include <map>

#include "Log.h"
#include "VideoIngestion.h"

#include "Command.h"

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

            Log::info("proper command:" + cmd.Name);
            Log::info("proper command args:" + std::to_string(cmd.Args.size()));

            // Basic parsing: "START <ID> <URL>"
            if (cmd.Name == "START") {
                try {

                    std::string idStr = cmd.Args.front();
                    cmd.Args.pop();
                    std::string url = cmd.Args.front();
                    cmd.Args.pop();

                    Log::info("id, url:" + idStr + " " + url );

                    // Respond to Go
                    Log::info("{\"status\":\"starting\", \"cam\":" + idStr + "}");

                    // Run logic
                    int camID = std::stoi(idStr);
                    activeCameras[camID] = std::make_unique<VideoIngestion>(camID, url);

                } catch (...) {
                    Log::error("Error parsing command");
                }
            }

        }

    }
    return 0;
}