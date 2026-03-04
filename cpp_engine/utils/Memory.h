#pragma once
#include <string>
#include <queue>

/**
 * Calculate the buffer needed for by stream VBR.
 * @param  r [VBR]
 * @return   [description]
 */
size_t calcBuffer(size_t r);