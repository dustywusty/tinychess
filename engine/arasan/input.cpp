// Your Move transport for Arasan's Input interface. No host stdin or threads.
#include "input.h"
#include <emscripten.h>
#include <cstdlib>

EM_JS(char*, nextCommand, (), {
 const value = Module.commands.shift();
 return value === undefined ? 0 : stringToNewUTF8(value);
});

Input::Input() : buf_index(0) {}

bool Input::checkInput(std::vector<std::string>& commands, std::mutex& mutex) {
 // Yield at Arasan's search checkpoints so stop can interrupt a live search.
 emscripten_sleep(0);
 char* command = nextCommand();
 if (!command) return false;
 std::lock_guard<std::mutex> lock(mutex);
 commands.emplace_back(command);
 free(command);
 return true;
}

bool Input::readInput(std::vector<std::string>& commands, std::mutex& mutex) {
 while (!checkInput(commands, mutex)) emscripten_sleep(10);
 return true;
}
