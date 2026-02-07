#include <iostream>
#include <string>

#include "Log.h"
#include "VideoIngestion.h"

int main() {
    // Optimize I/O
    std::ios_base::sync_with_stdio(false);
    std::cin.tie(NULL);

    std::string line;
    while (std::getline(std::cin, line)) {
        if (line == "EXIT") break;

        // Basic parsing: "START <ID> <URL>"
        if (line.substr(0, 5) == "START") {
            try {
                size_t firstSpace = line.find(' ');
                size_t secondSpace = line.find(' ', firstSpace + 1);

                std::string idStr = line.substr(firstSpace + 1, secondSpace - firstSpace - 1);
                std::string url = line.substr(secondSpace + 1);

                // Respond to Go: "I received your command"
                // std::cout << "{\"status\":\"starting\", \"cam\":" << idStr << "}" << std::endl;
                Log::info("{\"status\":\"starting\", \"cam\":" + idStr + "}");

                // Run logic
                // startIngestion(std::stoi(idStr), url);
                VideoIngestion cam1(std::stoi(idStr), url);

            } catch (...) {
                Log::error("Error parsing command");
            }
        }
    }
    return 0;
}