#pragma once
#include <string>
#include <queue>

// Struct to hold the parsed result
struct Command
{
    std::string Name;
    std::queue<std::string> Args;
};

Command parseCommand(std::string line);